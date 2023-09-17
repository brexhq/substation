resource "aws_sns_topic" "topic" {
  name                        = var.config.name
  kms_master_key_id           = var.kms.id
  fifo_topic                  = endswith(var.config.name, ".fifo") ? true : false
  content_based_deduplication = endswith(var.config.name, ".fifo") ? true : false

  tags = var.tags
}

# Applies the policy to each role in the access list.
resource "aws_iam_role_policy_attachment" "access" {
  for_each = toset(var.access)
  role = each.value
  policy_arn = aws_iam_policy.access.arn
}

resource "aws_iam_policy" "access" {
  name        = var.config.name
  description = "Policy for the ${var.config.name} SNS topic"
  policy      = data.aws_iam_policy_document.access.json
}

data "aws_iam_policy_document" "access" {
  statement {
    sid = "KMS"

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
    sid = "SNS"

    effect = "Allow"
    actions = [
      "sns:Publish",
    ]

    resources = [
      aws_sns_topic.topic.arn,
    ]
  }
}
