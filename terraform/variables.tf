variable "aws_region" {
  description = "AWS region for resources"
  type        = string
  default     = "us-east-1"
}

variable "project" {
  description = "Project name used for resource naming"
  type        = string
  default     = "location"
}

variable "environment" {
  description = "Environment name (e.g., dev, staging, prod)"
  type        = string
  default     = "prod"
}

variable "lambda_function_name" {
  description = "Name of the Lambda function"
  type        = string
  default     = "location-handler"
}

variable "dynamodb_table_name" {
  description = "Name of the DynamoDB table"
  type        = string
  default     = "locations"
}

variable "dynamodb_gsi_name" {
  description = "Name of the DynamoDB Global Secondary Index"
  type        = string
  default     = "AccountIndex"
}

variable "lambda_timeout" {
  description = "Lambda function timeout in seconds"
  type        = number
  default     = 30

  validation {
    condition     = var.lambda_timeout >= 1 && var.lambda_timeout <= 900
    error_message = "Lambda timeout must be between 1 and 900 seconds."
  }
}

variable "lambda_memory_size" {
  description = "Lambda function memory size in MB"
  type        = number
  default     = 256

  validation {
    condition     = var.lambda_memory_size >= 128 && var.lambda_memory_size <= 10240
    error_message = "Lambda memory size must be between 128 and 10240 MB."
  }
}

variable "lambda_runtime" {
  description = "Lambda runtime"
  type        = string
  default     = "provided.al2023"
}

variable "lambda_architecture" {
  description = "Lambda function architecture"
  type        = string
  default     = "x86_64"

  validation {
    condition     = contains(["x86_64", "arm64"], var.lambda_architecture)
    error_message = "Lambda architecture must be either x86_64 or arm64."
  }
}

variable "go_version" {
  description = "Go version for building the Lambda"
  type        = string
  default     = "1.22"
}

variable "additional_tags" {
  description = "Additional tags to apply to all resources"
  type        = map(string)
  default     = {}
}