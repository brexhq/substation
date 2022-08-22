# cmd

This directory contains applications (apps) used in Substation deployments. Apps are organized by either the infrastructure they are deployed to (e.g., AWS) or the source of the data (e.g., file, http).

Any app that implements the ingest, transform, load (ITL) functionality of Substation is named `substation` and shares the same configuration file format (see [config/](/config/) for more information).

## app.go

Contains the core Substation application code. This code can be used to create new Substation applications.

### design

Substation operates through a system of channels and concurrent goroutines; ingest, transform, and load are handled in separate goroutines and managed by the cmd that invokes the app. Below is a diagram that describes the execution model:

```
cmd/main loads configuration
cmd/main creates channels
cmd/main executes anonymous goroutine
  - anon goroutine creates waitgroups for sink and transform goroutines
  - anon goroutine executes sink goroutine
  - anon goroutine executes transform goroutines
  - anon goroutine sends data to the tranform channel
  - anon goroutine closes transform channel, waits for transform goroutines to finish
  - anon goroutine closes sink channel, waits for sink goroutine to finish
cmd/main blocks waiting for feedback from non-anonymous goroutines
```

This execution model was chosen for its ability to support horizontal scaling, high-latency data processing, and efficient delivery of data.

## aws/lambda

Contains Substation apps deployed as AWS Lambda functions. More information is available in 
[cmd/aws/lambda/README.md](/cmd/aws/lambda/README.md).

## file/substation

This app handles ITL for data stored in files. The app can be deployed anywhere, including non-container infrastructure, and is recommended for local testing of Substation configs.
