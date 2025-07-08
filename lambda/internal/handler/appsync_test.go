package handler

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/steverhoton/location-lambda/internal/models"
	"github.com/steverhoton/location-lambda/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// mockRepository is a mock implementation of the repository.Repository interface.
type mockRepository struct {
	mock.Mock
}

func (m *mockRepository) Create(ctx context.Context, location models.Location) (string, error) {
	args := m.Called(ctx, location)
	return args.String(0), args.Error(1)
}

func (m *mockRepository) Get(ctx context.Context, accountID, locationID string) (models.Location, error) {
	args := m.Called(ctx, accountID, locationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(models.Location), args.Error(1)
}

func (m *mockRepository) Update(ctx context.Context, location models.Location, locationID string) error {
	args := m.Called(ctx, location, locationID)
	return args.Error(0)
}

func (m *mockRepository) Delete(ctx context.Context, accountID, locationID string) error {
	args := m.Called(ctx, accountID, locationID)
	return args.Error(0)
}

func (m *mockRepository) List(ctx context.Context, accountID string, options *repository.ListOptions) (*repository.ListResult, error) {
	args := m.Called(ctx, accountID, options)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.ListResult), args.Error(1)
}

func TestAppSyncHandlerCreateLocation(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mockRepository)
	handler := NewAppSyncHandler(mockRepo)

	addressLocationJSON := `{
		"accountId": "acc-12345",
		"locationType": "address",
		"address": {
			"streetAddress": "123 Main St",
			"city": "Springfield",
			"postalCode": "12345",
			"country": "US"
		}
	}`

	arguments := json.RawMessage(`{"input": ` + addressLocationJSON + `}`)

	event := AppSyncEvent{
		Field:     "createLocation",
		Arguments: arguments,
	}

	t.Run("Successful create", func(t *testing.T) {
		mockRepo.On("Create", ctx, mock.MatchedBy(func(loc models.Location) bool {
			addrLoc, ok := loc.(models.AddressLocation)
			return ok && addrLoc.AccountID == "acc-12345"
		})).Return("test-location-id-123", nil).Once()

		result, err := handler.Handle(ctx, event)
		require.NoError(t, err)

		locationID, ok := result.(string)
		require.True(t, ok)
		assert.NotEmpty(t, locationID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Invalid location data", func(t *testing.T) {
		invalidArguments := json.RawMessage(`{"input": {"invalid": "data"}}`)
		invalidEvent := AppSyncEvent{
			Field:     "createLocation",
			Arguments: invalidArguments,
		}

		result, err := handler.Handle(ctx, invalidEvent)
		assert.Error(t, err)
		assert.Equal(t, "", result)
		assert.Contains(t, err.Error(), "failed to unmarshal location")
	})

	t.Run("Repository error", func(t *testing.T) {
		mockRepo.On("Create", ctx, mock.Anything).Return("", errors.New("database error")).Once()

		result, err := handler.Handle(ctx, event)
		assert.Error(t, err)
		assert.Equal(t, "", result)
		assert.Contains(t, err.Error(), "failed to create location")
		mockRepo.AssertExpectations(t)
	})
}

func TestAppSyncHandlerGetLocation(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mockRepository)
	handler := NewAppSyncHandler(mockRepo)

	arguments := json.RawMessage(`{"accountId": "acc-12345", "locationId": "loc-001"}`)
	event := AppSyncEvent{
		Field:     "getLocation",
		Arguments: arguments,
	}

	expectedLocation := models.AddressLocation{
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

	t.Run("Successful get", func(t *testing.T) {
		mockRepo.On("Get", ctx, "acc-12345", "loc-001").Return(expectedLocation, nil).Once()

		result, err := handler.Handle(ctx, event)
		require.NoError(t, err)

		locationMap, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "acc-12345", locationMap["accountId"])
		assert.Equal(t, "loc-001", locationMap["locationId"])
		assert.Equal(t, "AddressLocation", locationMap["__typename"])
		mockRepo.AssertExpectations(t)
	})

	t.Run("Location not found", func(t *testing.T) {
		mockRepo.On("Get", ctx, "acc-12345", "loc-001").Return(nil, errors.New("location not found")).Once()

		result, err := handler.Handle(ctx, event)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to get location")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Invalid arguments", func(t *testing.T) {
		invalidArguments := json.RawMessage(`{"invalid": "arguments"}`)
		invalidEvent := AppSyncEvent{
			Field:     "getLocation",
			Arguments: invalidArguments,
		}

		// The handler will try to call Get with empty strings due to missing fields
		// This is expected behavior - the arguments unmarshal to zero values
		mockRepo.On("Get", ctx, "", "").Return(nil, errors.New("location not found")).Once()

		result, err := handler.Handle(ctx, invalidEvent)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to get location")
		mockRepo.AssertExpectations(t)
	})
}

func TestAppSyncHandlerUpdateLocation(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mockRepository)
	handler := NewAppSyncHandler(mockRepo)

	updatedLocationJSON := `{
		"accountId": "acc-12345",
		"locationType": "address",
		"address": {
			"streetAddress": "456 Oak Ave",
			"city": "Springfield",
			"postalCode": "12345",
			"country": "US"
		}
	}`

	arguments := json.RawMessage(`{"locationId": "loc-001", "input": ` + updatedLocationJSON + `}`)
	event := AppSyncEvent{
		Field:     "updateLocation",
		Arguments: arguments,
	}

	t.Run("Successful update", func(t *testing.T) {
		mockRepo.On("Update", ctx, mock.MatchedBy(func(loc models.Location) bool {
			addrLoc, ok := loc.(models.AddressLocation)
			return ok && addrLoc.Address.StreetAddress == "456 Oak Ave"
		}), "loc-001").Return(nil).Once()

		result, err := handler.Handle(ctx, event)
		require.NoError(t, err)

		success, ok := result.(bool)
		require.True(t, ok)
		assert.True(t, success)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Update non-existent location", func(t *testing.T) {
		mockRepo.On("Update", ctx, mock.Anything, "loc-001").Return(errors.New("location not found")).Once()

		result, err := handler.Handle(ctx, event)
		assert.Error(t, err)
		assert.Equal(t, false, result)
		assert.Contains(t, err.Error(), "failed to update location")
		mockRepo.AssertExpectations(t)
	})
}

func TestAppSyncHandlerDeleteLocation(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mockRepository)
	handler := NewAppSyncHandler(mockRepo)

	arguments := json.RawMessage(`{"accountId": "acc-12345", "locationId": "loc-001"}`)
	event := AppSyncEvent{
		Field:     "deleteLocation",
		Arguments: arguments,
	}

	t.Run("Successful delete", func(t *testing.T) {
		mockRepo.On("Delete", ctx, "acc-12345", "loc-001").Return(nil).Once()

		result, err := handler.Handle(ctx, event)
		require.NoError(t, err)

		success, ok := result.(bool)
		require.True(t, ok)
		assert.True(t, success)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Delete non-existent location", func(t *testing.T) {
		mockRepo.On("Delete", ctx, "acc-12345", "loc-001").Return(errors.New("location not found")).Once()

		result, err := handler.Handle(ctx, event)
		assert.Error(t, err)
		assert.Equal(t, false, result)
		assert.Contains(t, err.Error(), "failed to delete location")
		mockRepo.AssertExpectations(t)
	})
}

func TestAppSyncHandlerListLocations(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mockRepository)
	handler := NewAppSyncHandler(mockRepo)

	arguments := json.RawMessage(`{"accountId": "acc-12345"}`)
	event := AppSyncEvent{
		Field:     "listLocations",
		Arguments: arguments,
	}

	expectedLocations := []models.Location{
		models.AddressLocation{
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
		},
		models.CoordinatesLocation{
			LocationBase: models.LocationBase{
				AccountID:    "acc-12345",
				LocationType: models.LocationTypeCoordinates,
			},
			Coordinates: models.Coordinates{
				Latitude:  40.7128,
				Longitude: -74.0060,
			},
		},
	}

	t.Run("Successful list", func(t *testing.T) {
		expectedResult := &repository.ListResult{
			Locations:   expectedLocations,
			LocationIDs: []string{"loc-123", "loc-456"},
			NextCursor:  nil,
		}
		mockRepo.On("List", ctx, "acc-12345", mock.AnythingOfType("*repository.ListOptions")).Return(expectedResult, nil).Once()

		result, err := handler.Handle(ctx, event)
		require.NoError(t, err)

		response, ok := result.(*ListLocationsResponse)
		require.True(t, ok)
		assert.Len(t, response.Locations, 2)
		assert.Nil(t, response.NextCursor)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Empty list", func(t *testing.T) {
		expectedResult := &repository.ListResult{
			Locations:   []models.Location{},
			LocationIDs: []string{},
			NextCursor:  nil,
		}
		mockRepo.On("List", ctx, "acc-12345", mock.AnythingOfType("*repository.ListOptions")).Return(expectedResult, nil).Once()

		result, err := handler.Handle(ctx, event)
		require.NoError(t, err)

		response, ok := result.(*ListLocationsResponse)
		require.True(t, ok)
		assert.Empty(t, response.Locations)
		assert.Nil(t, response.NextCursor)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Repository error", func(t *testing.T) {
		mockRepo.On("List", ctx, "acc-12345", mock.AnythingOfType("*repository.ListOptions")).Return(nil, errors.New("database error")).Once()

		result, err := handler.Handle(ctx, event)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to list locations")
		mockRepo.AssertExpectations(t)
	})
}

func TestAppSyncHandlerUnknownField(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mockRepository)
	handler := NewAppSyncHandler(mockRepo)

	event := AppSyncEvent{
		Field:     "unknownOperation",
		Arguments: json.RawMessage(`{}`),
	}

	result, err := handler.Handle(ctx, event)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unknown field: unknownOperation")
}
