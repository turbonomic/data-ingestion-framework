package pkg

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.ibm.com/turbonomic/data-ingestion-framework/pkg/conf"
	"github.ibm.com/turbonomic/data-ingestion-framework/pkg/discovery"
	"github.ibm.com/turbonomic/data-ingestion-framework/pkg/registration"
	"github.ibm.com/turbonomic/data-ingestion-framework/version"

	"github.com/golang/glog"

	"github.ibm.com/turbonomic/turbo-go-sdk/pkg/probe"
	"github.ibm.com/turbonomic/turbo-go-sdk/pkg/service"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

const (
	k8sDefaultNamespace   = "default"
	kubernetesServiceName = "kubernetes"
)

type disconnectFromTurboFunc func()

// DIFTAPService defines TurboDIF TAP Service.
// This service registers the TurboDIF probe to the Turbo server and perform periodic discovery to send
// entities created from the DIF metric endpoints
type DIFTAPService struct {
	tapService *service.TAPService
}

// NewDIFTAPService creates a new TAP service for the TurboDIF probe
func NewDIFTAPService(args *conf.DIFProbeArgs) (*DIFTAPService, error) {
	tapService, err := createTAPService(args)

	if err != nil {
		glog.Errorf("Error while building TurboDIF TAP service on target %v", err)
		return nil, err
	}

	return &DIFTAPService{tapService}, nil
}

// Start the TAP service
func (p *DIFTAPService) Start() {
	glog.V(0).Infof("Starting TurboDIF TAP service...")

	// Disconnect from Turbo server when DIF probe is shutdown
	handleExit(func() { p.tapService.DisconnectFromTurbo() })

	// Connect to the Turbo server
	p.tapService.ConnectToTurbo()
}

func createTAPService(args *conf.DIFProbeArgs) (*service.TAPService, error) {
	// Read and parse the turbo server config and supply chain config
	confPath := conf.DefaultConfPath
	supplyChainConf := conf.DefaultSupplyChainConfPath

	if os.Getenv("TURBODIF_LOCAL_DEBUG") == "1" {
		confPath = conf.LocalDebugConfPath
		supplyChainConf = conf.LocalDebugSupplyChainConfPath
		glog.V(2).Infof("Using config files %s, %s for local debugging",
			confPath, supplyChainConf)
	}

	// Load the Turbo server and target config
	difConf, err := conf.NewDIFConf(confPath)
	if err != nil {
		glog.Errorf("Error while parsing the service config file %s: %v", confPath, err)
		os.Exit(1)
	}

	glog.V(3).Infof("Read service configuration from %s: %++v", confPath, difConf)

	// Server communicator
	communicator := difConf.Communicator

	// Target - this spec is optional since the target can be added from the UI
	// But if specified, we require both the target name and address to be present
	var targetAddr, targetName string
	if difConf.TargetConf != nil {
		targetAddr = difConf.TargetConf.Address // HTTP URL for metric server endpoint serving the json data
		targetName = difConf.TargetConf.Name    // User friendly name for metric server endpoint
	}

	// Load the supply chain config and create the registration client
	supplyChainConfig, err := conf.LoadSupplyChain(supplyChainConf)
	if err != nil {
		glog.Errorf("Error while parsing the supply chain config file %s: %++v", supplyChainConf, err)
		os.Exit(1)
	}
	supplyChain, err := registration.NewSupplyChain(supplyChainConfig)
	if err != nil {
		glog.Errorf("Error while parsing the supply chain: %+v", err)
		os.Exit(1)
	}
	supplyChain.IgnoreIfPresent(*args.IgnoreCommodityIfPresent)

	// Registration client - configured with the supply chain definition
	registrationClient := registration.NewRegistrationClient(supplyChain)

	// Discovery client - target type, target address, supply chain
	targetType := registrationClient.TargetType()
	probeCategory := registrationClient.ProbeCategory()
	probeUICategory := registrationClient.ProbeUICategory()

	var optionalTargetAddr *string
	if len(targetAddr) > 0 {
		optionalTargetAddr = &targetAddr
	}
	var optionalTargetName *string
	if len(targetName) > 0 {
		optionalTargetName = &targetName
	}

	targetParams := &discovery.TargetParams{
		TargetType:            targetType,
		ProbeCategory:         probeCategory,
		OptionalTargetName:    optionalTargetName,
		OptionalTargetAddress: optionalTargetAddr,
	}
	keepStandalone := args.KeepStandalone
	bindingChannel := getBindingChannel()
	discoveryClient, targetProvider := discovery.NewDiscoveryClient(targetParams, *keepStandalone, supplyChain, bindingChannel)

	probeVersion := version.Version
	probeDisplayName := getProbeDisplayName(targetType)

	// Turbo probe
	builder := probe.NewProbeBuilder(targetType, probeCategory, probeUICategory).
		WithVersion(probeVersion).
		WithDisplayName(probeDisplayName).
		WithDiscoveryOptions(probe.FullRediscoveryIntervalSecondsOption(int32(*args.DiscoveryIntervalSec))).
		WithEntityMetadata(registrationClient).
		RegisteredBy(registrationClient).
		WithActionPolicies(registrationClient).
		WithActionMergePolicies(registrationClient)

	if len(targetAddr) > 0 && len(targetName) > 0 {
		// Preconfigured with target address or DIF metric endpoint
		glog.Infof("***** Should discover target %s with URL %s ", targetName, targetAddr)
		// Target will be added as part of probe registration only when communicating with the server
		// on a secure websocket, else secret containing Turbo server admin user must be configured
		// to auto-add the target using API
		// Kubernetes Probe Registration Client
		builder.WithSecureTargetProvider(targetProvider)
		builder = builder.DiscoversTarget(targetName, discoveryClient)
	} else {
		// Target will be entered from the UI
		glog.Infof("Not discovering target")
		builder = builder.WithDiscoveryClient(discoveryClient)
	}

	return service.NewTAPServiceBuilder().
		WithCommunicationBindingChannel(bindingChannel).
		WithTurboCommunicator(communicator).
		WithTurboProbe(builder).
		Create()
}

// getProbeDisplayName constructs a display name for the probe based on the input probe type
func getProbeDisplayName(probeType string) string {
	return strings.Join([]string{probeType, "Probe"}, " ")
}

// getBindingChannel tries to fetch the in-cluster k8s service id as the binding channel; if that's not successful, an
// empty string will be returned as the binding channel
func getBindingChannel() string {
	kubeConfig, err := restclient.InClusterConfig()
	if err != nil {
		glog.V(2).Infof("Setting an empty binding channel being unable to acquire the in-cluster kube config: %++v", err)
		return ""
	}
	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		glog.V(2).Infof("Setting an empty binding channel being unable to acquire the kube client: %++v", err)
		return ""
	}
	kubeSvc, err := kubeClient.CoreV1().Services(k8sDefaultNamespace).Get(context.TODO(), kubernetesServiceName, metav1.GetOptions{})
	if err != nil {
		glog.V(2).Infof("Setting an empty binding channel being unable to retrieve the k8s service id: %++v", err)
		return ""
	}
	return string(kubeSvc.UID)
}

// handleExit disconnects the tap service from Turbo service when DIF probe is terminated
func handleExit(disconnectFunc disconnectFromTurboFunc) {
	glog.V(4).Infof("*** Handling TurboDIF Termination ***")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGHUP)

	go func() {
		select {
		case sig := <-sigChan:
			// Close the mediation container including the endpoints. It avoids the
			// invalid endpoints remaining in the server side. See OM-28801.
			glog.V(2).Infof("Signal %s received. Disconnecting TurboDIF from Turbo server...\n", sig)
			disconnectFunc()
		}
	}()
}
