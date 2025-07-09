# DynamoDB table for locations
resource "aws_dynamodb_table" "locations" {
  name         = local.table_name_full
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "PK"
  range_key    = "SK"

  attribute {
    name = "PK"
    type = "S"
  }

  attribute {
    name = "SK"
    type = "S"
  }

  attribute {
    name = "accountId"
    type = "S"
  }

  global_secondary_index {
    name            = var.dynamodb_gsi_name
    hash_key        = "accountId"
    projection_type = "ALL"
  }

  point_in_time_recovery {
    enabled = true
  }

  server_side_encryption {
    enabled = true
  }

  tags = merge(
    local.common_tags,
    {
      Name = local.table_name_full
    }
  )
}