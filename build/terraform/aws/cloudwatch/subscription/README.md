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
| [aws_cloudwatch_log_subscription_filter.subscription_filter](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/cloudwatch_log_subscription_filter) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_config"></a> [config](#input\_config) | Configuration for the CloudWatch subscription filter:<br><br>    * name: The name of the CloudWatch subscription filter.<br>    * destination\_arn: The ARN of the CloudWatch destination.<br>    * log\_groups: The list of log groups to associate with the subscription filter.<br>    * filter\_pattern: The filter pattern to use for the subscription filter. If not provided, all log events are sent to the destination. | <pre>object({<br>    name            = string<br>    destination_arn = string<br>    log_groups      = list(string)<br>    filter_pattern  = optional(string, "")<br><br>  })</pre> | n/a | yes |
| <a name="input_tags"></a> [tags](#input\_tags) | n/a | `map(any)` | `{}` | no |

## Outputs

No outputs.
<!-- END_TF_DOCS -->