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
| [aws_cloudwatch_metric_alarm.metric_alarm_downscale](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/cloudwatch_metric_alarm) | resource |
| [aws_cloudwatch_metric_alarm.metric_alarm_upscale](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/cloudwatch_metric_alarm) | resource |
| [aws_iam_policy.access](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_policy) | resource |
| [aws_iam_role_policy_attachment.access](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role_policy_attachment) | resource |
| [aws_kinesis_stream.stream](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kinesis_stream) | resource |
| [random_uuid.id](https://registry.terraform.io/providers/hashicorp/random/latest/docs/resources/uuid) | resource |
| [aws_iam_policy_document.access](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_access"></a> [access](#input\_access) | List of IAM ARNs that are granted access to the resource. | `list(string)` | `[]` | no |
| <a name="input_config"></a> [config](#input\_config) | Configuration for the Kinesis Data Stream:<br><br>    * name: The name of the Kinesis Data Stream.<br>    * autoscaling\_topic: The ARN of the SNS topic that will be used for autoscaling.<br>    * shards: The number of shards to create for the stream. Defaults to 2.<br>    * retention: The number of hours to retain data records in the stream. Defaults to 24. | <pre>object({<br>    name              = string<br>    autoscaling_topic = string<br>    shards            = optional(number, 2)<br>    retention         = optional(number, 24)<br>  })</pre> | n/a | yes |
| <a name="input_kms"></a> [kms](#input\_kms) | KMS key used to encrypt the stream. If not provided, then no server-side encryption is used. See https://docs.aws.amazon.com/streams/latest/dev/what-is-sse.html for more information. | <pre>object({<br>    arn = string<br>    id  = string<br>  })</pre> | `null` | no |
| <a name="input_tags"></a> [tags](#input\_tags) | Tags to apply to all resources. | `map(any)` | `{}` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_arn"></a> [arn](#output\_arn) | The ARN of the Kinesis Stream. |
| <a name="output_name"></a> [name](#output\_name) | The name of the Kinesis Stream. |
<!-- END_TF_DOCS -->