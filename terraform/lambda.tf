# Data source for Lambda build
data "archive_file" "lambda_zip" {
  type             = "zip"
  source_dir       = "${path.module}/../lambda/build"
  output_path      = "${path.module}/lambda-deployment.zip"
  output_file_mode = "0666"

  depends_on = [null_resource.lambda_build]
}

# Null resource to build the Go Lambda
resource "null_resource" "lambda_build" {
  triggers = {
    # Rebuild when Go source files change
    go_mod_hash   = filemd5("${path.module}/../lambda/go.mod")
    go_sum_hash   = filemd5("${path.module}/../lambda/go.sum")
    source_hash   = filesha256("${path.module}/../lambda/cmd/handler/main.go")
    makefile_hash = filemd5("${path.module}/../lambda/Makefile")
  }

  provisioner "local-exec" {
    command     = "make clean && make build"
    working_dir = "${path.module}/../lambda"
  }
}

# Lambda function
resource "aws_lambda_function" "location_handler" {
  filename         = data.archive_file.lambda_zip.output_path
  function_name    = local.function_name_full
  role             = aws_iam_role.lambda_execution_role.arn
  handler          = "bootstrap"
  source_code_hash = data.archive_file.lambda_zip.output_base64sha256
  runtime          = var.lambda_runtime
  timeout          = var.lambda_timeout
  memory_size      = var.lambda_memory_size

  architectures = [var.lambda_architecture]

  environment {
    variables = {
      DYNAMODB_TABLE_NAME = aws_dynamodb_table.locations.name
      DYNAMODB_GSI_NAME   = var.dynamodb_gsi_name
      GO_VERSION          = var.go_version
    }
  }

  depends_on = [
    aws_iam_role_policy_attachment.lambda_basic_execution,
    aws_iam_role_policy_attachment.lambda_dynamodb_policy_attachment,
    aws_cloudwatch_log_group.lambda_logs
  ]

  tags = merge(
    local.common_tags,
    {
      Name = local.function_name_full
    }
  )
}