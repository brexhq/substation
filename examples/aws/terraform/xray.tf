# Substation Lambdas use X-Ray for performance monitoring and tuning. If the AWS account's X-Ray data is not encrypted, then use this file to setup encryption. The ARN produced by `xray_key` (or the ARN of a previously created encryption key) must be added as a kms_read and kms_write resource on all Lambda IAM policies. Alternatively, if no encryption is wanted, then exclude this file and any references to the `xray_key` ARN from Lambda IAM policies.

# resource "aws_kms_key" "xray_key" {
#   description         = "KMS used for server-side encryption of X-Ray data"
#   enable_key_rotation = true
# }

# resource "aws_kms_alias" "xray_key_alias" {
#   name          = "alias/xray"
#   target_key_id = aws_kms_key.xray_key.key_id
# }

# # applying this configuration can take up to 5 minutes
# resource "aws_xray_encryption_config" "xray_encryption_config" {
#   type   = "KMS"
#   key_id = aws_kms_key.xray_key.arn
# }
