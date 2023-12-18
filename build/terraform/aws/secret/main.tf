data "aws_region" "current" {}

resource "aws_secretsmanager_secret" "secret" {
  name       = var.config.secret.name
  kms_key_id = var.kms.id
  tags       = var.tags
}

# Applies the policy to each role in the access list.
resource "aws_iam_role_policy_attachment" "access" {
  count      = length(var.access)
  role       = var.access[count.index]
  policy_arn = aws_iam_policy.access.arn
}

resource "aws_iam_policy" "access" {
  name        = "sub-secret-access-${var.config.name}-${data.aws_region.current.name}"
  description = "Policy that grants access to the Substation ${var.config.name} secret."
  policy      = data.aws_iam_policy_document.access.json
}

data "aws_iam_policy_document" "access" {
  statement {
    effect = "Allow"
    actions = [
      "kms:Decrypt",
      "kms:GenerateDataKey"
    ]

    resources = [
      var.kms.arn,
    ]
  }

  statement {
    effect = "Allow"
    actions = [
      "secretsmanager:GetSecretValue",
    ]

    resources = [
      aws_secretsmanager_secret.secret.arn,
    ]
  }
}
