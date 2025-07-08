package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddressValidation(t *testing.T) {
	tests := []struct {
		name    string
		address Address
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid address",
			address: Address{
				StreetAddress: "123 Main St",
				City:          "Springfield",
				PostalCode:    "12345",
				Country:       "US",
			},
			wantErr: false,
		},
		{
			name: "Valid address with optional fields",
			address: Address{
				StreetAddress:  "123 Main St",
				StreetAddress2: "Apt 4B",
				City:           "Springfield",
				StateProvince:  "IL",
				PostalCode:     "12345",
				Country:        "US",
			},
			wantErr: false,
		},
		{
			name: "Missing street address",
			address: Address{
				City:       "Springfield",
				PostalCode: "12345",
				Country:    "US",
			},
			wantErr: true,
			errMsg:  "streetAddress is required",
		},
		{
			name: "Missing city",
			address: Address{
				StreetAddress: "123 Main St",
				PostalCode:    "12345",
				Country:       "US",
			},
			wantErr: true,
			errMsg:  "city is required",
		},
		{
			name: "Missing postal code",
			address: Address{
				StreetAddress: "123 Main St",
				City:          "Springfield",
				Country:       "US",
			},
			wantErr: true,
			errMsg:  "postalCode is required",
		},
		{
			name: "Missing country",
			address: Address{
				StreetAddress: "123 Main St",
				City:          "Springfield",
				PostalCode:    "12345",
			},
			wantErr: true,
			errMsg:  "country is required",
		},
		{
			name: "Invalid country code",
			address: Address{
				StreetAddress: "123 Main St",
				City:          "Springfield",
				PostalCode:    "12345",
				Country:       "USA",
			},
			wantErr: true,
			errMsg:  "country must be a 2-character ISO 3166-1 alpha-2 code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.address.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCoordinatesValidation(t *testing.T) {
	tests := []struct {
		name        string
		coordinates Coordinates
		wantErr     bool
		errMsg      string
	}{
		{
			name: "Valid coordinates",
			coordinates: Coordinates{
				Latitude:  40.7128,
				Longitude: -74.0060,
			},
			wantErr: false,
		},
		{
			name: "Valid coordinates with optional fields",
			coordinates: Coordinates{
				Latitude:  40.7128,
				Longitude: -74.0060,
				Altitude:  floatPtr(100.5),
				Accuracy:  floatPtr(5.0),
			},
			wantErr: false,
		},
		{
			name: "Invalid latitude too low",
			coordinates: Coordinates{
				Latitude:  -91.0,
				Longitude: -74.0060,
			},
			wantErr: true,
			errMsg:  "latitude must be between -90 and 90",
		},
		{
			name: "Invalid latitude too high",
			coordinates: Coordinates{
				Latitude:  91.0,
				Longitude: -74.0060,
			},
			wantErr: true,
			errMsg:  "latitude must be between -90 and 90",
		},
		{
			name: "Invalid longitude too low",
			coordinates: Coordinates{
				Latitude:  40.7128,
				Longitude: -181.0,
			},
			wantErr: true,
			errMsg:  "longitude must be between -180 and 180",
		},
		{
			name: "Invalid longitude too high",
			coordinates: Coordinates{
				Latitude:  40.7128,
				Longitude: 181.0,
			},
			wantErr: true,
			errMsg:  "longitude must be between -180 and 180",
		},
		{
			name: "Invalid negative accuracy",
			coordinates: Coordinates{
				Latitude:  40.7128,
				Longitude: -74.0060,
				Accuracy:  floatPtr(-1.0),
			},
			wantErr: true,
			errMsg:  "accuracy must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.coordinates.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAddressLocationValidation(t *testing.T) {
	tests := []struct {
		name     string
		location AddressLocation
		wantErr  bool
		errMsg   string
	}{
		{
			name: "Valid address location",
			location: AddressLocation{
				LocationBase: LocationBase{
					AccountID:    "acc-12345",
					LocationType: LocationTypeAddress,
					ExtendedAttributes: map[string]interface{}{
						"businessName": "Acme Corp",
					},
				},
				Address: Address{
					StreetAddress: "123 Main St",
					City:          "Springfield",
					PostalCode:    "12345",
					Country:       "US",
				},
			},
			wantErr: false,
		},
		{
			name: "Missing account ID",
			location: AddressLocation{
				LocationBase: LocationBase{
					LocationType: LocationTypeAddress,
				},
				Address: Address{
					StreetAddress: "123 Main St",
					City:          "Springfield",
					PostalCode:    "12345",
					Country:       "US",
				},
			},
			wantErr: true,
			errMsg:  "accountId is required",
		},
		{
			name: "Wrong location type",
			location: AddressLocation{
				LocationBase: LocationBase{
					AccountID:    "acc-12345",
					LocationType: LocationTypeCoordinates,
				},
				Address: Address{
					StreetAddress: "123 Main St",
					City:          "Springfield",
					PostalCode:    "12345",
					Country:       "US",
				},
			},
			wantErr: true,
			errMsg:  "invalid locationType for AddressLocation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.location.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCoordinatesLocationValidation(t *testing.T) {
	tests := []struct {
		name     string
		location CoordinatesLocation
		wantErr  bool
		errMsg   string
	}{
		{
			name: "Valid coordinates location",
			location: CoordinatesLocation{
				LocationBase: LocationBase{
					AccountID:    "acc-12345",
					LocationType: LocationTypeCoordinates,
					ExtendedAttributes: map[string]interface{}{
						"sensorType": "weather",
					},
				},
				Coordinates: Coordinates{
					Latitude:  40.7128,
					Longitude: -74.0060,
				},
			},
			wantErr: false,
		},
		{
			name: "Missing account ID",
			location: CoordinatesLocation{
				LocationBase: LocationBase{
					LocationType: LocationTypeCoordinates,
				},
				Coordinates: Coordinates{
					Latitude:  40.7128,
					Longitude: -74.0060,
				},
			},
			wantErr: true,
			errMsg:  "accountId is required",
		},
		{
			name: "Wrong location type",
			location: CoordinatesLocation{
				LocationBase: LocationBase{
					AccountID:    "acc-12345",
					LocationType: LocationTypeAddress,
				},
				Coordinates: Coordinates{
					Latitude:  40.7128,
					Longitude: -74.0060,
				},
			},
			wantErr: true,
			errMsg:  "invalid locationType for CoordinatesLocation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.location.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestShopValidation(t *testing.T) {
	tests := []struct {
		name    string
		shop    Shop
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid shop",
			shop: Shop{
				Name:          "Coffee Shop",
				ContactID:     "contact-123e4567-e89b-12d3-a456-426614174000",
				StreetAddress: "123 Main St",
				City:          "Springfield",
				PostalCode:    "12345",
				Country:       "US",
			},
			wantErr: false,
		},
		{
			name: "Valid shop with optional fields",
			shop: Shop{
				Name:           "Coffee Shop",
				ContactID:      "contact-123e4567-e89b-12d3-a456-426614174000",
				StreetAddress:  "123 Main St",
				StreetAddress2: "Suite 100",
				City:           "Springfield",
				StateProvince:  "IL",
				PostalCode:     "12345",
				Country:        "US",
			},
			wantErr: false,
		},
		{
			name: "Missing name",
			shop: Shop{
				ContactID:     "contact-123e4567-e89b-12d3-a456-426614174000",
				StreetAddress: "123 Main St",
				City:          "Springfield",
				PostalCode:    "12345",
				Country:       "US",
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "Missing contactId",
			shop: Shop{
				Name:          "Coffee Shop",
				StreetAddress: "123 Main St",
				City:          "Springfield",
				PostalCode:    "12345",
				Country:       "US",
			},
			wantErr: true,
			errMsg:  "contactId is required",
		},
		{
			name: "Missing street address",
			shop: Shop{
				Name:       "Coffee Shop",
				ContactID:  "contact-123e4567-e89b-12d3-a456-426614174000",
				City:       "Springfield",
				PostalCode: "12345",
				Country:    "US",
			},
			wantErr: true,
			errMsg:  "streetAddress is required",
		},
		{
			name: "Missing city",
			shop: Shop{
				Name:          "Coffee Shop",
				ContactID:     "contact-123e4567-e89b-12d3-a456-426614174000",
				StreetAddress: "123 Main St",
				PostalCode:    "12345",
				Country:       "US",
			},
			wantErr: true,
			errMsg:  "city is required",
		},
		{
			name: "Missing postal code",
			shop: Shop{
				Name:          "Coffee Shop",
				ContactID:     "contact-123e4567-e89b-12d3-a456-426614174000",
				StreetAddress: "123 Main St",
				City:          "Springfield",
				Country:       "US",
			},
			wantErr: true,
			errMsg:  "postalCode is required",
		},
		{
			name: "Missing country",
			shop: Shop{
				Name:          "Coffee Shop",
				ContactID:     "contact-123e4567-e89b-12d3-a456-426614174000",
				StreetAddress: "123 Main St",
				City:          "Springfield",
				PostalCode:    "12345",
			},
			wantErr: true,
			errMsg:  "country is required",
		},
		{
			name: "Invalid country code",
			shop: Shop{
				Name:          "Coffee Shop",
				ContactID:     "contact-123e4567-e89b-12d3-a456-426614174000",
				StreetAddress: "123 Main St",
				City:          "Springfield",
				PostalCode:    "12345",
				Country:       "USA",
			},
			wantErr: true,
			errMsg:  "country must be a 2-character ISO 3166-1 alpha-2 code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.shop.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestShopLocationValidation(t *testing.T) {
	tests := []struct {
		name     string
		location ShopLocation
		wantErr  bool
		errMsg   string
	}{
		{
			name: "Valid shop location",
			location: ShopLocation{
				LocationBase: LocationBase{
					AccountID:    "acc-12345",
					LocationType: LocationTypeShop,
					ExtendedAttributes: map[string]interface{}{
						"verified": true,
					},
				},
				Shop: Shop{
					Name:          "Coffee Shop",
					ContactID:     "contact-123e4567-e89b-12d3-a456-426614174000",
					StreetAddress: "123 Main St",
					City:          "Springfield",
					PostalCode:    "12345",
					Country:       "US",
				},
			},
			wantErr: false,
		},
		{
			name: "Missing account ID",
			location: ShopLocation{
				LocationBase: LocationBase{
					LocationType: LocationTypeShop,
				},
				Shop: Shop{
					Name:          "Coffee Shop",
					ContactID:     "contact-123e4567-e89b-12d3-a456-426614174000",
					StreetAddress: "123 Main St",
					City:          "Springfield",
					PostalCode:    "12345",
					Country:       "US",
				},
			},
			wantErr: true,
			errMsg:  "accountId is required",
		},
		{
			name: "Wrong location type",
			location: ShopLocation{
				LocationBase: LocationBase{
					AccountID:    "acc-12345",
					LocationType: LocationTypeAddress,
				},
				Shop: Shop{
					Name:          "Coffee Shop",
					ContactID:     "contact-123e4567-e89b-12d3-a456-426614174000",
					StreetAddress: "123 Main St",
					City:          "Springfield",
					PostalCode:    "12345",
					Country:       "US",
				},
			},
			wantErr: true,
			errMsg:  "invalid locationType for ShopLocation",
		},
		{
			name: "Invalid shop",
			location: ShopLocation{
				LocationBase: LocationBase{
					AccountID:    "acc-12345",
					LocationType: LocationTypeShop,
				},
				Shop: Shop{
					ContactID:     "contact-123e4567-e89b-12d3-a456-426614174000",
					StreetAddress: "123 Main St",
					City:          "Springfield",
					PostalCode:    "12345",
					Country:       "US",
				},
			},
			wantErr: true,
			errMsg:  "name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.location.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUnmarshalLocation(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
		check   func(t *testing.T, loc Location)
	}{
		{
			name: "Valid address location",
			json: `{
				"accountId": "acc-12345",
				"locationType": "address",
				"address": {
					"streetAddress": "123 Main St",
					"city": "Springfield",
					"postalCode": "12345",
					"country": "US"
				},
				"extendedAttributes": {
					"businessName": "Acme Corp"
				}
			}`,
			wantErr: false,
			check: func(t *testing.T, loc Location) {
				assert.IsType(t, AddressLocation{}, loc)
				addrLoc := loc.(AddressLocation)
				assert.Equal(t, "acc-12345", addrLoc.AccountID)
				assert.Equal(t, LocationTypeAddress, addrLoc.LocationType)
				assert.Equal(t, "123 Main St", addrLoc.Address.StreetAddress)
				assert.Equal(t, "Acme Corp", addrLoc.ExtendedAttributes["businessName"])
			},
		},
		{
			name: "Valid coordinates location",
			json: `{
				"accountId": "acc-67890",
				"locationType": "coordinates",
				"coordinates": {
					"latitude": 40.7128,
					"longitude": -74.0060,
					"accuracy": 5.0
				},
				"extendedAttributes": {
					"sensorType": "weather"
				}
			}`,
			wantErr: false,
			check: func(t *testing.T, loc Location) {
				assert.IsType(t, CoordinatesLocation{}, loc)
				coordLoc := loc.(CoordinatesLocation)
				assert.Equal(t, "acc-67890", coordLoc.AccountID)
				assert.Equal(t, LocationTypeCoordinates, coordLoc.LocationType)
				assert.Equal(t, 40.7128, coordLoc.Coordinates.Latitude)
				assert.Equal(t, -74.0060, coordLoc.Coordinates.Longitude)
				assert.Equal(t, 5.0, *coordLoc.Coordinates.Accuracy)
				assert.Equal(t, "weather", coordLoc.ExtendedAttributes["sensorType"])
			},
		},
		{
			name: "Valid shop location",
			json: `{
				"accountId": "acc-54321",
				"locationType": "shop",
				"shop": {
					"name": "Coffee Shop",
					"contactId": "contact-123e4567-e89b-12d3-a456-426614174000",
					"streetAddress": "123 Main St",
					"streetAddress2": "Suite 100",
					"city": "Springfield",
					"stateProvince": "IL",
					"postalCode": "12345",
					"country": "US"
				},
				"extendedAttributes": {
					"verified": true
				}
			}`,
			wantErr: false,
			check: func(t *testing.T, loc Location) {
				assert.IsType(t, ShopLocation{}, loc)
				shopLoc := loc.(ShopLocation)
				assert.Equal(t, "acc-54321", shopLoc.AccountID)
				assert.Equal(t, LocationTypeShop, shopLoc.LocationType)
				assert.Equal(t, "Coffee Shop", shopLoc.Shop.Name)
				assert.Equal(t, "contact-123e4567-e89b-12d3-a456-426614174000", shopLoc.Shop.ContactID)
				assert.Equal(t, "123 Main St", shopLoc.Shop.StreetAddress)
				assert.Equal(t, "Suite 100", shopLoc.Shop.StreetAddress2)
				assert.Equal(t, "Springfield", shopLoc.Shop.City)
				assert.Equal(t, "IL", shopLoc.Shop.StateProvince)
				assert.Equal(t, "12345", shopLoc.Shop.PostalCode)
				assert.Equal(t, "US", shopLoc.Shop.Country)
				assert.Equal(t, true, shopLoc.ExtendedAttributes["verified"])
			},
		},
		{
			name: "Unknown location type",
			json: `{
				"accountId": "acc-12345",
				"locationType": "unknown"
			}`,
			wantErr: true,
		},
		{
			name:    "Invalid JSON",
			json:    `{invalid json}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc, err := UnmarshalLocation([]byte(tt.json))
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

func TestLocationWrapperUnmarshalJSON(t *testing.T) {
	addressJSON := `{
		"accountId": "acc-12345",
		"locationType": "address",
		"address": {
			"streetAddress": "123 Main St",
			"city": "Springfield",
			"postalCode": "12345",
			"country": "US"
		}
	}`

	var wrapper LocationWrapper
	err := json.Unmarshal([]byte(addressJSON), &wrapper)
	require.NoError(t, err)
	
	assert.NotNil(t, wrapper.Location)
	assert.IsType(t, AddressLocation{}, wrapper.Location)
	assert.Equal(t, "acc-12345", wrapper.Location.GetAccountID())
	assert.Equal(t, LocationTypeAddress, wrapper.Location.GetLocationType())
}

func floatPtr(f float64) *float64 {
	return &f
}