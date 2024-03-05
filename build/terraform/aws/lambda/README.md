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
| <a name="provider_random"></a> [random](#provider\_random) | n/a |

## Modules

No modules.

## Resources

| Name | Type |
|------|------|
| [aws_appconfig_configuration_profile.config](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/appconfig_configuration_profile) | resource |
| [aws_iam_policy.access](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_policy) | resource |
| [aws_iam_policy.custom_policy](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_policy) | resource |
| [aws_iam_role.role](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role) | resource |
| [aws_iam_role_policy_attachment.access](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role_policy_attachment) | resource |
| [aws_iam_role_policy_attachment.custom_policy_attachment](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role_policy_attachment) | resource |
| [aws_iam_role_policy_attachment.lambda_basic_execution_role](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role_policy_attachment) | resource |
| [aws_iam_role_policy_attachment.lambda_vpc_access_execution_role](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role_policy_attachment) | resource |
| [aws_iam_role_policy_attachment.xray_write_only_access](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role_policy_attachment) | resource |
| [aws_lambda_function.lambda_function](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/lambda_function) | resource |
| [random_uuid.id](https://registry.terraform.io/providers/hashicorp/random/latest/docs/resources/uuid) | resource |
| [aws_iam_policy_document.access](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |
| [aws_iam_policy_document.policy](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |
| [aws_iam_policy_document.service_policy_document](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_access"></a> [access](#input\_access) | List of IAM ARNs that are granted access to the resource. | `list(string)` | `[]` | no |
| <a name="input_appconfig"></a> [appconfig](#input\_appconfig) | AppConfig application used for configuring the function. If not provided, then no AppConfig configuration will be created for the function. | <pre>object({<br>    arn = string<br>    id  = string<br>    lambda = optional(object({<br>      name = string<br>      arn  = string<br>      role = object({<br>        name = string<br>        arn  = string<br>      })<br>    }))<br>  })</pre> | `null` | no |
| <a name="input_config"></a> [config](#input\_config) | Configuration for the Lambda function:<br><br>    * name: The name of the Lambda function.<br>    * description: The description of the Lambda function.<br>    * image\_uri: The URI of the container image that contains the function code.<br>    * image\_arm: Determines whether the image is an ARM64 image.<br>    * timeout: The amount of time that Lambda allows a function to run before stopping it. The default is 300 seconds.<br>    * memory: The amount of memory that your function has access to. The default is 1024 MB.<br>    * env: A map that defines environment variables for the function.<br>    * vpc\_config: A map that defines the VPC configuration for the function.<br>    * iam\_statements: A list of custom IAM policy statements to attach to the function's role. | <pre>object({<br>    name        = string<br>    description = string<br>    image_uri   = string<br>    image_arm   = bool<br>    timeout     = optional(number, 300)<br>    memory      = optional(number, 1024)<br>    env         = optional(map(any), null)<br>    vpc_config = optional(object({<br>      subnet_ids         = list(string)<br>      security_group_ids = list(string)<br>      }), {<br>      subnet_ids         = []<br>      security_group_ids = []<br>    })<br>    iam_statements = optional(list(object({<br>      sid       = string<br>      actions   = list(string)<br>      resources = list(string)<br>    })), [])<br>  })</pre> | n/a | yes |
| <a name="input_kms"></a> [kms](#input\_kms) | Customer managed KMS key used to encrypt the function's environment variables. If not provided, then an AWS managed key is used. See https://docs.aws.amazon.com/lambda/latest/dg/security-dataprotection.html#security-privacy-atrest for more information. | <pre>object({<br>    arn = string<br>    id  = string<br>  })</pre> | `null` | no |
| <a name="input_tags"></a> [tags](#input\_tags) | Tags to apply to all resources. | `map(any)` | `{}` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_arn"></a> [arn](#output\_arn) | The ARN of the Lambda function. |
| <a name="output_name"></a> [name](#output\_name) | The name of the Lambda function. |
| <a name="output_role"></a> [role](#output\_role) | The IAM role used by the Lambda function. |
<!-- END_TF_DOCS -->