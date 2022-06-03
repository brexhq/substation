################################################
# API Gateway
# sends data to raw Kinesis stream
################################################

module "example_gateway_kinesis" {
  source = "./modules/api_gateway/kinesis"
  name   = "substation_kinesis_example"
  stream = "substation_example_raw"
}
