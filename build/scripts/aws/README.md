# aws

This directory contains scripts that are used to manage Substation deployments in AWS.

## ecr_login.sh

Bash script that is used to log into the AWS Elastic Container Registry (ECR). After logging in, containers can be pushed to ECR using `docker push [image]` .

## appconfig/appconfig_delete.py

Python script that deletes application profiles, including all hosted configurations, in AWS AppConfig. This script has some dependencies:

* boto3 must be installed
* AWS credentials for reading from and deleting in AppConfig

This script can be used after deleting any AWS Lambda that have application profiles.

## appconfig/appconfig_upload.py

Python script that manages the upload and deployment of compiled Substation JSON configuration files in AWS AppConfig. This script has some dependencies:

* boto3 must be installed
* AWS credentials for reading from and writing to AppConfig
* AppConfig infrastructure must be ready to use (see [examples/aws/terraform/bootstrap.tf](/examples/aws/terraform/bootstrap.tf) for an example)

This script is intended to be deployed to a CI / CD pipeline (e.g., GitHub Actions, Circle CI, Jenkins, etc.), but can be run locally if needed. See [examples/aws/](/examples/aws/) for example usage.

## s3/s3_rehydration.py

Python script that rehydrates data from an S3 bucket into an SNS topic by simulating S3 
object creation events. This script has some dependencies:

* boto3 must be installed
* AWS credentials for reading from S3 and writing to SNS

## lambda/get_appconfig_extension.sh

Bash script that is used to download the [AWS AppConfig Lambda extension](https://docs.aws.amazon.com/appconfig/latest/userguide/appconfig-integration-lambda-extensions.html) for any AWS region. This extension is required for deploying Substation to AWS Lambda.
