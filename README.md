# Data Ingestion Framework: Open Framework for Turbonomic Platform

## Overview 
The Data Ingestion Framework (DIF) is a framework that allows Turbonomic to collect external metrics from customer and leverages [Turbonomic](https://www.ibm.com/products/turbonomic)'s patented analysis engine to provide visibility and control across the entire application stack in order to assure the performance, efficiency and compliance in real time.

## Documentation
Read the official documentation [HERE](https://www.ibm.com/docs/en/tarm/latest?topic=documentation-integration-data-ingestion-framework).

## How DIF works
The custom entities and entity metrics are declared in a predefined JSON format. The DIF probe takes the JSON input and convert them into the data structures known by Turbonomic, and push them to the server. The DIF probe is a generic Turbonomic SDK probe that performs all the usual supply chain and entity validations, as well as participating in the broader mediation framework to ensure all conflicts are resolved, and consistency is maintained.
![image](https://user-images.githubusercontent.com/10012486/88306380-a6b36b80-ccd8-11ea-9236-063577d60430.png)
