// Package handler provides AppSync event handling for location operations.
package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/steverhoton/location-lambda/internal/models"
	"github.com/steverhoton/location-lambda/internal/repository"
)

// AppSyncEvent represents an event from AWS AppSync.
type AppSyncEvent struct {
	Field     string          `json:"field"`
	Arguments json.RawMessage `json:"arguments"`
	Source    json.RawMessage `json:"source"`
	Identity  AppSyncIdentity `json:"identity"`
	Request   AppSyncRequest  `json:"request"`
}

// AppSyncIdentity represents the identity information from AppSync.
type AppSyncIdentity struct {
	UserArn             string                 `json:"userArn"`
	Username            string                 `json:"username"`
	Claims              map[string]interface{} `json:"claims"`
	SourceIP            []string               `json:"sourceIp"`
	DefaultAuthStrategy string                 `json:"defaultAuthStrategy"`
}

// AppSyncRequest represents request headers from AppSync.
type AppSyncRequest struct {
	Headers map[string]string `json:"headers"`
}

// CreateLocationArguments represents arguments for creating a location.
type CreateLocationArguments struct {
	Input json.RawMessage `json:"input"`
}

// GetLocationArguments represents arguments for getting a location.
type GetLocationArguments struct {
	AccountID  string `json:"accountId"`
	LocationID string `json:"locationId"`
}

// UpdateLocationArguments represents arguments for updating a location.
type UpdateLocationArguments struct {
	LocationID string          `json:"locationId"`
	Input      json.RawMessage `json:"input"`
}

// DeleteLocationArguments represents arguments for deleting a location.
type DeleteLocationArguments struct {
	AccountID  string `json:"accountId"`
	LocationID string `json:"locationId"`
}

// ListLocationsArguments represents arguments for listing locations.
type ListLocationsArguments struct {
	AccountID string  `json:"accountId"`
	Limit     *int32  `json:"limit,omitempty"`
	Cursor    *string `json:"cursor,omitempty"`
}

// LocationResponse wraps a location with metadata.
type LocationResponse struct {
	LocationID string          `json:"locationId"`
	Location   models.Location `json:"location"`
}

// DeleteResponse represents the response for a delete operation.
type DeleteResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ListLocationsResponse represents the response for listing locations with pagination.
type ListLocationsResponse struct {
	Locations  []map[string]interface{} `json:"locations"`
	NextCursor *string                  `json:"nextCursor,omitempty"`
}

// AppSyncHandler handles AppSync events for location operations.
type AppSyncHandler struct {
	repo repository.Repository
}

// NewAppSyncHandler creates a new AppSync handler.
func NewAppSyncHandler(repo repository.Repository) *AppSyncHandler {
	return &AppSyncHandler{
		repo: repo,
	}
}

// Handle processes an AppSync event and returns the appropriate response.
func (h *AppSyncHandler) Handle(ctx context.Context, event AppSyncEvent) (interface{}, error) {
	switch event.Field {
	case "createLocation", "createAddressLocation", "createCoordinatesLocation", "createShopLocation":
		return h.handleCreateLocation(ctx, event.Arguments)
	case "getLocation":
		return h.handleGetLocation(ctx, event.Arguments)
	case "updateLocation", "updateAddressLocation", "updateCoordinatesLocation", "updateShopLocation":
		return h.handleUpdateLocation(ctx, event.Arguments)
	case "deleteLocation":
		return h.handleDeleteLocation(ctx, event.Arguments)
	case "listLocations":
		return h.handleListLocations(ctx, event.Arguments)
	default:
		return nil, fmt.Errorf("unknown field: %s", event.Field)
	}
}

func (h *AppSyncHandler) handleCreateLocation(ctx context.Context, arguments json.RawMessage) (string, error) {
	var args CreateLocationArguments
	if err := json.Unmarshal(arguments, &args); err != nil {
		return "", fmt.Errorf("failed to unmarshal arguments: %w", err)
	}

	location, err := models.UnmarshalLocation(args.Input)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal location: %w", err)
	}

	locationID, err := h.repo.Create(ctx, location)
	if err != nil {
		return "", fmt.Errorf("failed to create location: %w", err)
	}

	return locationID, nil
}

func (h *AppSyncHandler) handleGetLocation(ctx context.Context, arguments json.RawMessage) (map[string]interface{}, error) {
	var args GetLocationArguments
	if err := json.Unmarshal(arguments, &args); err != nil {
		return nil, fmt.Errorf("failed to unmarshal arguments: %w", err)
	}

	location, err := h.repo.Get(ctx, args.AccountID, args.LocationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get location: %w", err)
	}

	// Convert location to map and add __typename
	locationBytes, err := json.Marshal(location)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal location: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(locationBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal location to map: %w", err)
	}

	// Add locationId to the result
	result["locationId"] = args.LocationID

	// Add __typename based on location type
	switch location.GetLocationType() {
	case models.LocationTypeAddress:
		result["__typename"] = "AddressLocation"
	case models.LocationTypeCoordinates:
		result["__typename"] = "CoordinatesLocation"
	case models.LocationTypeShop:
		result["__typename"] = "ShopLocation"
	}

	return result, nil
}

func (h *AppSyncHandler) handleUpdateLocation(ctx context.Context, arguments json.RawMessage) (bool, error) {
	var args UpdateLocationArguments
	if err := json.Unmarshal(arguments, &args); err != nil {
		return false, fmt.Errorf("failed to unmarshal arguments: %w", err)
	}

	location, err := models.UnmarshalLocation(args.Input)
	if err != nil {
		return false, fmt.Errorf("failed to unmarshal location: %w", err)
	}

	if err := h.repo.Update(ctx, location, args.LocationID); err != nil {
		return false, fmt.Errorf("failed to update location: %w", err)
	}

	return true, nil
}

func (h *AppSyncHandler) handleDeleteLocation(ctx context.Context, arguments json.RawMessage) (bool, error) {
	var args DeleteLocationArguments
	if err := json.Unmarshal(arguments, &args); err != nil {
		return false, fmt.Errorf("failed to unmarshal arguments: %w", err)
	}

	if err := h.repo.Delete(ctx, args.AccountID, args.LocationID); err != nil {
		return false, fmt.Errorf("failed to delete location: %w", err)
	}

	return true, nil
}

func (h *AppSyncHandler) handleListLocations(ctx context.Context, arguments json.RawMessage) (*ListLocationsResponse, error) {
	var args ListLocationsArguments
	if err := json.Unmarshal(arguments, &args); err != nil {
		return nil, fmt.Errorf("failed to unmarshal arguments: %w", err)
	}

	options := &repository.ListOptions{
		Limit:  args.Limit,
		Cursor: args.Cursor,
	}

	result, err := h.repo.List(ctx, args.AccountID, options)
	if err != nil {
		return nil, fmt.Errorf("failed to list locations: %w", err)
	}

	// Convert each location to map and add __typename
	locationMaps := make([]map[string]interface{}, len(result.Locations))
	for i, location := range result.Locations {
		// Convert location to map and add __typename
		locationBytes, err := json.Marshal(location)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal location: %w", err)
		}

		var locationMap map[string]interface{}
		if err := json.Unmarshal(locationBytes, &locationMap); err != nil {
			return nil, fmt.Errorf("failed to unmarshal location to map: %w", err)
		}

		// Add locationId to the result
		locationMap["locationId"] = result.LocationIDs[i]

		// Add __typename based on location type
		switch location.GetLocationType() {
		case models.LocationTypeAddress:
			locationMap["__typename"] = "AddressLocation"
		case models.LocationTypeCoordinates:
			locationMap["__typename"] = "CoordinatesLocation"
		case models.LocationTypeShop:
			locationMap["__typename"] = "ShopLocation"
		}

		locationMaps[i] = locationMap
	}

	return &ListLocationsResponse{
		Locations:  locationMaps,
		NextCursor: result.NextCursor,
	}, nil
}
