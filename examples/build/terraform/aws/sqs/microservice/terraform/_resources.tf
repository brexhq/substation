data "aws_caller_identity" "caller" {}

module "appconfig" {
  source = "../../../../../../../build/terraform/aws/appconfig"

  config = {
    name        = "substation"
    environments = [{
      name = "example"
    }]
  }
}

module "ecr" {
  source = "../../../../../../../build/terraform/aws/ecr"

  config = {
    name         = "substation"
    force_delete = true
  }
}

module "sqs" {
  source = "../../../../../../../build/terraform/aws/sqs"

  config = {
    name = "substation"
  }

  access = [
    # Reads from SQS.
    module.microservice.role.name,
    # Writes to SQS.
    module.frontend.role.name,
  ]
}

module "dynamodb" {
  source = "../../../../../../../build/terraform/aws/dynamodb"

  config = {
    name     = "substation"
    hash_key = "PK"
    attributes = [
      {
        name = "PK"
        type = "S"
      }
    ]
  }

  access = [
    # Writes to DynamoDB.
    module.microservice.role.name,
  ]
}
