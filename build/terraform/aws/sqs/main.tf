data "aws_region" "current" {}

locals {
  read_access = [
    "sqs:ReceiveMessage",
    "sqs:DeleteMessage",
    "sqs:GetQueue*",
  ]

  write_access = [
    "sqs:SendMessage*",
  ]
}

resource "aws_sqs_queue" "queue" {
  name                              = var.config.name
  delay_seconds                     = var.config.delay
  visibility_timeout_seconds        = var.config.timeout
  kms_master_key_id                 = var.kms.id
  kms_data_key_reuse_period_seconds = 300
  fifo_queue                        = endswith(var.config.name, ".fifo") ? true : false
  content_based_deduplication       = endswith(var.config.name, ".fifo") ? true : false

  tags = var.tags
}

# Applies the policy to each role in the access list.
resource "aws_iam_role_policy_attachment" "access" {
  count      = length(var.access)
  role       = var.access[count.index]
  policy_arn = aws_iam_policy.access.arn
}

resource "aws_iam_policy" "access" {
  name        = "sub-sqs-access-${var.config.name}-${data.aws_region.current.name}"
  description = "Policy that grants access to the Substation ${var.config.name} SQS queue."
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
    actions = concat(
      local.read_access,
      local.write_access,
    )

    resources = [
      aws_sqs_queue.queue.arn,
    ]
  }
}
