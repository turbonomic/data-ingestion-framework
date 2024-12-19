## CSV to Data Ingestion Framework (DIF)

### Description
This script takes a CSV and reformats it into the required JSON format for the Turbonomic DIF target type. It is intended as a proof-of-concept that can quickly be deployed to show *how* DIF can be used, but it is **not** intended to run as a production-level implementation of DIF.

This script takes a CSV as input and provides a target web address for the DIF target type in a Turbonomic instance. The JSON will be served from within the Turbonomic deployment, accessible on port 8081. It runs as a Deployment and is required to be in the Turbonomic Kubernetes cluster.

There are two main ways to deliver the CSV data to the script:

* Cloud-based storage (Azure Blob or AWS S3 are supported)
* Local FTP container  

The setup and files required will be different depending on whether you choose Cloud or FTP, so please read through the **Script Setup** carefully. The script will run continuously once deployed, executed as GET requests are made.

All credentials for accessing Cloud providers are presented through Kubernetes secrets, detailed below in the **Secret Details** section.  

Lastly, The script also requires a config file, implemented as a Kubernetes configMap, that includes mappings of the columns provided in the input CSV to a common name that the script can use. Additional configs are detailed below in the **ConfigMap Details** section.

Examples of each of the necessary Kubernetes YAML files are included in the /src/kubernetes folder of this repository, so please take a look at those if anything is unclear.  

### Script Setup    
1. Upload required containers  
    * Script container: *turbointegrations/csv-to-dif*
    * If using the FTP method, you will also need the ftp container: *turbointegrations/turbo-ftp* 
    * The container images are hosted on DockerHub:
        * [turbointegrations/csv-to-dif](https://hub.docker.com/r/turbointegrations/csv-to-dif) or using the command `docker pull turbointegrations/csv-to-dif`
        * [turbointegrations/turbo-ftp](https://hub.docker.com/r/turbointegrations/turbo-ftp) or using the command `docker pull turbointegrations/turbo-ftp`
2. Complete and upload secrets yaml (see **Secret Details** below for further details)
    * `kubectl apply -f csv-to-dif-secrets.yml`
3. Complete and upload the ConfigMap with the appropriate fields (see **ConfigMap Details** below for further details)
    * `kubectl apply -f csv-to-dif-configmap.yml`
4. Upload and apply Job yaml definition
    * FTP: `kubectl apply -f csv-to-dif-ftp.yml`
    * Cloud: `kubectl apply -f csv-to-dif-cloud.yml`
5. Upload CSV to either FTP or Cloud destination
    * FTP: The FTP can be accessed via port 31234 on the Turbonomic instance, with passive connection on ports 30020 and 30021
    * Cloud: The Turbonomic instance must have connection to the outside internet
6. Add the DIF target in Turbonomic
    * Provide whatever name is desired or appropriate for identifying the DIF target
    * For the URL field, use this: http://csv-to-dif-target.turbointegrations:8081/dif_metrics

### ConfigMap Details  
* CSV_LOCATION - Location of CSV. One of *AWS*, *AZURE* or *FTP*
* INPUT_CSV_NAME - CSV file name
* ENTITY_FIELD_MAP - Mapping for supported entity fields to columns defined in CSV
    - The format is a {Key: Value} mapping, where the *key* is the required name below and the *value* is the column name from the input CSV.
    - Required Fields: 
        - app_name - Business Application name
        - entity_name - Name of entity
        - entity_ip - IP address of entity (required for VirtualMachines only)
        - entity_type - Entity type (Must be a supported type in Turbonomic Supply Chain)
        - parent_name - Name of parent entity 
            - 'Parent' refers to the Buyer entity in the Turbo supply chain, e.g. the parent of an Application would be a Service. The script will strictly match based on supported parent types in Turbonomic, with the exception of VirtualMachine -> BusinessApplication, which is supported for convenience.
        - parent_type - Type of parent entity
* METRIC_FIELD_MAP - Mapping for supported metric fields to columns defined in CSV
    - The format is a {Key: Value} mapping, where the *key* is the required name below and the *value* is the column name from the input CSV.
    - Required field format:
        - {metric_type}\_{average|capacity|peak}
            - For example: *memory_average* or *cacheHitRate_capacity*
            - Supported metric types for each entity can be found [here](http://docs.turbonomic.com/docApp/doc/indexDIF.html?config=DIF#!/DIF/_DIF_Topics/CommoditiesBoughtAndSoldByEnts.xml)
            - The metric_type must match exactly with the name as defined in the documentation.
* APP_PREFIX - Optional Prefix for Business App names
* LOG_DIR - Optional log directory for persistent log files. Default is container STDOUT
* LOG_FILE - Optional log file for persistent log files. Default is container STDOUT
* LOG_LEVEL - Optional flag for setting logging level. One of *DEBUG*, *INFO*, *WARNING*, *ERROR*. Default is INFO

### Secret Details 
Secret must be named **dif-csv-location-info**  

Depending on CSV_LOCATION, you will need to add the following fields to the secret 

##### Azure Blob:  
1. AZURE_CONNECTION_STRING - The Azure Blob connection string  
    To find the Azure connection string:
    1. Sign in to the Azure portal.
    2. Locate your storage account.
    3. In the Settings section of the storage account overview, select Access keys. Here, you can view your account access keys and the complete connection string for each key.
    4. Find the Connection string value under key1, and select the Copy button to copy the connection string.
2. AZURE_CONTAINER_NAME - The Azure Blob Container name

##### AWS S3 Bucket:
1. AWS_ACCESS_KEY_ID - Account Access Key ID
2. AWS_SECRET_ACCESS_KEY - Account Secret Access Key
3. AWS_REGION_NAME - Region name where S3 bucket is located
4. AWS_BUCKET_NAME - S3 Bucket name

##### Local FTP Server:
1. TURBO_ADDRESS - Turbonomic instance IP

### Logs
By default, if no persistent logs are defined in the ConfigMap input, the script logs can be accessed by connecting directly to the container output:  
    `kubectl logs <csv-to-dif-container> -n turbointegrations -c csv-to-dif`