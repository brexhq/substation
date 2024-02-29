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
| [aws_iam_policy.access](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_policy) | resource |
| [aws_iam_role_policy_attachment.access](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role_policy_attachment) | resource |
| [aws_sqs_queue.queue](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/sqs_queue) | resource |
| [random_uuid.id](https://registry.terraform.io/providers/hashicorp/random/latest/docs/resources/uuid) | resource |
| [aws_iam_policy_document.access](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_access"></a> [access](#input\_access) | List of IAM ARNs that are granted access to the resource. These are typically the IAM role name output by other modules. | `list(string)` | `[]` | no |
| <a name="input_config"></a> [config](#input\_config) | Configuration for the SQS queue:<br><br>    * name:    The name of the queue.<br>    * delay:   The time in seconds that the delivery of all messages in the queue will be delayed. An integer from 0 to 900 (15 minutes). The default is 0.<br>    * timeout: The visibility timeout for the queue. An integer from 0 to 43200 (12 hours). The default is 30. | <pre>object({<br>    name    = string<br>    delay   = optional(number, 0)<br>    timeout = optional(number, 30)<br>  })</pre> | n/a | yes |
| <a name="input_kms"></a> [kms](#input\_kms) | Customer managed KMS key used to encrypt messages in the queue. If not provided, then no server-side encryption is used. See https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-server-side-encryption.html for more information. | <pre>object({<br>    arn = string<br>    id  = string<br>  })</pre> | `null` | no |
| <a name="input_tags"></a> [tags](#input\_tags) | Tags to apply to all resources. | `map(any)` | `{}` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_arn"></a> [arn](#output\_arn) | The ARN of the SQS queue. |
| <a name="output_id"></a> [id](#output\_id) | The ID of the SQS queue. |
<!-- END_TF_DOCS -->