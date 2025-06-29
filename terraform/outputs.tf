output "lambda_function_arn" {
  description = "ARN of the Lambda function"
  value       = aws_lambda_function.location_handler.arn
}

output "lambda_function_name" {
  description = "Name of the Lambda function"
  value       = aws_lambda_function.location_handler.function_name
}

output "lambda_invoke_arn" {
  description = "Invoke ARN of the Lambda function"
  value       = aws_lambda_function.location_handler.invoke_arn
}

output "dynamodb_table_name" {
  description = "Name of the DynamoDB table"
  value       = aws_dynamodb_table.locations.name
}

output "dynamodb_table_arn" {
  description = "ARN of the DynamoDB table"
  value       = aws_dynamodb_table.locations.arn
}

output "dynamodb_gsi_name" {
  description = "Name of the DynamoDB Global Secondary Index"
  value       = var.dynamodb_gsi_name
}

output "lambda_role_arn" {
  description = "ARN of the Lambda execution role"
  value       = aws_iam_role.lambda_execution_role.arn
}