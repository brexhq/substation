# cmd

Contains applications (apps) used in Substation deployments. Apps are organized by either the infrastructure they are deployed to (e.g., AWS) or the source of the data (e.g., file, http).

Any app that implements the ingest, transform, load (ITL) functionality of Substation is named `substation` and shares the same configuration file format (see [build/config/](/build/config/) for more information).

## app.go

Contains the core Substation application code. This code can be used to create new Substation applications.

## aws/lambda

Contains apps deployed as AWS Lambda functions.

## development/

Contains apps that aid in testing and development.
