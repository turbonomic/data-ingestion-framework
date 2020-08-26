# Azure Log Analytics
This example shows how to query memory metrics of virtual machines that are connected to Azure log analytics workspace, and expose these metrics through a simple web service to be consumed by [Turbonomic](https://turbonomic.com/)'s [data-ingestion-framework](https://github.com/turbonomic/data-ingestion-framework) probes.

## Getting started
### Prerequisites
To get the memory usage of virtual machines running in Azure, these virtual machines must be connected to the corresponding Azure log analytics workspaces. In addition, you must make sure the following performance counters are enabled for these workspaces. They are **NOT** enabled by default so you need manually add them:

* For Linux machines: **Used Memory MBytes**
* For Windows machines: **Memory(*)\Committed Bytes** 

To do so, navigate to your workspace, click **Advanced Settings**, **Data**. For example:
![image](https://user-images.githubusercontent.com/10012486/89071500-e1567d00-d344-11ea-9660-ffd9290c021e.png)

Make sure your Azure Active Directory App has required permission to access the Log Analytics API. For more details, refer to [Azure Log Analytics Search API](https://dev.loganalytics.io/documentation/1-Tutorials/Direct-API).

In addition, this guide assumes you already have a `turbodif` probe running in your Turbonomic environment. If you deploy Turbonomic in an appliance, the Target Turbo with Turbo is automatically enabled, and a default `turbodif` probe is already enabled:

![image](https://user-images.githubusercontent.com/10012486/91324907-963b5880-e790-11ea-978a-e48ebecc2752.png)

If your Turbonomic environment does not have a `turbodif` probe running, you must deploy one following the instructions [here](https://github.com/turbonomic/data-ingestion-framework/tree/master/deploy).

### Collect the following required information
* The following information is required to authenticate with Azure service, and query the specific workspaces that are connected to your virtual machines:

  * `AZURE_TENANT_ID`
  * `AZURE_CLIENT_ID`
  * `AZURE_CLIENT_SECRET`

* The IDs of the log analytics workspaces that are connected to the virtual machines for which you want to get the memory metrics. 

### Deploy the metric server
It is recommend to deploy the metric server in the same cluster where your `turbodif` probe is running. The [deploy](https://github.com/turbonomic/data-ingestion-framework/tree/master/example/azure-loganalytics/deploy) subddirectory provides a sample yaml to create a deployment and a service:

#### Create an `azure` kubernetes secret object that contains the azure account information:
* Create an `azure-target` file with the required azure target account information:
```
tenant: <AZURE_TENANT_ID>
client: <AZURE_CLIENT_ID>
key: <AZURE_CLIENT_SECRET>
```

* Create an `azure` kubernetes secret object from the above file:
```
$ kubectl create secret generic azure --from-file=./azure-target
secret/azure created
```
The `azure-target` is the target ID:
```
$ kubectl get secret azure -o yaml
apiVersion: v1
data:
  azure-target: Y2xpZW50OiBkMTk5MDY1Yi1mZmMxLTQyN2YtOWZkMi0zNWRlOGI3YTBiZmQKa2V5OiAuUGEyMz06PVhVT1kwWEgxeUVAN04udG1FRV9HZC1KQQp0ZW5hbnQ6IDhlNGYwNzEzLTVlZWEtNGRhMC05OWMwLWY3ZTQxNzk0YmU0YQo=
kind: Secret

```
If there is already an `azure` kubernetes secret object created for the `mediation-azure` probe in your cluster, it can be used directly, and the above two steps can be ignored. You must identify the target ID from the `data` field of the secret. 

#### Create the deployment and service
* Update the [deploy.yaml](https://github.com/turbonomic/data-ingestion-framework/tree/master/example/azure-loganalytics/deploy/deploy.yaml), and replace the following fields:
  * `<WORKSPACE_IDS_SEPARATED_BY_COMMA>`
  * If you are using an existing `azure` secret object, replace the `azure-target` value with the appropriate target ID

* Create the deployment
```
$ kubectl create -f deployment.yaml
```

* Make sure the pod and service are started:
```
$ kubectl get po | grep loganalytics
azure-loganalytics-8557cddbf5-5qc5s              1/1     Running   0          72m

$ kubectl get svc | grep loganalytics
azure-loganalytics          ClusterIP      10.233.52.141   <none>          8081/TCP       96m
```
If your metric server and `turbodif` are deployed in the same cluster, the metric endpoint can be accessed by `turbodif` at `http://azure-loganalytics:8081/metrics`.

If your metric server and `turbodif` are deployed in different clusters, you **must** make sure that the metric endpoint can be accessed from `turbodif`. One simple way to do so is to expose the `loganalytics` service as a `LoadBalancer`.

### Add the metric server as a target in Turbonomic UI
![image](https://user-images.githubusercontent.com/10012486/89074115-c3d7e200-d349-11ea-9043-08d02cd1a5e7.png)

## Source Code
For your convenience, this example provides reference implementations in two languages: [Golang](https://github.com/turbonomic/data-ingestion-framework/tree/master/example/azure-loganalytics/golang) (1.14+) and [Python](https://github.com/turbonomic/data-ingestion-framework/tree/master/example/azure-loganalytics/python) (3.5.3+). Follow the instructions in the individual subdirectory if would like to customize the code, and build your own docker images.
