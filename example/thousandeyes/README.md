# ThousandEyes test example
Goal: Monitor and trend ThousandEyes Service Latency in Turbonomic

Pre-requisite: ThousandEyes user name and authentication token

ThousandEyes service latency for various application or SaaS endpoints:
https://api.thousandeyes.com/v6/tests.json

In this example, we will query ThousandEyes to list of tests for a project, and push each service with its respective response time
In addition, this guide assumes you already have a `turbodif` probe running in your Turbonomic environment. If you deploy Turbonomic in an appliance, the Target Turbo with Turbo is automatically enabled, and a default `turbodif` probe is already enabled:

![image](https://user-images.githubusercontent.com/10012486/91324907-963b5880-e790-11ea-978a-e48ebecc2752.png)

If your Turbonomic environment does not have a `turbodif` probe running, you must deploy one following the instructions [here](https://github.com/turbonomic/data-ingestion-framework/tree/master/deploy).

### Collect the following required information
* The following information is required to authenticate with the ThousandEyes service:

  * `username`
  * `token`

### Deploy the metric server
The [deploy](https://github.com/turbonomic/data-ingestion-framework/tree/master/example/thousandeyes/deploy) subddirectory provides a sample yaml to create a deployment and a service:

#### Create an `thousandeyes` kubernetes secret object that contains the thousandeyes account information:
* Create an `credentials` file with the required thousandeyes credentials information:
```
username: <thousandeyes_email>
token: <thousandeyes_token>
```

* Create an `thousandeyes` kubernetes secret object from the above file:
```
$ kubectl create secret generic thousandeyes --from-file=./credentials
secret/thousandeyes created
```
The `thousandeyes` is the target ID:
```
$ kubectl get secret thousandeyes -o yaml
apiVersion: v1
data:
  credentials: XYZ
kind: Secret

```

#### Create the deployment and service
* Update the [deploy.yaml](https://github.com/turbonomic/data-ingestion-framework/tree/master/example/thousandeyes/deploy/deploy.yaml), and replace the following fields:

* Create the deployment
```
$ kubectl create -f deployment.yaml
```

* Make sure the pod and service are started:
```
$ kubectl get pods | grep thousandeyes
thousandeyes-7fcc78547-gq68d                1/1     Running            0          10m

$ kubectl get svc | grep thousandeyes
thousandeyes                                             ClusterIP      10.233.46.226   <none>         8081/TCP                                                     38m
```
If your metric server and `turbodif` are deployed in the same cluster, the metric endpoint can be accessed by `turbodif` at `http://thousandeyes:8081/metrics`.

If your metric server and `turbodif` are deployed in different clusters, you **must** make sure that the metric endpoint can be accessed from `turbodif`. One simple way to do so is to expose the `thousandeyes` service as a `LoadBalancer`.

### Add the metric server as a target in Turbonomic UI
![image](https://user-images.githubusercontent.com/10012486/89074115-c3d7e200-d349-11ea-9043-08d02cd1a5e7.png)

## Source Code
For your convenience, this example provides reference implementations in [Golang](https://github.com/turbonomic/data-ingestion-framework/tree/master/example/thousandeyes/golang) (1.14+). Follow the instructions in the individual subdirectory if would like to customize the code, and build your own docker images.
