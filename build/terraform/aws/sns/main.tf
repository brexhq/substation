resource "aws_sns_topic" "topic" {
  name                        = var.name
  kms_master_key_id           = var.kms_key_id
  fifo_topic                  = var.fifo ? true : false
  content_based_deduplication = var.fifo ? true : false

  tags = var.tags
}
