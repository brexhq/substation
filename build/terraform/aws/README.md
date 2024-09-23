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

Refer to each module's README for more information. 

## Guides

### Bootstrapping Deployments

Many resources can be shared across a single deployment, such as AppConfig configurations, ECR repositories, and S3 buckets:

```hcl

module "appconfig" {
  source = "build/terraform/aws/appconfig"

  config = {
    name         = "substation"
    environments = [{ name = "example" }]
  }
}

module "ecr" {
  source = "build/terraform/aws/ecr"

  config = {
    name         = "substation"
    force_delete = true
  }
}

module "s3" {
  source = "build/terraform/aws/s3"

  config = {
    # Bucket names must be globally unique.
    name = "substation-xxxxxx"
  }

  # Access is granted by providing the role name of a
  # resource. This access applies least privilege and
  # grants access to dependent resources, such as KMS.
  access = [
    module.node.role.name,
  ]
}
```

### Deploying Nodes

Substation [nodes](https://substation.readme.io/docs/nodes) are deployed as Lambda functions:

```hcl

module "node" {
  source    = "build/terraform/aws/lambda"
  appconfig = module.appconfig

  config = {
    name        = "node"
    description = "Substation node that is invoked by an API Gateway."
    image_uri   = "${module.ecr.url}:v2.0.0"
    image_arm   = true

    env = {
      "SUBSTATION_CONFIG" : "http://localhost:2772/applications/substation/environments/example/configurations/node"
      "SUBSTATION_LAMBDA_HANDLER" : "AWS_API_GATEWAY"
      "SUBSTATION_DEBUG" : true
    }
  }
}

module "node_gateway" {
  source = "build/terraform/aws/api_gateway/lambda"
  lambda = module.node

  config = {
    name = "node_gateway"
  }

  depends_on = [
    module.node
  ]
}
```

### Connecting Nodes

Nodes are connected by AWS services, such as an SQS queue:

```hcl

module "sqs" {
  source = "build/terraform/aws/sqs"

  config = {
    name = "substation"
  }

  access = [
    # Writes to the queue.
    module.node.role.name,
    # Reads from the queue.
    module.consumer.role.name,
  ]
}

module "consumer" {
  source    = "build/terraform/aws/lambda"
  appconfig = module.appconfig

  config = {
    name        = "consumer"
    description = "Substation consumer that is invoked by an SQS queue."
    image_uri   = "${module.ecr.url}:v2.0.0"
    image_arm   = true

    env = {
      "SUBSTATION_CONFIG" : "http://localhost:2772/applications/substation/environments/example/configurations/consumer"
      "SUBSTATION_LAMBDA_HANDLER" : "AWS_SQS"
      "SUBSTATION_DEBUG" : true
    }
  }
}

resource "aws_lambda_event_source_mapping" "consumer" {
  event_source_arn                   = module.sqs.arn
  function_name                      = module.consumer.arn
  maximum_batching_window_in_seconds = 10
  batch_size                         = 100
}
```
