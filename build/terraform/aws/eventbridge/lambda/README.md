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
| [aws_cloudwatch_event_rule.rule](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/cloudwatch_event_rule) | resource |
| [aws_cloudwatch_event_target.target](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/cloudwatch_event_target) | resource |
| [aws_iam_policy.access](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_policy) | resource |
| [aws_iam_role_policy_attachment.access](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role_policy_attachment) | resource |
| [aws_lambda_permission.allow_cloudwatch](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/lambda_permission) | resource |
| [random_uuid.id](https://registry.terraform.io/providers/hashicorp/random/latest/docs/resources/uuid) | resource |
| [aws_caller_identity.current](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/caller_identity) | data source |
| [aws_iam_policy_document.access](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_access"></a> [access](#input\_access) | List of IAM ARNs that are granted access to the resource. | `list(string)` | `[]` | no |
| <a name="input_config"></a> [config](#input\_config) | Configuration for the EventBridge Lambda rule:<br><br>    * name:         The name of the rule.<br>    * description:  The description of the rule.<br>    * function:     The Lambda function to invoke when the rule is triggered.<br>    * event\_bus\_arn: The ARN of the event bus to associate with the rule. If not provided, the default event bus is used.<br>    * event\_pattern: The event pattern for the rule. If not provided, the rule is schedule-based.<br>    * schedule:     The schedule expression for the rule. If not provided, the rule is event-based. | <pre>object({<br>    name        = string<br>    description = string<br>    function = object({<br>      arn  = string<br>      name = string<br>    })<br>    <br>    # Optional<br>    event_bus_arn = optional(string, null)<br>    event_pattern = optional(string, null)<br>    schedule = optional(string, null)<br>  })</pre> | n/a | yes |
| <a name="input_tags"></a> [tags](#input\_tags) | Tags to apply to all resources. | `map(any)` | `{}` | no |

## Outputs

No outputs.
<!-- END_TF_DOCS -->