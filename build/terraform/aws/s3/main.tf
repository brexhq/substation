resource "random_uuid" "id" {}

resource "aws_s3_bucket" "bucket" {
  bucket              = var.config.name
  force_destroy       = var.config.force_destroy
  object_lock_enabled = var.config.compliance != null ? true : false

  tags = var.tags
}

resource "aws_s3_bucket_ownership_controls" "bucket" {
  bucket = aws_s3_bucket.bucket.id
  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}

resource "aws_s3_bucket_acl" "bucket" {
  bucket = aws_s3_bucket.bucket.id
  acl    = "private"

  depends_on = [
    aws_s3_bucket_ownership_controls.bucket,
  ]
}

resource "aws_s3_bucket_server_side_encryption_configuration" "encryption" {
  count = var.kms != null ? 1 : 0

  bucket = aws_s3_bucket.bucket.bucket

  rule {
    apply_server_side_encryption_by_default {
      kms_master_key_id = var.kms.arn
      sse_algorithm     = "aws:kms"
    }
  }
}

resource "aws_s3_bucket_object_lock_configuration" "object_lock" {
  count = var.config.compliance != null ? 1 : 0

  bucket = aws_s3_bucket.bucket.bucket

  rule {
    default_retention {
      mode = "COMPLIANCE"
      days = var.config.compliance.retention
    }
  }
}

# Applies the policy to each role in the access list.
resource "aws_iam_role_policy_attachment" "access" {
  count      = length(var.access)
  role       = var.access[count.index]
  policy_arn = aws_iam_policy.access.arn
}

resource "aws_iam_policy" "access" {
  name        = "substation-s3-${resource.random_uuid.id.id}"
  description = "Policy that grants access to the Substation ${var.config.name} S3 bucket."
  policy      = data.aws_iam_policy_document.access.json
}

data "aws_iam_policy_document" "access" {
  statement {
    effect = "Allow"
    actions = [
      "s3:GetObject",
      "s3:PutObject",
    ]

    resources = [
      aws_s3_bucket.bucket.arn,
      "${aws_s3_bucket.bucket.arn}/*",
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
