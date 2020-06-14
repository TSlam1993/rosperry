package documents

import (
	"time"
)

type UserDocument struct {
	Username string `bson:"_id,omitempty"`
	Email string `bson:"_email,omitempty"`
	Password []byte `bson:"_password,omitempty"`
	BusinessName string `bson:"_businessName"`
	AgeOfBusiness int64 `bson:"_ageOfBusiness"`
	Location string `bson:"_location"`
	CreatedAt time.Time `bson:"_createdAt"`
	UpdatedAt time.Time `bson:"_updatedAt"`
	LastLogged time.Time `bson:"_lastLogged"`
}

type TemplateUserDocument struct {
	Username string
	Email string
	Password []byte
	BusinessName string
	AgeOfBusiness int64
	Location string
	CreatedAt string
	UpdatedAt string
	LastLogged string
	IsUser bool
	Message string
}