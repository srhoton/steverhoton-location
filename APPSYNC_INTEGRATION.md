# AppSync Lambda Resolver Integration Guide

This document explains how to integrate the deployed Go Lambda function (`location-prod-location-handler`) with AWS AppSync as a Lambda resolver for location operations.

## Overview

The Lambda function is designed to handle AppSync direct Lambda resolvers for CRUD operations on location data. It processes AppSync events and interacts with the DynamoDB table (`location-prod-locations`) to manage location records.

## Prerequisites

- AWS AppSync GraphQL API created
- Lambda function deployed via Terraform (see `terraform/` directory)
- Appropriate IAM permissions configured

## GraphQL Schema

Add the following types and operations to your AppSync GraphQL schema:

```graphql
# Location Types
enum LocationType {
  address
  coordinates
}

# Address Type
type Address {
  streetAddress: String!
  streetAddress2: String
  city: String!
  stateProvince: String
  postalCode: String!
  country: String!
}

# Coordinates Type
type Coordinates {
  latitude: Float!
  longitude: Float!
  altitude: Float
  accuracy: Float
}

# Location Interface
interface Location {
  accountId: String!
  locationType: LocationType!
  extendedAttributes: AWSJSON
}

# Concrete Location Types
type AddressLocation implements Location {
  accountId: String!
  locationType: LocationType!
  extendedAttributes: AWSJSON
  address: Address!
}

type CoordinatesLocation implements Location {
  accountId: String!
  locationType: LocationType!
  extendedAttributes: AWSJSON
  coordinates: Coordinates!
}

# Union Type for Location Results
union LocationResult = AddressLocation | CoordinatesLocation

# Input Types
input AddressInput {
  streetAddress: String!
  streetAddress2: String
  city: String!
  stateProvince: String
  postalCode: String!
  country: String!
}

input CoordinatesInput {
  latitude: Float!
  longitude: Float!
  altitude: Float
  accuracy: Float
}

input CreateAddressLocationInput {
  accountId: String!
  address: AddressInput!
  extendedAttributes: AWSJSON
}

input CreateCoordinatesLocationInput {
  accountId: String!
  coordinates: CoordinatesInput!
  extendedAttributes: AWSJSON
}

input UpdateAddressLocationInput {
  accountId: String!
  address: AddressInput!
  extendedAttributes: AWSJSON
}

input UpdateCoordinatesLocationInput {
  accountId: String!
  coordinates: CoordinatesInput!
  extendedAttributes: AWSJSON
}

# List Result Type
type LocationListResult {
  locations: [LocationResult!]!
  nextCursor: String
}

# List Options Input
input ListLocationsInput {
  limit: Int
  cursor: String
}

# Root Types
type Query {
  getLocation(accountId: String!, locationId: String!): LocationResult
  listLocations(accountId: String!, options: ListLocationsInput): LocationListResult!
}

type Mutation {
  createAddressLocation(input: CreateAddressLocationInput!): String!
  createCoordinatesLocation(input: CreateCoordinatesLocationInput!): String!
  updateAddressLocation(locationId: String!, input: UpdateAddressLocationInput!): Boolean!
  updateCoordinatesLocation(locationId: String!, input: UpdateCoordinatesLocationInput!): Boolean!
  deleteLocation(accountId: String!, locationId: String!): Boolean!
}
```

## Lambda Data Source Configuration

### 1. Create Lambda Data Source

In the AppSync console or via CDK/CloudFormation:

```yaml
# CloudFormation example
LocationLambdaDataSource:
  Type: AWS::AppSync::DataSource
  Properties:
    ApiId: !Ref YourGraphQLApi
    Name: LocationLambdaDataSource
    Type: AWS_LAMBDA
    Description: "Lambda data source for location operations"
    LambdaConfig:
      LambdaFunctionArn: "arn:aws:lambda:us-east-1:705740530616:function:location-prod-location-handler"
    ServiceRoleArn: !GetAtt AppSyncLambdaRole.Arn
```

### 2. IAM Role for AppSync

Create an IAM role that allows AppSync to invoke the Lambda function:

```yaml
AppSyncLambdaRole:
  Type: AWS::IAM::Role
  Properties:
    AssumeRolePolicyDocument:
      Version: '2012-10-17'
      Statement:
        - Effect: Allow
          Principal:
            Service: appsync.amazonaws.com
          Action: sts:AssumeRole
    Policies:
      - PolicyName: AppSyncLambdaPolicy
        PolicyDocument:
          Version: '2012-10-17'
          Statement:
            - Effect: Allow
              Action:
                - lambda:InvokeFunction
              Resource: "arn:aws:lambda:us-east-1:705740530616:function:location-prod-location-handler"
```

## Resolver Configuration

Configure direct Lambda resolvers for each GraphQL operation. The Lambda function expects specific field names in the AppSync event.

### Query Resolvers

#### getLocation Resolver

**Field**: `Query.getLocation`  
**Data Source**: LocationLambdaDataSource  
**Request Mapping**: Direct Lambda Invocation  
**Response Mapping**: Direct Lambda Invocation  

The Lambda function expects:
```json
{
  "field": "getLocation",
  "arguments": {
    "accountId": "string",
    "locationId": "string"
  }
}
```

#### listLocations Resolver

**Field**: `Query.listLocations`  
**Data Source**: LocationLambdaDataSource  
**Request Mapping**: Direct Lambda Invocation  
**Response Mapping**: Direct Lambda Invocation  

The Lambda function expects:
```json
{
  "field": "listLocations", 
  "arguments": {
    "accountId": "string",
    "options": {
      "limit": 20,
      "cursor": "optional_cursor_string"
    }
  }
}
```

### Mutation Resolvers

#### createAddressLocation Resolver

**Field**: `Mutation.createAddressLocation`  
**Data Source**: LocationLambdaDataSource  
**Request Mapping**: Direct Lambda Invocation  
**Response Mapping**: Direct Lambda Invocation  

The Lambda function expects:
```json
{
  "field": "createAddressLocation",
  "arguments": {
    "input": {
      "accountId": "string",
      "address": {
        "streetAddress": "string",
        "streetAddress2": "optional_string",
        "city": "string", 
        "stateProvince": "optional_string",
        "postalCode": "string",
        "country": "string"
      },
      "extendedAttributes": {}
    }
  }
}
```

#### createCoordinatesLocation Resolver

**Field**: `Mutation.createCoordinatesLocation`  
**Data Source**: LocationLambdaDataSource  
**Request Mapping**: Direct Lambda Invocation  
**Response Mapping**: Direct Lambda Invocation  

The Lambda function expects:
```json
{
  "field": "createCoordinatesLocation",
  "arguments": {
    "input": {
      "accountId": "string",
      "coordinates": {
        "latitude": 0.0,
        "longitude": 0.0,
        "altitude": 0.0,
        "accuracy": 0.0
      },
      "extendedAttributes": {}
    }
  }
}
```

#### updateAddressLocation Resolver

**Field**: `Mutation.updateAddressLocation`  
**Data Source**: LocationLambdaDataSource  
**Request Mapping**: Direct Lambda Invocation  
**Response Mapping**: Direct Lambda Invocation  

The Lambda function expects:
```json
{
  "field": "updateAddressLocation",
  "arguments": {
    "locationId": "string",
    "input": {
      "accountId": "string",
      "address": {
        "streetAddress": "string",
        "streetAddress2": "optional_string", 
        "city": "string",
        "stateProvince": "optional_string",
        "postalCode": "string",
        "country": "string"
      },
      "extendedAttributes": {}
    }
  }
}
```

#### updateCoordinatesLocation Resolver

**Field**: `Mutation.updateCoordinatesLocation`  
**Data Source**: LocationLambdaDataSource  
**Request Mapping**: Direct Lambda Invocation  
**Response Mapping**: Direct Lambda Invocation  

The Lambda function expects:
```json
{
  "field": "updateCoordinatesLocation", 
  "arguments": {
    "locationId": "string",
    "input": {
      "accountId": "string",
      "coordinates": {
        "latitude": 0.0,
        "longitude": 0.0,
        "altitude": 0.0,
        "accuracy": 0.0
      },
      "extendedAttributes": {}
    }
  }
}
```

#### deleteLocation Resolver

**Field**: `Mutation.deleteLocation`  
**Data Source**: LocationLambdaDataSource  
**Request Mapping**: Direct Lambda Invocation  
**Response Mapping**: Direct Lambda Invocation  

The Lambda function expects:
```json
{
  "field": "deleteLocation",
  "arguments": {
    "accountId": "string", 
    "locationId": "string"
  }
}
```

## Lambda Function Response Format

The Lambda function returns different response formats based on the operation:

### Query Responses

**getLocation**: Returns the location object or null if not found
```json
{
  "accountId": "string",
  "locationType": "address|coordinates", 
  "extendedAttributes": {},
  "address": { /* address object */ },
  "coordinates": { /* coordinates object */ }
}
```

**listLocations**: Returns paginated results
```json
{
  "locations": [/* array of location objects */],
  "nextCursor": "optional_cursor_for_next_page"
}
```

### Mutation Responses

**create operations**: Return the generated location ID (UUID string)
```json
"550e8400-e29b-41d4-a716-446655440000"
```

**update/delete operations**: Return boolean success indicator
```json
true
```

## Error Handling

The Lambda function returns standard GraphQL errors for various scenarios:

- **Validation Errors**: Invalid input data (e.g., invalid coordinates, missing required fields)
- **Not Found Errors**: Location doesn't exist or access denied
- **Server Errors**: Internal processing errors

Example error response:
```json
{
  "errorType": "ValidationError",
  "errorMessage": "Invalid email format"
}
```

## Testing

### Sample Queries

```graphql
# Get a specific location
query GetLocation {
  getLocation(accountId: "user123", locationId: "550e8400-e29b-41d4-a716-446655440000") {
    ... on AddressLocation {
      accountId
      locationType
      address {
        streetAddress
        city
        postalCode
        country
      }
    }
    ... on CoordinatesLocation {
      accountId
      locationType
      coordinates {
        latitude
        longitude
        altitude
        accuracy
      }
    }
  }
}

# List locations for an account
query ListLocations {
  listLocations(accountId: "user123", options: { limit: 10 }) {
    locations {
      ... on AddressLocation {
        accountId
        locationType
        address {
          streetAddress
          city
          country
        }
      }
      ... on CoordinatesLocation {
        accountId
        locationType
        coordinates {
          latitude
          longitude
        }
      }
    }
    nextCursor
  }
}
```

### Sample Mutations

```graphql
# Create an address location
mutation CreateAddressLocation {
  createAddressLocation(input: {
    accountId: "user123"
    address: {
      streetAddress: "123 Main St"
      city: "San Francisco"
      stateProvince: "CA"
      postalCode: "94105"
      country: "US"
    }
  })
}

# Create a coordinates location
mutation CreateCoordinatesLocation {
  createCoordinatesLocation(input: {
    accountId: "user123"
    coordinates: {
      latitude: 37.7749
      longitude: -122.4194
      accuracy: 10.0
    }
  })
}

# Update an address location
mutation UpdateAddressLocation {
  updateAddressLocation(
    locationId: "550e8400-e29b-41d4-a716-446655440000"
    input: {
      accountId: "user123"
      address: {
        streetAddress: "456 Oak Ave"
        city: "San Francisco" 
        stateProvince: "CA"
        postalCode: "94102"
        country: "US"
      }
    }
  )
}

# Delete a location
mutation DeleteLocation {
  deleteLocation(accountId: "user123", locationId: "550e8400-e29b-41d4-a716-446655440000")
}
```

## Security Considerations

1. **Authentication**: Ensure AppSync API has proper authentication configured (API Key, Cognito, IAM, etc.)
2. **Authorization**: Implement field-level authorization rules to ensure users can only access their own locations
3. **Account ID Validation**: The Lambda function relies on the provided accountId - ensure this comes from authenticated context
4. **Input Validation**: The Lambda function validates all inputs, but additional GraphQL validation rules can be added
5. **Rate Limiting**: Consider implementing rate limiting at the AppSync level

## Monitoring and Logging

- **CloudWatch Logs**: Lambda function logs are available at `/aws/lambda/location-prod-location-handler`
- **AppSync Logs**: Enable AppSync logging for request/response monitoring
- **DynamoDB Metrics**: Monitor table performance via CloudWatch
- **Lambda Metrics**: Monitor function performance, errors, and duration

## Troubleshooting

### Common Issues

1. **Permission Errors**: Verify AppSync role has permission to invoke Lambda
2. **Schema Mismatches**: Ensure GraphQL schema matches Lambda expected inputs
3. **DynamoDB Errors**: Check Lambda has proper permissions for DynamoDB operations
4. **Validation Failures**: Review Lambda logs for detailed validation error messages

### Debug Steps

1. Check AppSync request/response logs
2. Review Lambda CloudWatch logs for detailed error information
3. Verify DynamoDB table exists and has correct structure
4. Test Lambda function directly with sample AppSync events
5. Validate IAM permissions for all components

## Next Steps

1. **Deploy AppSync API**: Create and configure your AppSync GraphQL API
2. **Add Resolvers**: Configure all the resolvers as documented above
3. **Test Operations**: Use the GraphQL playground to test all operations
4. **Add Authorization**: Implement proper authorization rules for your use case
5. **Monitor Performance**: Set up monitoring and alerting for production usage