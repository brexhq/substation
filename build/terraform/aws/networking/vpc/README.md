<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | ~> 1.2 |
| <a name="requirement_aws"></a> [aws](#requirement\_aws) | ~> 5.0 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_aws"></a> [aws](#provider\_aws) | ~> 5.0 |

## Modules

No modules.

## Resources

| Name | Type |
|------|------|
| [aws_eip.eip](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/eip) | resource |
| [aws_internet_gateway.internet_gateway](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/internet_gateway) | resource |
| [aws_nat_gateway.nat_gateway](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/nat_gateway) | resource |
| [aws_route_table.private_route](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/route_table) | resource |
| [aws_route_table.public_route](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/route_table) | resource |
| [aws_route_table_association.private_subnet_association](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/route_table_association) | resource |
| [aws_route_table_association.public_subnet_association](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/route_table_association) | resource |
| [aws_subnet.private_subnet](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/subnet) | resource |
| [aws_subnet.public_subnet](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/subnet) | resource |
| [aws_vpc.vpc](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/vpc) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_config"></a> [config](#input\_config) | Configuration for the VPC:<br><br>    * cidr\_block: The CIDR block for the VPC. Defaults to 10.0.0.0/16.<br>    * public\_subnet: A map of CIDR blocks to availability zones for the public subnets. Defaults to 10.0.0.0/18 in us-east-1a.<br>    * private\_subnets: A map of CIDR blocks to availability zones for the private subnets. Defaults to 10.0.64.0/18 in us-east-1a, 10.0.128.0/18 in us-east-1b, and 10.0.192.0/18 in us-east-1c. | <pre>object({<br>    cidr_block = optional(string, "10.0.0.0/16")<br>    public_subnet = optional(map(string), {<br>      "10.0.0.0/18" = "us-east-1a"<br>    })<br>    private_subnets = optional(map(string), {<br>      "10.0.64.0/18"  = "us-east-1a"<br>      "10.0.128.0/18" = "us-east-1b"<br>      "10.0.192.0/18" = "us-east-1c"<br>    })<br>  })</pre> | <pre>{<br>  "cidr_block": "10.0.0.0/16",<br>  "private_subnets": {<br>    "10.0.128.0/18": "us-east-1b",<br>    "10.0.192.0/18": "us-east-1c",<br>    "10.0.64.0/18": "us-east-1a"<br>  },<br>  "public_subnet": {<br>    "10.0.0.0/18": "us-east-1a"<br>  }<br>}</pre> | no |
| <a name="input_tags"></a> [tags](#input\_tags) | Tags to apply to all resources. | `map(any)` | `{}` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_default_security_group_id"></a> [default\_security\_group\_id](#output\_default\_security\_group\_id) | The ID of the default security group. |
| <a name="output_private_subnet_ids"></a> [private\_subnet\_ids](#output\_private\_subnet\_ids) | The IDs of the private subnets. |
| <a name="output_public_subnet_id"></a> [public\_subnet\_id](#output\_public\_subnet\_id) | The ID of the public subnet. |
| <a name="output_vpc_id"></a> [vpc\_id](#output\_vpc\_id) | The ID of the VPC. |
<!-- END_TF_DOCS -->