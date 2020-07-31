# Azure Log Analytics
This example shows how to query memory metrics of virtual machines that are connected to Azure log analytics workspace, and expose these metrics through a simple web service to be consumed by [Turbonomic](https://turbonomic.com/)'s [data-ingestion-framework](https://github.com/turbonomic/data-ingestion-framework) probes.

## Getting started
### Prerequisites
To get the memory usage of virtual machines running in Azure, these virtual machines must be connected to the corresponding Azure log analytics workspaces. In addition, you must make sure the following performance counters are enabled for these workspaces:

* For Linux machines: **Used Memory MBytes**
* For Windows machines: **Memory(*)\Committed Bytes** 

To do so, navigate to your workspace, click **Advanced Settings**, **Data**. For example:
![image](https://user-images.githubusercontent.com/10012486/89071500-e1567d00-d344-11ea-9660-ffd9290c021e.png)

Make sure your Azure Active Directory App has required permission to access the Log Analytics API. For more details, refer to [Azure Log Analytics Search API](https://dev.loganalytics.io/documentation/1-Tutorials/Direct-API).

### Collect the following required information
The following information is required to authenticate with Azure service, and query the specific workspaces that are connected to your virtual machines. You must set them as environment variables to be picked up by the metric server:

* `AZURE_TENANT_ID`
* `AZURE_CLIENT_ID`
* `AZURE_CLIENT_SECRET`
* `AZURE_LOG_ANALYTICS_WORKSPACES`: 
This is a list of workspace IDs, separated by comma

### Build and deploy the metric server
For your convenience, this example provides reference implementations in two languages: [Golang](https://github.com/turbonomic/data-ingestion-framework/tree/azure-loganalytics/example/azure-loganalytics/golang) (1.14+) and [Python](https://github.com/turbonomic/data-ingestion-framework/tree/azure-loganalytics/example/azure-loganalytics/python) (3.5.3+). Follow the instructions in the individual subdirectory to build and generate docker images.

It is recommend to deploy the metric server in the same cluster where your `turbodif` probe is running. The [deploy](https://github.com/turbonomic/data-ingestion-framework/tree/azure-loganalytics/example/azure-loganalytics/deploy) subddirectory provides a sample yaml to create a deployment and a service:

* Create the `clientid` and `clientsecret` as kubernetes secret object:

```
$ kubectl create secret generic azure-loganalytics-clientid --from-file=./clientid
secret/azure-loganalytics-clientid created

$ kubectl create secret generic azure-loganalytics-clientsecret --from-file=./clientsecret 
secret/azure-loganalytics-clientsecret created
```
* Update the [deployment.yaml](https://github.com/turbonomic/data-ingestion-framework/tree/azure-loganalytics/example/azure-loganalytics/deploy/deployment.yaml), and replace the `<IMAGE>`, `<WORKSPACE_IDS_SEPARATED_BY_COMMA>` and `<TENANT_ID>`, and create the deployment

* Make sure the pod and service are started:
```
$ kubectl get po | grep loganalytics
azure-loganalytics-8557cddbf5-5qc5s              1/1     Running   0          72m

$ kubectl get svc | grep loganalytics
azure-loganalytics          ClusterIP      10.233.52.141   <none>          8081/TCP       96m
```

### Add the metric server as a target in Turbonomic UI
![image](https://user-images.githubusercontent.com/10012486/89074115-c3d7e200-d349-11ea-9043-08d02cd1a5e7.png)

