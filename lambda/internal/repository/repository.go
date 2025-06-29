// Package repository provides DynamoDB operations for location records.
package repository

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/steverhoton/location-lambda/internal/models"
)

// ListResult represents the result of a paginated list operation.
type ListResult struct {
	Locations  []models.Location `json:"locations"`
	NextCursor *string           `json:"nextCursor,omitempty"`
}

// ListOptions contains options for listing operations.
type ListOptions struct {
	Limit  *int32  `json:"limit,omitempty"`
	Cursor *string `json:"cursor,omitempty"`
}

// Repository defines the interface for location storage operations.
type Repository interface {
	Create(ctx context.Context, location models.Location) (string, error)
	Get(ctx context.Context, accountID, locationID string) (models.Location, error)
	Update(ctx context.Context, location models.Location, locationID string) error
	Delete(ctx context.Context, accountID, locationID string) error
	List(ctx context.Context, accountID string, options *ListOptions) (*ListResult, error)
}

// DynamoDBRepository implements Repository using DynamoDB.
type DynamoDBRepository struct {
	client      DynamoDBClient
	tableName   string
	gsiName     string
	defaultLimit int32
}

// NewDynamoDBRepository creates a new DynamoDB repository.
func NewDynamoDBRepository(client DynamoDBClient, tableName, gsiName string) *DynamoDBRepository {
	return &DynamoDBRepository{
		client:       client,
		tableName:    tableName,
		gsiName:      gsiName,
		defaultLimit: 20,
	}
}

// locationRecord represents a location record in DynamoDB.
type locationRecord struct {
	PK                 string                 `dynamodbav:"PK"`                 // locationId (UUID) - this IS the locationId
	SK                 string                 `dynamodbav:"SK"`                 // accountId
	AccountID          string                 `dynamodbav:"accountId"`          // accountId (for GSI)
	LocationType       models.LocationType    `dynamodbav:"locationType"`
	ExtendedAttributes map[string]interface{} `dynamodbav:"extendedAttributes,omitempty"`
	Address            *models.Address        `dynamodbav:"address,omitempty"`
	Coordinates        *models.Coordinates    `dynamodbav:"coordinates,omitempty"`
}

// paginationCursor represents the cursor for pagination.
type paginationCursor struct {
	PK        string `json:"pk"`  // This is the locationId (UUID)
	SK        string `json:"sk"`  // This is the accountId
	AccountID string `json:"accountId"`
}

// toLocationRecord converts a Location to a DynamoDB record.
func toLocationRecord(location models.Location, locationID string) (*locationRecord, error) {
	record := &locationRecord{
		PK:                 locationID,                         // UUID as PK (this IS the locationId)
		SK:                 location.GetAccountID(),           // accountId as SK
		AccountID:          location.GetAccountID(),           // accountId as attribute (for GSI)
		LocationType:       location.GetLocationType(),
		ExtendedAttributes: location.GetExtendedAttributes(),
	}

	switch loc := location.(type) {
	case models.AddressLocation:
		record.Address = &loc.Address
	case models.CoordinatesLocation:
		record.Coordinates = &loc.Coordinates
	default:
		return nil, errors.New("unknown location type")
	}

	return record, nil
}

// toLocation converts a DynamoDB record to a Location.
func (r *locationRecord) toLocation() (models.Location, error) {
	base := models.LocationBase{
		AccountID:          r.AccountID,
		LocationType:       r.LocationType,
		ExtendedAttributes: r.ExtendedAttributes,
	}

	switch r.LocationType {
	case models.LocationTypeAddress:
		if r.Address == nil {
			return nil, errors.New("address is nil for address location type")
		}
		return models.AddressLocation{
			LocationBase: base,
			Address:      *r.Address,
		}, nil
	case models.LocationTypeCoordinates:
		if r.Coordinates == nil {
			return nil, errors.New("coordinates is nil for coordinates location type")
		}
		return models.CoordinatesLocation{
			LocationBase: base,
			Coordinates:  *r.Coordinates,
		}, nil
	default:
		return nil, fmt.Errorf("unknown location type: %s", r.LocationType)
	}
}

// encodeCursor encodes a pagination cursor to base64.
func (r *DynamoDBRepository) encodeCursor(cursor *paginationCursor) (*string, error) {
	if cursor == nil {
		return nil, nil
	}
	
	data, err := json.Marshal(cursor)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal cursor: %w", err)
	}
	
	encoded := base64.StdEncoding.EncodeToString(data)
	return &encoded, nil
}

// decodeCursor decodes a base64 pagination cursor.
func (r *DynamoDBRepository) decodeCursor(cursorStr *string) (*paginationCursor, error) {
	if cursorStr == nil || *cursorStr == "" {
		return nil, nil
	}
	
	data, err := base64.StdEncoding.DecodeString(*cursorStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode cursor: %w", err)
	}
	
	var cursor paginationCursor
	if err := json.Unmarshal(data, &cursor); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cursor: %w", err)
	}
	
	return &cursor, nil
}

// cursorToLastEvaluatedKey converts a cursor to DynamoDB LastEvaluatedKey.
func (r *DynamoDBRepository) cursorToLastEvaluatedKey(cursor *paginationCursor) map[string]types.AttributeValue {
	if cursor == nil {
		return nil
	}
	
	return map[string]types.AttributeValue{
		"PK":        &types.AttributeValueMemberS{Value: cursor.PK},        // PK is the locationId
		"SK":        &types.AttributeValueMemberS{Value: cursor.SK},        // SK is the accountId
		"accountId": &types.AttributeValueMemberS{Value: cursor.AccountID}, // accountId for GSI
	}
}

// lastEvaluatedKeyToCursor converts DynamoDB LastEvaluatedKey to a cursor.
func (r *DynamoDBRepository) lastEvaluatedKeyToCursor(lek map[string]types.AttributeValue) *paginationCursor {
	if lek == nil {
		return nil
	}
	
	cursor := &paginationCursor{}
	
	if pk, ok := lek["PK"]; ok {
		if s, ok := pk.(*types.AttributeValueMemberS); ok {
			cursor.PK = s.Value
		}
	}
	
	if sk, ok := lek["SK"]; ok {
		if s, ok := sk.(*types.AttributeValueMemberS); ok {
			cursor.SK = s.Value
		}
	}
	
	// PK already contains the locationId, so no need to extract it separately
	
	if accID, ok := lek["accountId"]; ok {
		if s, ok := accID.(*types.AttributeValueMemberS); ok {
			cursor.AccountID = s.Value
		}
	}
	
	return cursor
}

// Create creates a new location record and returns the location ID.
func (r *DynamoDBRepository) Create(ctx context.Context, location models.Location) (string, error) {
	if err := location.Validate(); err != nil {
		return "", fmt.Errorf("validation failed: %w", err)
	}

	// Generate a new UUID for location ID
	locationID := uuid.New().String()
	
	record, err := toLocationRecord(location, locationID)
	if err != nil {
		return "", fmt.Errorf("failed to convert location to record: %w", err)
	}

	av, err := attributevalue.MarshalMap(record)
	if err != nil {
		return "", fmt.Errorf("failed to marshal location: %w", err)
	}

	// Add condition to ensure the item doesn't already exist
	input := &dynamodb.PutItemInput{
		TableName:           aws.String(r.tableName),
		Item:                av,
		ConditionExpression: aws.String("attribute_not_exists(PK) AND attribute_not_exists(SK)"),
	}

	_, err = r.client.PutItem(ctx, input)
	if err != nil {
		var ccf *types.ConditionalCheckFailedException
		if errors.As(err, &ccf) {
			return "", fmt.Errorf("location already exists")
		}
		return "", fmt.Errorf("failed to create location: %w", err)
	}

	return locationID, nil
}

// Get retrieves a location by account ID and location ID.
func (r *DynamoDBRepository) Get(ctx context.Context, accountID, locationID string) (models.Location, error) {
	key := map[string]types.AttributeValue{
		"PK": &types.AttributeValueMemberS{Value: locationID},  // locationID as PK
		"SK": &types.AttributeValueMemberS{Value: accountID},  // accountID as SK
	}

	input := &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key:       key,
	}

	result, err := r.client.GetItem(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get location: %w", err)
	}

	if result.Item == nil {
		return nil, fmt.Errorf("location not found")
	}

	var record locationRecord
	if err := attributevalue.UnmarshalMap(result.Item, &record); err != nil {
		return nil, fmt.Errorf("failed to unmarshal location: %w", err)
	}

	return record.toLocation()
}

// Update updates an existing location.
func (r *DynamoDBRepository) Update(ctx context.Context, location models.Location, locationID string) error {
	if err := location.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	record, err := toLocationRecord(location, locationID)
	if err != nil {
		return fmt.Errorf("failed to convert location to record: %w", err)
	}

	av, err := attributevalue.MarshalMap(record)
	if err != nil {
		return fmt.Errorf("failed to marshal location: %w", err)
	}

	// Add condition to ensure the item exists and belongs to the correct account
	input := &dynamodb.PutItemInput{
		TableName:           aws.String(r.tableName),
		Item:                av,
		ConditionExpression: aws.String("attribute_exists(PK) AND attribute_exists(SK) AND accountId = :accountId"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":accountId": &types.AttributeValueMemberS{Value: location.GetAccountID()},
		},
	}

	_, err = r.client.PutItem(ctx, input)
	if err != nil {
		var ccf *types.ConditionalCheckFailedException
		if errors.As(err, &ccf) {
			return fmt.Errorf("location not found or access denied")
		}
		return fmt.Errorf("failed to update location: %w", err)
	}

	return nil
}

// Delete deletes a location.
func (r *DynamoDBRepository) Delete(ctx context.Context, accountID, locationID string) error {
	key := map[string]types.AttributeValue{
		"PK": &types.AttributeValueMemberS{Value: locationID},  // locationID as PK
		"SK": &types.AttributeValueMemberS{Value: accountID},  // accountID as SK
	}

	input := &dynamodb.DeleteItemInput{
		TableName:           aws.String(r.tableName),
		Key:                 key,
		ConditionExpression: aws.String("attribute_exists(PK) AND attribute_exists(SK) AND accountId = :accountId"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":accountId": &types.AttributeValueMemberS{Value: accountID},
		},
	}

	_, err := r.client.DeleteItem(ctx, input)
	if err != nil {
		var ccf *types.ConditionalCheckFailedException
		if errors.As(err, &ccf) {
			return fmt.Errorf("location not found or access denied")
		}
		return fmt.Errorf("failed to delete location: %w", err)
	}

	return nil
}

// List lists all locations for an account with cursor-based pagination.
func (r *DynamoDBRepository) List(ctx context.Context, accountID string, options *ListOptions) (*ListResult, error) {
	// Set default limit if not provided
	limit := r.defaultLimit
	if options != nil && options.Limit != nil {
		limit = *options.Limit
	}

	// Decode cursor if provided
	var startKey map[string]types.AttributeValue
	if options != nil && options.Cursor != nil {
		cursor, err := r.decodeCursor(options.Cursor)
		if err != nil {
			return nil, fmt.Errorf("failed to decode cursor: %w", err)
		}
		startKey = r.cursorToLastEvaluatedKey(cursor)
	}

	// Query the GSI to get all locations for the account
	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		IndexName:              aws.String(r.gsiName),
		KeyConditionExpression: aws.String("accountId = :accountId"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":accountId": &types.AttributeValueMemberS{Value: accountID},
		},
		Limit:                 aws.Int32(limit),
		ExclusiveStartKey:     startKey,
		ScanIndexForward:      aws.Bool(true), // Sort by locationId ascending for deterministic ordering
	}

	result, err := r.client.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to list locations: %w", err)
	}

	// Convert items to locations
	locations := make([]models.Location, 0, len(result.Items))
	for _, item := range result.Items {
		var record locationRecord
		if err := attributevalue.UnmarshalMap(item, &record); err != nil {
			return nil, fmt.Errorf("failed to unmarshal location: %w", err)
		}

		location, err := record.toLocation()
		if err != nil {
			return nil, fmt.Errorf("failed to convert record to location: %w", err)
		}

		locations = append(locations, location)
	}

	// Create next cursor if there are more items
	var nextCursor *string
	if result.LastEvaluatedKey != nil {
		cursor := r.lastEvaluatedKeyToCursor(result.LastEvaluatedKey)
		nextCursor, err = r.encodeCursor(cursor)
		if err != nil {
			return nil, fmt.Errorf("failed to encode cursor: %w", err)
		}
	}

	return &ListResult{
		Locations:  locations,
		NextCursor: nextCursor,
	}, nil
}

