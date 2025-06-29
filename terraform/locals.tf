locals {
  common_tags = merge(
    {
      Project     = var.project
      Environment = var.environment
      ManagedBy   = "terraform"
      Owner       = "engineering"
    },
    var.additional_tags
  )

  function_name_full = "${var.project}-${var.environment}-${var.lambda_function_name}"
  table_name_full    = "${var.project}-${var.environment}-${var.dynamodb_table_name}"
}