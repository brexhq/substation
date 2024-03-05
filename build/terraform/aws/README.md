# terraform

This directory contains Terraform modules for deploying Substation to AWS.

## Using Terraform

An overview on how Terraform is used and the write, plan, apply workflow is available [here](https://www.terraform.io/intro). Refer to the [AWS provider documentation](https://registry.terraform.io/providers/hashicorp/aws/latest/docs) for more information on using Terraform with AWS.

## Modules

Due to the potentially endless number of deployment configurations, Substation includes Terraform modules for these AWS services:

* API Gateway
* AppConfig
* CloudWatch
* DynamoDB
* Elastic Container Registry (ECR)
* Event Bridge
* Kinesis Data Streams
* Key Management Service (KMS)
* Lambda
* S3
* Secrets Manager
* SNS
* SQS
* VPC Networking

Refer to each module's README for more information. Several examples of how to use these modules are available [here](/examples/terraform/aws/).
