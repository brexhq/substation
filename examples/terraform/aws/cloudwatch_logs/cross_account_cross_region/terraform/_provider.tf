provider "aws" {
  # profile = "default"
  region = "us-east-1"
}

provider "aws" {
  # profile = "default"
  alias  = "usw2"
  region = "us-west-2"
}
