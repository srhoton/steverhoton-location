# Location Lambda Infrastructure

This Terraform configuration deploys the infrastructure for the Go-based location Lambda function including:

- AWS Lambda function for handling location operations
- DynamoDB table for storing location data with Global Secondary Index
- IAM roles and policies for secure Lambda execution
- CloudWatch Log Group for Lambda logging

## Architecture

- **Lambda Function**: Handles AppSync events for location CRUD operations
- **DynamoDB Table**: Stores location records with composite primary key (PK: locationId, SK: accountId)
- **Global Secondary Index**: Enables querying locations by accountId
- **IAM Role**: Provides minimal required permissions for Lambda to access DynamoDB

## Prerequisites

1. AWS CLI configured with appropriate credentials
2. Terraform >= 1.0 installed
3. Go 1.22+ installed (for building the Lambda)
4. Make installed (for building the Lambda)
5. S3 bucket `srhoton-tfstate` must exist for state storage

## Usage

### Initialize Terraform

```bash
terraform init
```

### Plan Deployment

```bash
terraform plan
```

### Deploy Infrastructure

```bash
terraform apply
```

### Destroy Infrastructure

```bash
terraform destroy
```

## Configuration

### Variables

The following variables can be customized:

| Variable | Description | Default |
|----------|-------------|---------|
| `aws_region` | AWS region for resources | `us-east-1` |
| `project` | Project name used for resource naming | `location` |
| `environment` | Environment name (e.g., dev, staging, prod) | `dev` |
| `lambda_function_name` | Name of the Lambda function | `location-handler` |
| `dynamodb_table_name` | Name of the DynamoDB table | `locations` |
| `dynamodb_gsi_name` | Name of the DynamoDB Global Secondary Index | `AccountIndex` |
| `lambda_timeout` | Lambda function timeout in seconds | `30` |
| `lambda_memory_size` | Lambda function memory size in MB | `256` |
| `lambda_runtime` | Lambda runtime | `provided.al2023` |
| `lambda_architecture` | Lambda function architecture | `x86_64` |

### Environment-specific Deployment

Create a `terraform.tfvars` file for environment-specific values:

```hcl
project     = "location"
environment = "prod"
aws_region  = "us-west-2"

lambda_memory_size = 512
lambda_timeout     = 60

additional_tags = {
  CostCenter = "engineering"
  Team       = "platform"
}
```

## DynamoDB Table Schema

### Primary Table

- **Hash Key (PK)**: `locationId` (String) - UUID
- **Range Key (SK)**: `accountId` (String)

### Global Secondary Index (AccountIndex)

- **Hash Key**: `accountId` (String)
- **Projection**: ALL

### Attributes

- `accountId` (String) - Account identifier
- `locationType` (String) - Type of location (address|coordinates)
- `extendedAttributes` (Map) - Optional additional attributes
- `address` (Map) - Address details (for address type locations)
- `coordinates` (Map) - GPS coordinates (for coordinate type locations)

## Lambda Environment Variables

The Lambda function receives the following environment variables:

- `DYNAMODB_TABLE_NAME`: Name of the DynamoDB table
- `DYNAMODB_GSI_NAME`: Name of the Global Secondary Index
- `GO_VERSION`: Go version used for building

## Outputs

| Output | Description |
|--------|-------------|
| `lambda_function_arn` | ARN of the Lambda function |
| `lambda_function_name` | Name of the Lambda function |
| `lambda_invoke_arn` | Invoke ARN of the Lambda function |
| `dynamodb_table_name` | Name of the DynamoDB table |
| `dynamodb_table_arn` | ARN of the DynamoDB table |
| `dynamodb_gsi_name` | Name of the DynamoDB Global Secondary Index |
| `lambda_role_arn` | ARN of the Lambda execution role |

## Build Process

The Terraform configuration automatically builds the Go Lambda binary when source files change:

1. Detects changes in go.mod, go.sum, main.go, and Makefile
2. Runs `make clean && make build` in the lambda directory
3. Creates a deployment zip file from the build directory
4. Updates the Lambda function with the new code

## Security Features

- IAM role with least-privilege permissions
- DynamoDB encryption at rest enabled
- Point-in-time recovery enabled for DynamoDB
- CloudWatch logging for monitoring and debugging
- Conditional expressions in DynamoDB operations for data integrity

## Monitoring

- CloudWatch logs are automatically created for the Lambda function
- Log retention is set to 14 days
- DynamoDB metrics are available in CloudWatch

## Cost Optimization

- DynamoDB uses PAY_PER_REQUEST billing mode
- Lambda uses ARM64 architecture (configurable) for cost efficiency
- CloudWatch log retention prevents indefinite log storage costs