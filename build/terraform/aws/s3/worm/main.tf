resource "aws_s3_bucket" "bucket" {
  bucket        = var.config.name
  force_destroy = var.config.force_destroy
  tags          = var.tags

  object_lock_configuration {
    object_lock_enabled = "Enabled"
  }
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
  bucket = aws_s3_bucket.bucket.bucket

  rule {
    apply_server_side_encryption_by_default {
      kms_master_key_id = var.kms.arn
      sse_algorithm     = "aws:kms"
    }
  }
}

resource "aws_s3_bucket_object_lock_configuration" "object_lock" {
  bucket = aws_s3_bucket.bucket.bucket

  rule {
    default_retention {
      mode = "COMPLIANCE"
      days = var.config.retention
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
  name        = "sub-s3-${var.config.name}"
  description = "Policy for the ${var.config.name} S3 bucket."
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
    sid = "S3"

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
}
