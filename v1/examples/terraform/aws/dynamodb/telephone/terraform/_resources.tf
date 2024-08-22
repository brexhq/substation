data "aws_caller_identity" "caller" {}

module "appconfig" {
  source = "../../../../../../build/terraform/aws/appconfig"

  config = {
    name         = "substation"
    environments = [{ name = "example" }]
  }
}

module "ecr" {
  source = "../../../../../../build/terraform/aws/ecr"

  config = {
    name         = "substation"
    force_delete = true
  }
}

module "ecr_autoscale" {
  source = "../../../../../../build/terraform/aws/ecr"

  config = {
    name         = "autoscale"
    force_delete = true
  }
}

module "dynamodb" {
  source = "../../../../../../build/terraform/aws/dynamodb"

  config = {
    name      = "substation"
    hash_key  = "PK"
    range_key = "SK"
    ttl       = "TTL"

    attributes = [
      {
        name = "PK"
        type = "S"
      },
      {
        name = "SK"
        type = "S"
      },
    ]
  }

  access = [
    module.edr_enrichment.role.name,
    module.edr_transform.role.name,
    module.idp_enrichment.role.name,
    module.dvc_mgmt_enrichment.role.name,
  ]
}
