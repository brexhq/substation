resource "aws_iam_policy_attachment" "policy_attachment" {
  name       = "${var.id}_attachment"
  roles      = var.roles
  policy_arn = aws_iam_policy.policy.arn
}

resource "aws_iam_policy" "policy" {
  name        = "${var.id}_policy"
  description = "${var.id} policy"
  policy      = var.policy
}
