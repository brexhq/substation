# aws

This directory contains scripts for AWS deployments.

## ecr_login.sh

Bash script that is used to log into the AWS Elastic Container Registry (ECR). After logging in, containers can be pushed to ECR using `docker push [image]` .

## lambda/get_appconfig_extension.sh

Bash script that is used to download the [AWS AppConfig Lambda extension](https://docs.aws.amazon.com/appconfig/latest/userguide/appconfig-integration-lambda-extensions.html) for any AWS region. This extension is required for deploying Substation to AWS Lambda.
