package test

import (
	"crypto/rand"
	"reflect"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/store/models"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// Global counter so data can be unique without users having to track that
var counter int = 0

// DataSetup represents configuration for data setup
type DataSetup map[string]any

// DataUtils contains database connection
type DataUtils struct {
	db *gorm.DB
}

// CreateUser creates a new user with data from DataSetup and random values for missing fields
func (du *DataUtils) CreateUser(setup DataSetup) (*models.User, error) {
	user := &models.User{}

	// Get the type and value of the user struct
	userType := reflect.TypeOf(*user)
	userValue := reflect.ValueOf(user).Elem()

	// Iterate through all fields in the user struct
	for i := 0; i < userType.NumField(); i++ {
		field := userType.Field(i)
		fieldName := field.Name

		// Skip fields that should be handled by the database
		if fieldName == "ID" || fieldName == "CreatedAt" || fieldName == "UpdatedAt" {
			continue
		}

		// Check if the field is in the setup
		if val, ok := setup[fieldName]; ok {
			// Set the field from setup
			fieldValue := userValue.FieldByName(fieldName)
			if fieldValue.CanSet() {
				switch fieldName {
				case "Password", "Salt":
					// Handle byte slices
					switch v := val.(type) {
					case string:
						fieldValue.Set(reflect.ValueOf([]byte(v)))
					case []byte:
						fieldValue.Set(reflect.ValueOf(v))
					default:
						// If it's neither string nor []byte, generate random bytes
						bytes := make([]byte, 16)
						rand.Read(bytes)
						fieldValue.Set(reflect.ValueOf(bytes))
					}
				case "Roles":
					// Handle pq.StringArray
					switch v := val.(type) {
					case []string:
						fieldValue.Set(reflect.ValueOf(pq.StringArray(v)))
					case string:
						fieldValue.Set(reflect.ValueOf(pq.StringArray{v}))
					default:
						// Default to user role
						fieldValue.Set(reflect.ValueOf(pq.StringArray{string(models.UserRoleUser)}))
					}
				default:
					// Handle other types
					rv := reflect.ValueOf(val)
					if rv.Type().AssignableTo(fieldValue.Type()) {
						fieldValue.Set(rv)
					}
				}
			}
		} else {
			// Field not in setup, fill with random data
			fieldValue := userValue.FieldByName(fieldName)
			if fieldValue.CanSet() {
				switch fieldName {
				case "Username":
					fieldValue.SetString(gofakeit.Username())
				case "Email":
					fieldValue.SetString(gofakeit.Email())
				case "Password":
					bytes := make([]byte, 16)
					rand.Read(bytes)
					fieldValue.Set(reflect.ValueOf(bytes))
				case "Salt":
					bytes := make([]byte, 16)
					rand.Read(bytes)
					fieldValue.Set(reflect.ValueOf(bytes))
				case "Roles":
					fieldValue.Set(reflect.ValueOf(pq.StringArray{string(models.UserRoleUser)}))
				}
			}
		}
	}

	// Save the user to the database
	err := du.db.Create(user).Error
	return user, err
}

// CreateSession creates a new session with data from DataSetup
// If UserID is not provided in setup, it will create a new user
func (du *DataUtils) CreateSession(setup DataSetup) (*models.Session, error) {
	session := &models.Session{}

	// Check if UserID is provided in setup
	userID, hasUserID := setup["UserID"]
	if !hasUserID || userID == "" {
		// No UserID provided, create a new user
		user, err := du.CreateUser(setup)
		if err != nil {
			return nil, err
		}

		// Use the new user's ID for the session
		setup["UserID"] = user.UserID
	}

	// Get the type and value of the session struct
	sessionType := reflect.TypeOf(*session)
	sessionValue := reflect.ValueOf(session).Elem()

	// Iterate through all fields in the session struct
	for i := 0; i < sessionType.NumField(); i++ {
		field := sessionType.Field(i)
		fieldName := field.Name

		// Skip fields that should be handled by the database
		if fieldName == "SessionID" || fieldName == "CreatedAt" || fieldName == "UpdatedAt" || fieldName == "User" {
			continue
		}

		// Check if the field is in the setup
		if val, ok := setup[fieldName]; ok {
			// Set the field from setup
			fieldValue := sessionValue.FieldByName(fieldName)
			if fieldValue.CanSet() {
				rv := reflect.ValueOf(val)
				if rv.Type().AssignableTo(fieldValue.Type()) {
					fieldValue.Set(rv)
				}
			}
		} else {
			// Field not in setup, fill with random data
			fieldValue := sessionValue.FieldByName(fieldName)
			if fieldValue.CanSet() {
				switch fieldName {
				case "Token":
					fieldValue.SetString(gofakeit.UUID())
				case "Rotated":
					fieldValue.SetBool(false)
				}
			}
		}
	}

	// Save the session to the database
	err := du.db.Create(session).Error
	return session, err
}

// CreateThingType creates a new thing type with data from DataSetup and random values for missing fields
func (du *DataUtils) CreateThingType(setup DataSetup) (*models.ThingType, error) {
	thingType := &models.ThingType{}

	// Get the type and value of the thingType struct
	thingTypeType := reflect.TypeOf(*thingType)
	thingTypeValue := reflect.ValueOf(thingType).Elem()

	// Iterate through all fields in the thingType struct
	for i := 0; i < thingTypeType.NumField(); i++ {
		field := thingTypeType.Field(i)
		fieldName := field.Name

		// Skip fields that should be handled by the database or relationships
		if fieldName == "ThingTypeID" || fieldName == "CreatedAt" || fieldName == "UpdatedAt" || fieldName == "OwnedThings" {
			continue
		}

		// Check if the field is in the setup
		if val, ok := setup[fieldName]; ok {
			// Set the field from setup
			fieldValue := thingTypeValue.FieldByName(fieldName)
			if fieldValue.CanSet() {
				rv := reflect.ValueOf(val)
				if rv.Type().AssignableTo(fieldValue.Type()) {
					fieldValue.Set(rv)
				}
			}
		} else {
			// Field not in setup, fill with random data
			fieldValue := thingTypeValue.FieldByName(fieldName)
			if fieldValue.CanSet() {
				switch fieldName {
				case "Name":
					fieldValue.SetString(gofakeit.ProductName())
				case "Description":
					fieldValue.SetString(gofakeit.ProductDescription())
				}
			}
		}
	}

	// Save the thing type to the database
	err := du.db.Create(thingType).Error
	return thingType, err
}

// CreateOwnedThing creates a new owned thing with data from DataSetup
// If ThingTypeID is not provided in setup, it will create a new thing type
// If OwnerID is not provided in setup, it will create a new user
func (du *DataUtils) CreateOwnedThing(setup DataSetup) (*models.OwnedThing, error) {
	ownedThing := &models.OwnedThing{}

	// Check if ThingTypeID is provided in setup
	thingTypeID, hasThingTypeID := setup["ThingTypeID"]
	if !hasThingTypeID || thingTypeID == "" {
		// No ThingTypeID provided, create a new thing type
		thingType, err := du.CreateThingType(DataSetup{})
		if err != nil {
			return nil, err
		}

		// Use the new thing type's ID for the owned thing
		setup["ThingTypeID"] = thingType.ThingTypeID
	}

	// Check if OwnerID is provided in setup
	ownerID, hasOwnerID := setup["OwnerID"]
	if !hasOwnerID || ownerID == "" {
		// No OwnerID provided, create a new user
		user, err := du.CreateUser(DataSetup{})
		if err != nil {
			return nil, err
		}

		// Use the new user's ID for the owned thing
		setup["OwnerID"] = user.UserID
	}

	// Get the type and value of the ownedThing struct
	ownedThingType := reflect.TypeOf(*ownedThing)
	ownedThingValue := reflect.ValueOf(ownedThing).Elem()

	// Iterate through all fields in the ownedThing struct
	for i := 0; i < ownedThingType.NumField(); i++ {
		field := ownedThingType.Field(i)
		fieldName := field.Name

		// Skip fields that should be handled by the database or relationships
		if fieldName == "OwnedThingID" || fieldName == "CreatedAt" || fieldName == "UpdatedAt" || 
		   fieldName == "ThingType" || fieldName == "Owner" || fieldName == "Borrows" {
			continue
		}

		// Check if the field is in the setup
		if val, ok := setup[fieldName]; ok {
			// Set the field from setup
			fieldValue := ownedThingValue.FieldByName(fieldName)
			if fieldValue.CanSet() {
				rv := reflect.ValueOf(val)
				if rv.Type().AssignableTo(fieldValue.Type()) {
					fieldValue.Set(rv)
				}
			}
		} else {
			// Field not in setup, fill with random data
			fieldValue := ownedThingValue.FieldByName(fieldName)
			if fieldValue.CanSet() {
				switch fieldName {
				case "Name":
					fieldValue.SetString(gofakeit.ProductName())
				}
			}
		}
	}

	// Save the owned thing to the database
	err := du.db.Create(ownedThing).Error
	return ownedThing, err
}

// CreateBorrow creates a new borrow with data from DataSetup
// If BorrowerID is not provided in setup, it will create a new user
// If OwnedThingID is not provided in setup, it will create a new owned thing
func (du *DataUtils) CreateBorrow(setup DataSetup) (*models.Borrow, error) {
	borrow := &models.Borrow{}

	// Check if BorrowerID is provided in setup
	borrowerID, hasBorrowerID := setup["BorrowerID"]
	if !hasBorrowerID || borrowerID == "" {
		// No BorrowerID provided, create a new user
		user, err := du.CreateUser(DataSetup{})
		if err != nil {
			return nil, err
		}

		// Use the new user's ID for the borrow
		setup["BorrowerID"] = user.UserID
	}

	// Check if OwnedThingID is provided in setup
	ownedThingID, hasOwnedThingID := setup["OwnedThingID"]
	if !hasOwnedThingID || ownedThingID == "" {
		// No OwnedThingID provided, create a new owned thing
		ownedThing, err := du.CreateOwnedThing(DataSetup{})
		if err != nil {
			return nil, err
		}

		// Use the new owned thing's ID for the borrow
		setup["OwnedThingID"] = ownedThing.OwnedThingID
	}

	// Get the type and value of the borrow struct
	borrowType := reflect.TypeOf(*borrow)
	borrowValue := reflect.ValueOf(borrow).Elem()

	// Iterate through all fields in the borrow struct
	for i := 0; i < borrowType.NumField(); i++ {
		field := borrowType.Field(i)
		fieldName := field.Name

		// Skip fields that should be handled by the database or relationships
		if fieldName == "BorrowID" || fieldName == "CreatedAt" || fieldName == "UpdatedAt" || 
		   fieldName == "BorrowerRef" || fieldName == "OwnedThing" {
			continue
		}

		// Check if the field is in the setup
		if val, ok := setup[fieldName]; ok {
			// Set the field from setup
			fieldValue := borrowValue.FieldByName(fieldName)
			if fieldValue.CanSet() {
				// For time.Time and *time.Time fields
				if fieldName == "BorrowedAt" || fieldName == "BorrowedUntil" || fieldName == "ReturnedAt" {
					switch v := val.(type) {
					case time.Time:
						fieldValue.Set(reflect.ValueOf(v))
					case string:
						t, err := time.Parse(time.RFC3339, v)
						if err == nil {
							if fieldName == "BorrowedAt" {
								fieldValue.Set(reflect.ValueOf(t))
							} else {
								// For pointer fields (BorrowedUntil, ReturnedAt)
								fieldValue.Set(reflect.ValueOf(&t))
							}
						}
					default:
						// Just use the value if it's already of the right type
						rv := reflect.ValueOf(val)
						if rv.Type().AssignableTo(fieldValue.Type()) {
							fieldValue.Set(rv)
						}
					}
				} else {
					// For non-time fields
					rv := reflect.ValueOf(val)
					if rv.Type().AssignableTo(fieldValue.Type()) {
						fieldValue.Set(rv)
					}
				}
			}
		} else {
			// Field not in setup, fill with random data
			fieldValue := borrowValue.FieldByName(fieldName)
			if fieldValue.CanSet() {
				switch fieldName {
				case "BorrowedAt":
					fieldValue.Set(reflect.ValueOf(time.Now().UTC()))
				case "BorrowedUntil":
					// 50% chance of having a return date
					if gofakeit.Bool() {
						futureTime := time.Now().UTC().Add(time.Duration(gofakeit.Number(1, 30)) * 24 * time.Hour)
						fieldValue.Set(reflect.ValueOf(&futureTime))
					}
				case "ReturnedAt":
					// 25% chance of being already returned
					if gofakeit.Bool() && gofakeit.Bool() {
						returnTime := time.Now().UTC().Add(-time.Duration(gofakeit.Number(1, 10)) * 24 * time.Hour)
						fieldValue.Set(reflect.ValueOf(&returnTime))
					}
				case "Reposessed":
					fieldValue.SetBool(gofakeit.Bool() && gofakeit.Bool() && gofakeit.Bool()) // 12.5% chance
				}
			}
		}
	}

	// Save the borrow to the database
	err := du.db.Create(borrow).Error
	return borrow, err
}
