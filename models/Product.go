package models

import (
//	"gopkg.in/mgo.v2/bson"
	"time"
)

type Product struct {
	Id string
	Title string
	Price int64
	Owner string
	Type string
	CreatedAt time.Time
	UpdatedAt time.Time
}

//func NewProduct(id string, title string, price int64, owner string) *Product {
//	return &Product{id, title, price, owner}
//}