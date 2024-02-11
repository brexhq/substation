resource "random_uuid" "id" {}

resource "aws_secretsmanager_secret" "secret" {
  name       = var.config.name
  kms_key_id = var.kms != null ? var.kms.id : null
  tags       = var.tags
}

# Applies the policy to each role in the access list.
resource "aws_iam_role_policy_attachment" "access" {
  count      = length(var.access)
  role       = var.access[count.index]
  policy_arn = aws_iam_policy.access.arn
}

resource "aws_iam_policy" "access" {
  name        = "substation-secret-access-${resource.random_uuid.id.id}"
  description = "Policy that grants access to the Substation ${var.config.name} secret."
  policy      = data.aws_iam_policy_document.access.json
}

data "aws_iam_policy_document" "access" {
  statement {
    effect = "Allow"
    actions = [
      "secretsmanager:GetSecretValue",
    ]

    resources = [
      aws_secretsmanager_secret.secret.arn,
    ]
  }

  dynamic "statement" {
    for_each = var.kms != null ? [1] : []

    content {
      effect = "Allow"
      actions = [
        "kms:Decrypt",
        "kms:GenerateDataKey"
      ]

      resources = [
        var.kms.arn,
      ]
    }
  }
}
