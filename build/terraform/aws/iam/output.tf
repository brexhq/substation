output "appconfig_read_policy" {
  value = data.aws_iam_policy_document.appconfig_read.json
}

output "cloudwatch_modify_policy" {
  value = data.aws_iam_policy_document.cloudwatch_modify.json
}

output "cloudwatch_write_policy" {
  value = data.aws_iam_policy_document.cloudwatch_write.json
}

output "kinesis_modify_policy" {
  value = data.aws_iam_policy_document.kinesis_modify.json
}

output "kinesis_read_policy" {
  value = data.aws_iam_policy_document.kinesis_read.json
}

output "kinesis_write_policy" {
  value = data.aws_iam_policy_document.kinesis_write.json
}

output "kinesis_firehose_write_policy" {
  value = data.aws_iam_policy_document.kinesis_firehose_write.json
}

output "dynamodb_read_policy" {
  value = data.aws_iam_policy_document.dynamodb_read.json
}

output "dynamodb_write_policy" {
  value = data.aws_iam_policy_document.dynamodb_write.json
}

output "kms_read_policy" {
  value = data.aws_iam_policy_document.kms_read.json
}

output "kms_write_policy" {
  value = data.aws_iam_policy_document.kms_write.json
}

output "lambda_execute_policy" {
  value = data.aws_iam_policy_document.lambda_execute.json
}

output "s3_read_policy" {
  value = data.aws_iam_policy_document.s3_read.json
}

output "s3_write_policy" {
  value = data.aws_iam_policy_document.s3_write.json
}

output "secretsmanager_read_policy" {
  value = data.aws_iam_policy_document.secretsmanager_read.json
}

output "sqs_read_policy" {
  value = data.aws_iam_policy_document.sqs_read.json
}

output "sqs_write_policy" {
  value = data.aws_iam_policy_document.sqs_write.json
}
