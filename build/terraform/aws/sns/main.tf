resource "random_uuid" "id" {}

resource "aws_sns_topic" "topic" {
  name                        = var.config.name
  kms_master_key_id           = var.kms != null ? var.kms.id : null
  fifo_topic                  = endswith(var.config.name, ".fifo") ? true : false
  content_based_deduplication = endswith(var.config.name, ".fifo") ? true : false

  tags = var.tags
}

# Applies the policy to each role in the access list.
resource "aws_iam_role_policy_attachment" "access" {
  count      = length(var.access)
  role       = var.access[count.index]
  policy_arn = aws_iam_policy.access.arn
}

resource "aws_iam_policy" "access" {
  name        = "substation-sns-${resource.random_uuid.id.id}"
  description = "Policy that grants access to the Substation ${var.config.name} SNS topic."
  policy      = data.aws_iam_policy_document.access.json
}

data "aws_iam_policy_document" "access" {
  statement {
    effect = "Allow"
    actions = [
      "sns:Publish",
    ]

    resources = [
      aws_sns_topic.topic.arn,
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
