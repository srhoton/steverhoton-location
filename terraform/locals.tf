locals {
  common_tags = merge(
    {
      Project     = var.project_name
      Environment = var.environment
      ManagedBy   = "terraform"
      Owner       = "engineering"
    },
    var.additional_tags
  )

  function_name_full = "${var.project_name}-${var.environment}"
  table_name_full    = "${var.project_name}-${var.environment}"
}