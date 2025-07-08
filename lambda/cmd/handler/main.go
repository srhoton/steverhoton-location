// Package main provides the Lambda function handler for location operations.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/steverhoton/location-lambda/internal/handler"
	"github.com/steverhoton/location-lambda/internal/repository"
)

// getEnvVar retrieves an environment variable or returns a default value.
func getEnvVar(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// initializeHandler creates and configures the AppSync handler.
func initializeHandler(ctx context.Context) (*handler.AppSyncHandler, error) {
	// Get table name from environment
	tableName := os.Getenv("DYNAMODB_TABLE_NAME")
	if tableName == "" {
		return nil, fmt.Errorf("DYNAMODB_TABLE_NAME environment variable is required")
	}

	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create DynamoDB client
	dynamoClient := dynamodb.NewFromConfig(cfg)

	// Create repository
	repo := repository.NewDynamoDBRepository(dynamoClient, tableName)

	// Create handler
	return handler.NewAppSyncHandler(repo), nil
}

// lambdaHandler handles the Lambda invocation.
func lambdaHandler(ctx context.Context, event handler.AppSyncEvent) (interface{}, error) {
	// Initialize handler
	h, err := initializeHandler(ctx)
	if err != nil {
		log.Printf("ERROR: Failed to initialize handler: %v", err)
		return nil, fmt.Errorf("initialization error: %w", err)
	}

	// Log the incoming event
	log.Printf("INFO: Processing AppSync event - Field: %s", event.Field)

	// Handle the event
	result, err := h.Handle(ctx, event)
	if err != nil {
		log.Printf("ERROR: Failed to handle event: %v", err)
		return nil, err
	}

	log.Printf("INFO: Successfully processed event")
	return result, nil
}

func main() {
	// Start the Lambda handler
	lambda.Start(lambdaHandler)
}
