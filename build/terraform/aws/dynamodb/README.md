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
| [aws_appautoscaling_policy.read_policy](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/appautoscaling_policy) | resource |
| [aws_appautoscaling_policy.write_policy](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/appautoscaling_policy) | resource |
| [aws_appautoscaling_target.read_target](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/appautoscaling_target) | resource |
| [aws_appautoscaling_target.write_target](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/appautoscaling_target) | resource |
| [aws_dynamodb_table.table](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/dynamodb_table) | resource |
| [aws_iam_policy.access](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_policy) | resource |
| [aws_iam_role_policy_attachment.access](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role_policy_attachment) | resource |
| [random_uuid.id](https://registry.terraform.io/providers/hashicorp/random/latest/docs/resources/uuid) | resource |
| [aws_iam_policy_document.access](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_access"></a> [access](#input\_access) | List of IAM ARNs that are granted access to the resource. | `list(string)` | `[]` | no |
| <a name="input_config"></a> [config](#input\_config) | Configuration for the DynamoDB table:<br><br>    * name:         The name of the table.<br>    * hash\_key:     The name of the hash key (aka Partition Key).<br>    * range\_key:    The name of the range key (aka Sort Key).<br>    * ttl:          The name of the attribute to use for TTL.<br>    * attributes:   A list of attributes for the table. The first attribute is the hash key, and the second is the range key.<br>    * read\_capacity:  The read capacity settings for the table.<br>    * write\_capacity: The write capacity settings for the table.<br>    * stream\_view\_type: The type of data from the table to be written to the stream. Valid values are NEW\_IMAGE, OLD\_IMAGE, NEW\_AND\_OLD\_IMAGES, and KEYS\_ONLY. The default value is NEW\_AND\_OLD\_IMAGES. See https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_StreamSpecification.html for more information. | <pre>object({<br>    name     = string<br>    hash_key = string<br>    attributes = list(object({<br>      name = string<br>      type = string<br>    }))<br><br>    range_key = optional(string, null)<br>    ttl       = optional(string, null)<br>    read_capacity = optional(object({<br>      min    = optional(number, 5)<br>      max    = optional(number, 1000)<br>      target = optional(number, 70)<br>    }))<br>    write_capacity = optional(object({<br>      min    = optional(number, 5)<br>      max    = optional(number, 1000)<br>      target = optional(number, 70)<br>    }))<br><br>    # Change Data Capture via Streams is enabled by default for the table.<br>    # https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Streams.html<br>    stream_view_type = optional(string, "NEW_AND_OLD_IMAGES")<br>  })</pre> | n/a | yes |
| <a name="input_kms"></a> [kms](#input\_kms) | Customer managed KMS key used to encrypt the table. If not provided, then an AWS owned key is used. See https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/EncryptionAtRest.html for more information. | <pre>object({<br>    arn = string<br>    id  = string<br>  })</pre> | `null` | no |
| <a name="input_tags"></a> [tags](#input\_tags) | Tags to apply to all resources. | `map(any)` | `{}` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_arn"></a> [arn](#output\_arn) | The ARN of the DynamoDB table. |
| <a name="output_stream_arn"></a> [stream\_arn](#output\_stream\_arn) | The ARN of the DynamoDB table stream. |
<!-- END_TF_DOCS -->