package pkg

import (
	"github.com/turbonomic/data-ingestion-framework/pkg/conf"
	"github.com/turbonomic/data-ingestion-framework/pkg/discovery"
	"github.com/turbonomic/data-ingestion-framework/pkg/registration"

	"os"
	"os/signal"
	"syscall"

	"github.com/golang/glog"

	"github.com/turbonomic/turbo-go-sdk/pkg/probe"
	"github.com/turbonomic/turbo-go-sdk/pkg/service"
)

type disconnectFromTurboFunc func()

// The TurboDIF TAP Service.
// This service registers the TurboDIF probe to the Turbo server and perform periodic discovery to send
// entities created from the DIF metric endpoints
type DIFTAPService struct {
	tapService *service.TAPService
}

// Create new TAP service for the TurboDIF probe
func NewDIFTAPService(args *conf.DIFProbeArgs) (*DIFTAPService, error) {
	tapService, err := createTAPService(args)

	if err != nil {
		glog.Errorf("Error while building TurboDIF TAP service on target %v", err)
		return nil, err
	}

	return &DIFTAPService{tapService}, nil
}

// Start the TAP serrvice
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

	// Target
	var targetAddr, targetName string
	if difConf.TargetConf != nil {
		targetAddr = difConf.TargetConf.Address //HTTP URL for metric json data
		targetName = difConf.TargetConf.Name
	}

	// Load the supply chain config
	supplyChainConfig, err := conf.LoadSupplyChain(supplyChainConf)
	if err != nil {
		glog.Errorf("Error while parsing the supply chain config file %s: %++v", supplyChainConf, err)
		os.Exit(1)
	}
	// Registration client - configured with the supply chain definition
	registrationClient, err := registration.NewDIFRegistrationClient(supplyChainConfig, difConf.TargetTypeSuffix)

	if err != nil {
		glog.Fatalf("error: %v", err)
	}
	// Discovery client - target type, target address, supply chain
	targetType := registrationClient.TargetType()
	probeCategory := registrationClient.ProbeCategory()

	var optionalTargetAddr *string
	if len(targetAddr) > 0 {
		optionalTargetAddr = &targetAddr
	}
	discoveryTargetParams := &discovery.DiscoveryTargetParams{
		TargetType:            targetType,
		ProbeCategory:         probeCategory,
		TargetName:            targetName,
		OptionalTargetAddress: optionalTargetAddr,
	}
	keepStandalone := args.KeepStandalone

	discoveryClient := discovery.NewDiscoveryClient(discoveryTargetParams, *keepStandalone, supplyChainConfig)

	builder := probe.NewProbeBuilder(targetType, *supplyChainConfig.ProbeCategory).
		WithDiscoveryOptions(probe.FullRediscoveryIntervalSecondsOption(int32(*args.DiscoveryIntervalSec))).
		WithEntityMetadata(registrationClient).
		RegisteredBy(registrationClient)

	if len(targetAddr) > 0 {
		// Preconfigured with target address or DIF metric endpoint
		glog.Infof("***** Should discover target %s", targetAddr)
		builder = builder.DiscoversTarget(targetAddr, discoveryClient)
	} else {
		// Target will be entered from the UI
		glog.Infof("Not discovering target")
		builder = builder.WithDiscoveryClient(discoveryClient)
	}

	return service.NewTAPServiceBuilder().
		WithTurboCommunicator(communicator).
		WithTurboProbe(builder).
		Create()
}

// handleExit disconnects the tap service from Turbo service when DIF probe is terminated
func handleExit(disconnectFunc disconnectFromTurboFunc) {
	glog.V(4).Infof("*** Handling TurboDIF Termination ***")
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan,
		os.Interrupt,
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
