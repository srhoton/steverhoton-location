package repository

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/steverhoton/location-lambda/internal/models"
)

// mockDynamoDBClient is a mock implementation of the DynamoDB client.
type mockDynamoDBClient struct {
	mock.Mock
}

func (m *mockDynamoDBClient) PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dynamodb.PutItemOutput), args.Error(1)
}

func (m *mockDynamoDBClient) GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dynamodb.GetItemOutput), args.Error(1)
}

func (m *mockDynamoDBClient) DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dynamodb.DeleteItemOutput), args.Error(1)
}

func (m *mockDynamoDBClient) Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dynamodb.QueryOutput), args.Error(1)
}

func TestToLocationRecord(t *testing.T) {
	tests := []struct {
		name     string
		location models.Location
		locID    string
		wantErr  bool
		check    func(t *testing.T, record *locationRecord)
	}{
		{
			name: "Address location",
			location: models.AddressLocation{
				LocationBase: models.LocationBase{
					AccountID:    "acc-12345",
					LocationType: models.LocationTypeAddress,
					ExtendedAttributes: map[string]interface{}{
						"businessName": "Acme Corp",
					},
				},
				Address: models.Address{
					StreetAddress: "123 Main St",
					City:          "Springfield",
					PostalCode:    "12345",
					Country:       "US",
				},
			},
			locID:   "loc-001",
			wantErr: false,
			check: func(t *testing.T, record *locationRecord) {
				assert.Equal(t, "loc-001", record.PK)                    // PK is the locationID (UUID)
				assert.Equal(t, "acc-12345", record.SK)                  // accountID as SK
				assert.Equal(t, "acc-12345", record.AccountID)           // accountID as attribute
				assert.Equal(t, models.LocationTypeAddress, record.LocationType)
				assert.NotNil(t, record.Address)
				assert.Equal(t, "123 Main St", record.Address.StreetAddress)
				assert.Nil(t, record.Coordinates)
			},
		},
		{
			name: "Coordinates location",
			location: models.CoordinatesLocation{
				LocationBase: models.LocationBase{
					AccountID:    "acc-67890",
					LocationType: models.LocationTypeCoordinates,
					ExtendedAttributes: map[string]interface{}{
						"sensorType": "weather",
					},
				},
				Coordinates: models.Coordinates{
					Latitude:  40.7128,
					Longitude: -74.0060,
				},
			},
			locID:   "loc-002",
			wantErr: false,
			check: func(t *testing.T, record *locationRecord) {
				assert.Equal(t, "loc-002", record.PK)                    // PK is the locationID (UUID)
				assert.Equal(t, "acc-67890", record.SK)                  // accountID as SK
				assert.Equal(t, "acc-67890", record.AccountID)           // accountID as attribute
				assert.Equal(t, models.LocationTypeCoordinates, record.LocationType)
				assert.NotNil(t, record.Coordinates)
				assert.Equal(t, 40.7128, record.Coordinates.Latitude)
				assert.Nil(t, record.Address)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record, err := toLocationRecord(tt.location, tt.locID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, record)
				if tt.check != nil {
					tt.check(t, record)
				}
			}
		})
	}
}

func TestLocationRecordToLocation(t *testing.T) {
	tests := []struct {
		name    string
		record  locationRecord
		wantErr bool
		check   func(t *testing.T, loc models.Location)
	}{
		{
			name: "Address location record",
			record: locationRecord{
				PK:        "loc-001",      // PK is the locationID (UUID)
				SK:        "acc-12345",    // accountID as SK
				AccountID: "acc-12345",    // accountID as attribute
				LocationType: models.LocationTypeAddress,
				ExtendedAttributes: map[string]interface{}{
					"businessName": "Acme Corp",
				},
				Address: &models.Address{
					StreetAddress: "123 Main St",
					City:          "Springfield",
					PostalCode:    "12345",
					Country:       "US",
				},
			},
			wantErr: false,
			check: func(t *testing.T, loc models.Location) {
				assert.IsType(t, models.AddressLocation{}, loc)
				addrLoc := loc.(models.AddressLocation)
				assert.Equal(t, "acc-12345", addrLoc.AccountID)
				assert.Equal(t, models.LocationTypeAddress, addrLoc.LocationType)
				assert.Equal(t, "123 Main St", addrLoc.Address.StreetAddress)
			},
		},
		{
			name: "Coordinates location record",
			record: locationRecord{
				PK:        "loc-002",      // PK is the locationID (UUID)
				SK:        "acc-67890",    // accountID as SK
				AccountID: "acc-67890",    // accountID as attribute
				LocationType: models.LocationTypeCoordinates,
				ExtendedAttributes: map[string]interface{}{
					"sensorType": "weather",
				},
				Coordinates: &models.Coordinates{
					Latitude:  40.7128,
					Longitude: -74.0060,
				},
			},
			wantErr: false,
			check: func(t *testing.T, loc models.Location) {
				assert.IsType(t, models.CoordinatesLocation{}, loc)
				coordLoc := loc.(models.CoordinatesLocation)
				assert.Equal(t, "acc-67890", coordLoc.AccountID)
				assert.Equal(t, models.LocationTypeCoordinates, coordLoc.LocationType)
				assert.Equal(t, 40.7128, coordLoc.Coordinates.Latitude)
			},
		},
		{
			name: "Invalid - address location without address",
			record: locationRecord{
				PK:           "loc-001",
				SK:           "acc-12345",
				AccountID:    "acc-12345",
				LocationType: models.LocationTypeAddress,
				Address:      nil,
			},
			wantErr: true,
		},
		{
			name: "Invalid - coordinates location without coordinates",
			record: locationRecord{
				PK:           "loc-002",
				SK:           "acc-67890",
				AccountID:    "acc-67890",
				LocationType: models.LocationTypeCoordinates,
				Coordinates:  nil,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc, err := tt.record.toLocation()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, loc)
				if tt.check != nil {
					tt.check(t, loc)
				}
			}
		})
	}
}

func TestDynamoDBRepositoryCreate(t *testing.T) {
	ctx := context.Background()
	mockClient := new(mockDynamoDBClient)
	repo := NewDynamoDBRepository(mockClient, "test-table", "test-gsi")

	location := models.AddressLocation{
		LocationBase: models.LocationBase{
			AccountID:    "acc-12345",
			LocationType: models.LocationTypeAddress,
		},
		Address: models.Address{
			StreetAddress: "123 Main St",
			City:          "Springfield",
			PostalCode:    "12345",
			Country:       "US",
		},
	}

	t.Run("Successful create", func(t *testing.T) {
		mockClient.On("PutItem", ctx, mock.MatchedBy(func(input *dynamodb.PutItemInput) bool {
			return *input.TableName == "test-table" &&
				input.ConditionExpression != nil &&
				*input.ConditionExpression == "attribute_not_exists(PK) AND attribute_not_exists(SK)"
		})).Return(&dynamodb.PutItemOutput{}, nil).Once()

		locationID, err := repo.Create(ctx, location)
		assert.NoError(t, err)
		assert.NotEmpty(t, locationID)
		// Verify it's a valid UUID format (36 characters with hyphens)
		assert.Len(t, locationID, 36)
		mockClient.AssertExpectations(t)
	})

	t.Run("Validation error", func(t *testing.T) {
		invalidLocation := models.AddressLocation{
			LocationBase: models.LocationBase{
				AccountID:    "", // Invalid - empty account ID
				LocationType: models.LocationTypeAddress,
			},
			Address: models.Address{
				StreetAddress: "123 Main St",
				City:          "Springfield",
				PostalCode:    "12345",
				Country:       "US",
			},
		}

		locationID, err := repo.Create(ctx, invalidLocation)
		assert.Error(t, err)
		assert.Empty(t, locationID)
		assert.Contains(t, err.Error(), "validation failed")
	})

	t.Run("Item already exists", func(t *testing.T) {
		mockClient.On("PutItem", ctx, mock.Anything).Return(
			nil,
			&types.ConditionalCheckFailedException{Message: aws.String("The conditional request failed")},
		).Once()

		locationID, err := repo.Create(ctx, location)
		assert.Error(t, err)
		assert.Empty(t, locationID)
		assert.Contains(t, err.Error(), "location already exists")
		mockClient.AssertExpectations(t)
	})
}

func TestDynamoDBRepositoryGet(t *testing.T) {
	ctx := context.Background()
	mockClient := new(mockDynamoDBClient)
	repo := NewDynamoDBRepository(mockClient, "test-table", "test-gsi")

	accountID := "acc-12345"
	locationID := "loc-001"

	t.Run("Successful get", func(t *testing.T) {
		item := map[string]types.AttributeValue{
			"PK":           &types.AttributeValueMemberS{Value: "loc-001"},      // PK is the locationID (UUID)
			"SK":           &types.AttributeValueMemberS{Value: "acc-12345"},   // accountID as SK
			"accountId":    &types.AttributeValueMemberS{Value: "acc-12345"},  // accountID as attribute
			"locationType": &types.AttributeValueMemberS{Value: "address"},
			"address": &types.AttributeValueMemberM{
				Value: map[string]types.AttributeValue{
					"streetAddress": &types.AttributeValueMemberS{Value: "123 Main St"},
					"city":          &types.AttributeValueMemberS{Value: "Springfield"},
					"postalCode":    &types.AttributeValueMemberS{Value: "12345"},
					"country":       &types.AttributeValueMemberS{Value: "US"},
				},
			},
		}

		mockClient.On("GetItem", ctx, mock.MatchedBy(func(input *dynamodb.GetItemInput) bool {
			return *input.TableName == "test-table"
		})).Return(&dynamodb.GetItemOutput{Item: item}, nil).Once()

		location, err := repo.Get(ctx, accountID, locationID)
		require.NoError(t, err)
		require.NotNil(t, location)
		assert.IsType(t, models.AddressLocation{}, location)
		mockClient.AssertExpectations(t)
	})

	t.Run("Item not found", func(t *testing.T) {
		mockClient.On("GetItem", ctx, mock.Anything).Return(
			&dynamodb.GetItemOutput{Item: nil}, nil,
		).Once()

		location, err := repo.Get(ctx, accountID, locationID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "location not found")
		assert.Nil(t, location)
		mockClient.AssertExpectations(t)
	})
}

func TestDynamoDBRepositoryUpdate(t *testing.T) {
	ctx := context.Background()
	mockClient := new(mockDynamoDBClient)
	repo := NewDynamoDBRepository(mockClient, "test-table", "test-gsi")

	location := models.AddressLocation{
		LocationBase: models.LocationBase{
			AccountID:    "acc-12345",
			LocationType: models.LocationTypeAddress,
		},
		Address: models.Address{
			StreetAddress: "456 Oak Ave",
			City:          "Springfield",
			PostalCode:    "12345",
			Country:       "US",
		},
	}
	locationID := "loc-001"

	t.Run("Successful update", func(t *testing.T) {
		mockClient.On("PutItem", ctx, mock.MatchedBy(func(input *dynamodb.PutItemInput) bool {
			return *input.TableName == "test-table" &&
				input.ConditionExpression != nil &&
				*input.ConditionExpression == "attribute_exists(PK) AND attribute_exists(SK) AND accountId = :accountId" &&
				input.ExpressionAttributeValues != nil &&
				len(input.ExpressionAttributeValues) == 1
		})).Return(&dynamodb.PutItemOutput{}, nil).Once()

		err := repo.Update(ctx, location, locationID)
		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("Item not found", func(t *testing.T) {
		mockClient.On("PutItem", ctx, mock.Anything).Return(
			nil,
			&types.ConditionalCheckFailedException{Message: aws.String("The conditional request failed")},
		).Once()

		err := repo.Update(ctx, location, locationID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "location not found")
		mockClient.AssertExpectations(t)
	})
}

func TestDynamoDBRepositoryDelete(t *testing.T) {
	ctx := context.Background()
	mockClient := new(mockDynamoDBClient)
	repo := NewDynamoDBRepository(mockClient, "test-table", "test-gsi")

	accountID := "acc-12345"
	locationID := "loc-001"

	t.Run("Successful delete", func(t *testing.T) {
		mockClient.On("DeleteItem", ctx, mock.MatchedBy(func(input *dynamodb.DeleteItemInput) bool {
			return *input.TableName == "test-table" &&
				input.ConditionExpression != nil &&
				*input.ConditionExpression == "attribute_exists(PK) AND attribute_exists(SK) AND accountId = :accountId" &&
				input.ExpressionAttributeValues != nil &&
				len(input.ExpressionAttributeValues) == 1
		})).Return(&dynamodb.DeleteItemOutput{}, nil).Once()

		err := repo.Delete(ctx, accountID, locationID)
		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("Item not found", func(t *testing.T) {
		mockClient.On("DeleteItem", ctx, mock.Anything).Return(
			nil,
			&types.ConditionalCheckFailedException{Message: aws.String("The conditional request failed")},
		).Once()

		err := repo.Delete(ctx, accountID, locationID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "location not found")
		mockClient.AssertExpectations(t)
	})
}

func TestDynamoDBRepositoryList(t *testing.T) {
	ctx := context.Background()
	mockClient := new(mockDynamoDBClient)
	repo := NewDynamoDBRepository(mockClient, "test-table", "test-gsi")

	accountID := "acc-12345"

	t.Run("Successful list", func(t *testing.T) {
		items := []map[string]types.AttributeValue{
			{
				"PK":           &types.AttributeValueMemberS{Value: "loc-001"},     // PK is the locationID (UUID)
				"SK":           &types.AttributeValueMemberS{Value: "acc-12345"},   // accountID as SK
				"accountId":    &types.AttributeValueMemberS{Value: "acc-12345"},  // accountID as attribute
				"locationType": &types.AttributeValueMemberS{Value: "address"},
				"address": &types.AttributeValueMemberM{
					Value: map[string]types.AttributeValue{
						"streetAddress": &types.AttributeValueMemberS{Value: "123 Main St"},
						"city":          &types.AttributeValueMemberS{Value: "Springfield"},
						"postalCode":    &types.AttributeValueMemberS{Value: "12345"},
						"country":       &types.AttributeValueMemberS{Value: "US"},
					},
				},
			},
			{
				"PK":           &types.AttributeValueMemberS{Value: "loc-002"},     // PK is the locationID (UUID)
				"SK":           &types.AttributeValueMemberS{Value: "acc-12345"},   // accountID as SK
				"accountId":    &types.AttributeValueMemberS{Value: "acc-12345"},  // accountID as attribute
				"locationType": &types.AttributeValueMemberS{Value: "coordinates"},
				"coordinates": &types.AttributeValueMemberM{
					Value: map[string]types.AttributeValue{
						"latitude":  &types.AttributeValueMemberN{Value: "40.7128"},
						"longitude": &types.AttributeValueMemberN{Value: "-74.0060"},
					},
				},
			},
		}

		mockClient.On("Query", ctx, mock.MatchedBy(func(input *dynamodb.QueryInput) bool {
			return *input.IndexName == "test-gsi" &&
				*input.KeyConditionExpression == "accountId = :accountId"
		})).Return(&dynamodb.QueryOutput{Items: items}, nil).Once()

		result, err := repo.List(ctx, accountID, &ListOptions{})
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Len(t, result.Locations, 2)
		assert.IsType(t, models.AddressLocation{}, result.Locations[0])
		assert.IsType(t, models.CoordinatesLocation{}, result.Locations[1])
		assert.Nil(t, result.NextCursor)
		mockClient.AssertExpectations(t)
	})

	t.Run("Empty list", func(t *testing.T) {
		mockClient.On("Query", ctx, mock.Anything).Return(
			&dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{}}, nil,
		).Once()

		result, err := repo.List(ctx, accountID, &ListOptions{})
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Empty(t, result.Locations)
		assert.Nil(t, result.NextCursor)
		mockClient.AssertExpectations(t)
	})
}