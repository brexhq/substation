# cmd

Contains applications (apps) used in Substation deployments. Apps are organized by either the infrastructure they are deployed to (e.g., AWS) or the source of the data (e.g., file, http).

Any app that implements the ingest, transform, load (ITL) functionality of Substation is named `substation` and shares the same configuration file format (see [build/config/](/build/config/) for more information).

## app.go

Contains the core Substation application code. This code can be used to create new Substation applications.

### design

Substation operates using a system of goroutines and channels:
* data ingest, transform, and load are handled by unique goroutines
* data streams between goroutines using a [pipeline pattern](https://go.dev/blog/pipelines)
* errors in any goroutine interrupt the application

This execution model was chosen for its ability to support horizontal scaling, high-latency data processing, and efficient delivery of data.

## aws/lambda

Contains Substation apps deployed as AWS Lambda functions. More information is available in 
[cmd/aws/lambda/README.md](/cmd/aws/lambda/README.md).

## file/substation

Reads and processes data stored in a local file. The app can be deployed anywhere, including non-container infrastructure, and is recommended for local testing and development.
