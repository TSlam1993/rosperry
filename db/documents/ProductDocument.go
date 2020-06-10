package documents

import (
	"time"
)

type ProductDocument struct {
	Id string  `bson:"_id,omitempty"`
	Title string `bson:"_title,omitempty"`
	Price int64 `bson:"_price,omitempty"`
	Owner string `bson:"_owner,omitempty"`
	Type string `bson:"_type,omitempty"`
	CreatedAt time.Time `bson:"_createdat,omitempty"`
	UpdatedAt time.Time `bson:"_updatedat,omitempty"`
}

type TemplateProductDocument struct {
	Id string
	Title string
	Price int64
	Owner string
	Type string
	CreatedAt string
	UpdatedAt string
	Message string
	IsCurrentUserOwner bool
}