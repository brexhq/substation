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
| [aws_appconfig_application.app](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/appconfig_application) | resource |
| [aws_appconfig_deployment_strategy.instant](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/appconfig_deployment_strategy) | resource |
| [aws_appconfig_environment.env](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/appconfig_environment) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_config"></a> [config](#input\_config) | Configuration for the AppConfig application:<br><br>    * name: The name of the AppConfig application.<br>    * environments: A list of environments to create for the AppConfig application. | <pre>object({<br>    name = string<br>    environments = list(object({<br>      name = string<br>    }))<br>  })</pre> | n/a | yes |
| <a name="input_tags"></a> [tags](#input\_tags) | n/a | `map(any)` | `{}` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_arn"></a> [arn](#output\_arn) | The ARN of the AppConfig application. |
| <a name="output_id"></a> [id](#output\_id) | The ID of the AppConfig application. |
<!-- END_TF_DOCS -->