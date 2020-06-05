package models

//import (
//	"gopkg.in/mgo.v2/bson"
//)

type Product struct {
	Id string
	Title string
	Price int64
}

func NewProduct(id string, title string, price int64) *Product {
	return &Product{id, title, price}
}