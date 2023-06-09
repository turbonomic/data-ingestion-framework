# Data Ingestion Framework: Open Framework for Turbonomic Platform

## Overview 
The Data Ingestion Framework (DIF) is a framework that allows Turbonomic to collect external metrics from customer and leverages [Turbonomic](https://turbonomic.com/)'s patented analysis engine to provide visibility and control across the entire application stack in order to assure the performance, efficiency and compliance in real time.

## How DIF works
The custom entities and entity metrics are declared in a predefined JSON format. The DIF probe takes the JSON input and convert them into the data structures known by Turbonomic, and push them to the server. The DIF probe is a generic Turbonomic SDK probe that performs all the usual supply chain and entity validations, as well as participating in the broader mediation framework to ensure all conflicts are resolved, and consistency is maintained.
![image](https://user-images.githubusercontent.com/10012486/88306380-a6b36b80-ccd8-11ea-9236-063577d60430.png)

## DIF JSON Examples
For detailed documentation on the JSON schema, see [here](https://www.ibm.com/docs/en/tarm/8.9.1?topic=framework-dif-schema-files#Schema_Files__dif-topology-schema).

#### A proxy Virtual Machine entity with VMEM metric
```json
{
  "version": "v1",
  "updateTime": 1595519486,
  "scope": "",
  "source": "",
  "topology": [
    {
      "uniqueId": "spcfq9keqj-worker-1",
      "type": "virtualMachine",
      "name": "spcfq9keqj-worker-1",
      "hostedOn": null,
      "matchIdentifiers": {
        "ipAddress": "172.23.0.5"
      },
      "partOf": [
        {
          "uniqueId": "DatabaseServer-10.10.169.38-turbonomic",
          "entity": "databaseServer"
        }
      ],
      "metrics": {
        "memory": [
          {
            "average": 1363148.8,
            "capacity": 3670016
          }
        ]
      }
    }
  ]
}

```
#### A Database Server entity hosted on a Virtual Machine
```json
{
  "version": "v1",
  "updateTime": 1595519551,
  "scope": "Prometheus",
  "source": "",
  "topology": [
    {
      "uniqueId": "DatabaseServer-10.10.169.38-turbonomic",
      "type": "databaseServer",
      "name": "DatabaseServer-10.10.169.38-turbonomic",
      "hostedOn": {
        "hostType": [
          "virtualMachine"
        ],
        "ipAddress": "10.10.169.38"
      },
      "metrics": {
        "connection": [
          {
            "average": 14,
            "capacity": 151
          }
        ],
        "dbCacheHitRate": [
          {
            "average": 100
          }
        ],
        "dbMem": [
          {
            "average": 16636512,
            "capacity": 16777216
          }
        ],
        "memory": [
          {
            "average": 16636512,
            "capacity": 16777216
          }
        ],
        "transaction": [
          {
            "average": 0.5388918827326818
          }
        ]
      }
    }
  ]
}  
```
## How to enable a metric server and add it as a DIF target
The most typical use case to enable DIF probe and add a metric server target is to bring in additional metrics for existing entities discovered by other probes. With the latest Turbonomic 7.22.4 appliance, a DIF probe is deployed by default (which is used by the `prometurbo` metric server). You can reuse this DIF probe to bring in additional metrics from other metric sources. Alternatively, you can deploy a standalone DIF probe by following the instructions [here](https://github.com/turbonomic/data-ingestion-framework/tree/master/deploy).

Follow the steps below to enable a metric server:
* Determine the type of the entity and the metrics that are available for the entity
* If the entity is a proxy entity, determine the matching identity of the entity, such that it can be stitched with the real entity in the platform
* Write a program to construct the JSON data, and expose it via a simple HTTP service
* Deploy the HTTP service, which serves as the metric server, and make sure it is accessible by the DIF probe
* Add this metric server as a DIF target in the Turbonomic UI under the Custom Probe category
![image](https://user-images.githubusercontent.com/10012486/88309708-b2089600-ccdc-11ea-8936-e150a27a0aef.png)
![image](https://user-images.githubusercontent.com/10012486/88310681-f3e60c00-ccdd-11ea-8a5b-6d68b26f216b.png)
* Wait for the next discovery and broadcast to complete, and confirm the metrics are available on the entities
