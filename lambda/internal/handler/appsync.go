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
	UserArn    string                 `json:"userArn"`
	Username   string                 `json:"username"`
	Claims     map[string]interface{} `json:"claims"`
	SourceIP   []string               `json:"sourceIp"`
	DefaultAuthStrategy string        `json:"defaultAuthStrategy"`
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
	Locations  []models.Location `json:"locations"`
	NextCursor *string           `json:"nextCursor,omitempty"`
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
	case "createLocation":
		return h.handleCreateLocation(ctx, event.Arguments)
	case "getLocation":
		return h.handleGetLocation(ctx, event.Arguments)
	case "updateLocation":
		return h.handleUpdateLocation(ctx, event.Arguments)
	case "deleteLocation":
		return h.handleDeleteLocation(ctx, event.Arguments)
	case "listLocations":
		return h.handleListLocations(ctx, event.Arguments)
	default:
		return nil, fmt.Errorf("unknown field: %s", event.Field)
	}
}

func (h *AppSyncHandler) handleCreateLocation(ctx context.Context, arguments json.RawMessage) (*LocationResponse, error) {
	var args CreateLocationArguments
	if err := json.Unmarshal(arguments, &args); err != nil {
		return nil, fmt.Errorf("failed to unmarshal arguments: %w", err)
	}

	location, err := models.UnmarshalLocation(args.Input)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal location: %w", err)
	}

	locationID, err := h.repo.Create(ctx, location)
	if err != nil {
		return nil, fmt.Errorf("failed to create location: %w", err)
	}

	return &LocationResponse{
		LocationID: locationID,
		Location:   location,
	}, nil
}

func (h *AppSyncHandler) handleGetLocation(ctx context.Context, arguments json.RawMessage) (models.Location, error) {
	var args GetLocationArguments
	if err := json.Unmarshal(arguments, &args); err != nil {
		return nil, fmt.Errorf("failed to unmarshal arguments: %w", err)
	}

	location, err := h.repo.Get(ctx, args.AccountID, args.LocationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get location: %w", err)
	}

	return location, nil
}

func (h *AppSyncHandler) handleUpdateLocation(ctx context.Context, arguments json.RawMessage) (models.Location, error) {
	var args UpdateLocationArguments
	if err := json.Unmarshal(arguments, &args); err != nil {
		return nil, fmt.Errorf("failed to unmarshal arguments: %w", err)
	}

	location, err := models.UnmarshalLocation(args.Input)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal location: %w", err)
	}

	if err := h.repo.Update(ctx, location, args.LocationID); err != nil {
		return nil, fmt.Errorf("failed to update location: %w", err)
	}

	return location, nil
}

func (h *AppSyncHandler) handleDeleteLocation(ctx context.Context, arguments json.RawMessage) (*DeleteResponse, error) {
	var args DeleteLocationArguments
	if err := json.Unmarshal(arguments, &args); err != nil {
		return nil, fmt.Errorf("failed to unmarshal arguments: %w", err)
	}

	if err := h.repo.Delete(ctx, args.AccountID, args.LocationID); err != nil {
		return &DeleteResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to delete location: %v", err),
		}, nil
	}

	return &DeleteResponse{
		Success: true,
		Message: "Location deleted successfully",
	}, nil
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

	return &ListLocationsResponse{
		Locations:  result.Locations,
		NextCursor: result.NextCursor,
	}, nil
}

