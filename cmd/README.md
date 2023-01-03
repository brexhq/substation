# cmd

Contains applications (apps) used in Substation deployments. Apps are organized by either the infrastructure they are deployed to (e.g., AWS) or the source of the data (e.g., file, http).

Any app that implements the ingest, transform, load (ITL) functionality of Substation is named `substation` and shares the same configuration file format (see [build/config/](/build/config/) for more information).

## app.go

Contains the core Substation application code. This code can be used to create new Substation applications.

## aws/lambda

Contains Substation apps deployed as AWS Lambda functions. More information is available in 
[cmd/aws/lambda/README.md](/cmd/aws/lambda/README.md).

## file/substation

Reads and processes data from local disk, HTTP(S) URL, or AWS S3 URL. The app is recommended for local testing and development.
