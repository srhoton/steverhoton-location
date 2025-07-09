// Package models contains the data structures for location records.
package models

import (
	"encoding/json"
	"errors"
	"fmt"
)

// LocationType represents the type of location.
type LocationType string

const (
	// LocationTypeAddress represents a location specified by mailing address.
	LocationTypeAddress LocationType = "address"
	// LocationTypeCoordinates represents a location specified by GPS coordinates.
	LocationTypeCoordinates LocationType = "coordinates"
	// LocationTypeShop represents a shop location with business details.
	LocationTypeShop LocationType = "shop"
)

// Location is the base interface for all location types.
type Location interface {
	GetAccountID() string
	GetLocationType() LocationType
	GetExtendedAttributes() map[string]interface{}
	Validate() error
}

// LocationBase contains common fields for all location types.
type LocationBase struct {
	AccountID          string                 `json:"accountId" dynamodbav:"accountId"`
	LocationType       LocationType           `json:"locationType" dynamodbav:"locationType"`
	ExtendedAttributes map[string]interface{} `json:"extendedAttributes,omitempty" dynamodbav:"extendedAttributes,omitempty"`
}

// GetAccountID returns the account ID.
func (l LocationBase) GetAccountID() string {
	return l.AccountID
}

// GetLocationType returns the location type.
func (l LocationBase) GetLocationType() LocationType {
	return l.LocationType
}

// GetExtendedAttributes returns the extended attributes.
func (l LocationBase) GetExtendedAttributes() map[string]interface{} {
	return l.ExtendedAttributes
}

// Address represents a mailing address.
type Address struct {
	StreetAddress  string `json:"streetAddress" dynamodbav:"streetAddress"`
	StreetAddress2 string `json:"streetAddress2,omitempty" dynamodbav:"streetAddress2,omitempty"`
	City           string `json:"city" dynamodbav:"city"`
	StateProvince  string `json:"stateProvince,omitempty" dynamodbav:"stateProvince,omitempty"`
	PostalCode     string `json:"postalCode" dynamodbav:"postalCode"`
	Country        string `json:"country" dynamodbav:"country"`
}

// Validate validates the address fields.
func (a Address) Validate() error {
	if a.StreetAddress == "" {
		return errors.New("streetAddress is required")
	}
	if a.City == "" {
		return errors.New("city is required")
	}
	if a.PostalCode == "" {
		return errors.New("postalCode is required")
	}
	if a.Country == "" {
		return errors.New("country is required")
	}
	if len(a.Country) != 2 {
		return errors.New("country must be a 2-character ISO 3166-1 alpha-2 code")
	}
	return nil
}

// AddressLocation represents a location specified by mailing address.
type AddressLocation struct {
	LocationBase
	Address Address `json:"address" dynamodbav:"address"`
}

// Validate validates the address location.
func (l AddressLocation) Validate() error {
	if l.AccountID == "" {
		return errors.New("accountId is required")
	}
	if l.LocationType != LocationTypeAddress {
		return fmt.Errorf("invalid locationType for AddressLocation: %s", l.LocationType)
	}
	return l.Address.Validate()
}

// Coordinates represents GPS coordinates.
type Coordinates struct {
	Latitude  float64  `json:"latitude" dynamodbav:"latitude"`
	Longitude float64  `json:"longitude" dynamodbav:"longitude"`
	Altitude  *float64 `json:"altitude,omitempty" dynamodbav:"altitude,omitempty"`
	Accuracy  *float64 `json:"accuracy,omitempty" dynamodbav:"accuracy,omitempty"`
}

// Validate validates the coordinates.
func (c Coordinates) Validate() error {
	if c.Latitude < -90 || c.Latitude > 90 {
		return fmt.Errorf("latitude must be between -90 and 90, got %f", c.Latitude)
	}
	if c.Longitude < -180 || c.Longitude > 180 {
		return fmt.Errorf("longitude must be between -180 and 180, got %f", c.Longitude)
	}
	if c.Accuracy != nil && *c.Accuracy < 0 {
		return fmt.Errorf("accuracy must be non-negative, got %f", *c.Accuracy)
	}
	return nil
}

// CoordinatesLocation represents a location specified by GPS coordinates.
type CoordinatesLocation struct {
	LocationBase
	Coordinates Coordinates `json:"coordinates" dynamodbav:"coordinates"`
}

// Validate validates the coordinates location.
func (l CoordinatesLocation) Validate() error {
	if l.AccountID == "" {
		return errors.New("accountId is required")
	}
	if l.LocationType != LocationTypeCoordinates {
		return fmt.Errorf("invalid locationType for CoordinatesLocation: %s", l.LocationType)
	}
	return l.Coordinates.Validate()
}

// Shop represents a shop or business location with address information.
type Shop struct {
	Name      string  `json:"name" dynamodbav:"name"`
	ContactID string  `json:"contactId" dynamodbav:"contactId"`
	Address   Address `json:"address" dynamodbav:"address"`
}

// Validate validates the shop fields.
func (s Shop) Validate() error {
	if s.Name == "" {
		return errors.New("name is required")
	}
	if s.ContactID == "" {
		return errors.New("contactId is required")
	}
	if err := s.Address.Validate(); err != nil {
		return err
	}
	return nil
}
// ShopLocation represents a shop location with business details.
type ShopLocation struct {
	LocationBase
	Shop Shop `json:"shop" dynamodbav:"shop"`
}

// Validate validates the shop location.
func (l ShopLocation) Validate() error {
	if l.AccountID == "" {
		return errors.New("accountId is required")
	}
	if l.LocationType != LocationTypeShop {
		return fmt.Errorf("invalid locationType for ShopLocation: %s", l.LocationType)
	}
	return l.Shop.Validate()
}

// UnmarshalLocation unmarshals a JSON byte slice into the appropriate Location type.
func UnmarshalLocation(data []byte) (Location, error) {
	var base struct {
		LocationType LocationType `json:"locationType"`
	}

	if err := json.Unmarshal(data, &base); err != nil {
		return nil, fmt.Errorf("failed to unmarshal location type: %w", err)
	}

	switch base.LocationType {
	case LocationTypeAddress:
		var loc AddressLocation
		if err := json.Unmarshal(data, &loc); err != nil {
			return nil, fmt.Errorf("failed to unmarshal address location: %w", err)
		}
		return loc, nil
	case LocationTypeCoordinates:
		var loc CoordinatesLocation
		if err := json.Unmarshal(data, &loc); err != nil {
			return nil, fmt.Errorf("failed to unmarshal coordinates location: %w", err)
		}
		return loc, nil
	case LocationTypeShop:
		var loc ShopLocation
		if err := json.Unmarshal(data, &loc); err != nil {
			return nil, fmt.Errorf("failed to unmarshal shop location: %w", err)
		}
		return loc, nil
	default:
		return nil, fmt.Errorf("unknown location type: %s", base.LocationType)
	}
}

// LocationWrapper is used for unmarshaling locations from DynamoDB.
type LocationWrapper struct {
	Location
}

// UnmarshalJSON implements custom JSON unmarshaling for LocationWrapper.
func (w *LocationWrapper) UnmarshalJSON(data []byte) error {
	loc, err := UnmarshalLocation(data)
	if err != nil {
		return err
	}
	w.Location = loc
	return nil
}
