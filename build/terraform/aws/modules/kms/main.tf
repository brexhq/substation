resource "aws_kms_key" "key" {
  enable_key_rotation = true
  # KMS key policies can be complex due to the potentially wide access KMS requires, so we let the user define the policy, otherwise the default KMS policy is applied
  policy = var.policy
  tags   = var.tags
}

resource "aws_kms_alias" "key_alias" {
  name          = var.name
  target_key_id = aws_kms_key.key.key_id
}
