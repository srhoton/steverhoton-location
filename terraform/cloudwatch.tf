# CloudWatch Log Group for Lambda
resource "aws_cloudwatch_log_group" "lambda_logs" {
  name              = "/aws/lambda/${local.function_name_full}"
  retention_in_days = 14

  tags = local.common_tags
}