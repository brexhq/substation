################################################
# API Gateway
# sends data to raw Kinesis stream
################################################

module "gateway_kinesis_source" {
  source = "/workspaces/substation/build/terraform/aws/api_gateway/kinesis"
  name   = "substation_kinesis_example"
  stream = "substation_raw"
}
