# Main configuration file for Location Lambda Infrastructure
# This file serves as the primary entry point and contains only essential
# configurations. Resource definitions are organized in separate files:
#
# - locals.tf     - Local values and computed configurations
# - dynamodb.tf   - DynamoDB table and related resources
# - iam.tf        - IAM roles, policies, and attachments
# - lambda.tf     - Lambda function and build process
# - cloudwatch.tf - CloudWatch logging resources
# - providers.tf  - Provider configurations
# - variables.tf  - Input variables
# - outputs.tf    - Output values
# - versions.tf   - Terraform and provider version constraints

# Data sources and additional configurations can be added here as needed