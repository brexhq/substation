
resource "aws_sqs_queue" "queue" {
  name                              = var.name
  delay_seconds                     = var.delay_seconds
  visibility_timeout_seconds        = var.visibility_timeout_seconds
  kms_master_key_id                 = var.kms_key_id
  kms_data_key_reuse_period_seconds = 300
  fifo_queue                        = endswith(var.name, ".fifo") ? true : false
  content_based_deduplication       = endswith(var.name, ".fifo") ? true : false

  tags = var.tags
}
