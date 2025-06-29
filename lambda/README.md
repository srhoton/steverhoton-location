# Location Lambda Function

A Go AWS Lambda function that provides CRUD operations for location records through AWS AppSync.

## Overview

This Lambda function handles location records that can be either:
- **Address-based locations**: Specified by mailing address components
- **GPS-based locations**: Specified by latitude/longitude coordinates

Each location is associated with an account ID and supports custom extended attributes.

## Architecture

The Lambda function follows a clean architecture pattern:

```
cmd/
└── handler/           # Main Lambda entry point
internal/
├── models/           # Domain models and validation
├── repository/       # DynamoDB data access layer
└── handler/          # AppSync event handling
```

## Features

- **Complete CRUD operations** for location records
- **JSON schema validation** against the defined location schema
- **Type-safe Go models** with comprehensive validation
- **DynamoDB integration** with optimized queries
- **AppSync event handling** for GraphQL operations
- **Comprehensive test coverage** with mocks
- **Linting and formatting** following Go best practices

## Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `DYNAMODB_TABLE_NAME` | Name of the DynamoDB table | Yes |

## DynamoDB Table Structure

The function expects a DynamoDB table with the following key structure:
- **Partition Key (PK)**: `ACCOUNT#{accountId}`
- **Sort Key (SK)**: `LOCATION#{locationId}`

## AppSync Operations

The Lambda supports the following GraphQL operations:

### createLocation
Creates a new location record.

**Arguments:**
```json
{
  "input": {
    "accountId": "string",
    "locationType": "address|coordinates",
    "address": { /* address fields */ },
    "coordinates": { /* GPS coordinates */ },
    "extendedAttributes": { /* custom attributes */ }
  }
}
```

### getLocation
Retrieves a location by account ID and location ID.

**Arguments:**
```json
{
  "accountId": "string",
  "locationId": "string"
}
```

### updateLocation
Updates an existing location record.

**Arguments:**
```json
{
  "locationId": "string",
  "input": { /* location data */ }
}
```

### deleteLocation
Deletes a location record.

**Arguments:**
```json
{
  "accountId": "string",
  "locationId": "string"
}
```

### listLocations
Lists all locations for an account.

**Arguments:**
```json
{
  "accountId": "string"
}
```

## Building and Deployment

### Prerequisites
- Go 1.21 or later
- AWS CLI configured
- golangci-lint for linting

### Build Commands

```bash
# Run all tests
make test

# Run linter
make lint

# Build for AWS Lambda
make build

# Create deployment package
make zip

# Run all checks and build
make all
```

### Local Development

```bash
# Install dependencies
make deps

# Run tests with coverage
make test-coverage

# Format code
make fmt

# Run all development checks
make dev
```

## Testing

The project includes comprehensive tests for all components:

- **Unit tests** for models, repository, and handlers
- **Mock-based testing** for external dependencies
- **Table-driven tests** for validation logic
- **Integration test patterns** for repository operations

Run tests with:
```bash
go test -v ./...
```

## Code Quality

The project follows Go best practices:

- **Gofmt** formatting
- **golangci-lint** with comprehensive rules
- **go vet** static analysis
- **Comprehensive error handling**
- **Proper logging**
- **Security best practices**

## Performance Considerations

- **Efficient DynamoDB queries** using partition and sort keys
- **Minimal memory allocations** in hot paths
- **Context-aware operations** with proper timeouts
- **Connection pooling** via AWS SDK v2

## Security

- **Input validation** against JSON schema
- **Type-safe unmarshaling** to prevent injection
- **Error sanitization** to prevent information leakage
- **AWS IAM integration** for access control

## Schema Compliance

The Lambda function validates all location records against the JSON schema defined in `../config/location-schema.json`, ensuring:

- **Data integrity**
- **Consistent structure**
- **Type safety**
- **Required field validation**
- **Mutually exclusive location types**