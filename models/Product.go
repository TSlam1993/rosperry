package models

//import (
//	"gopkg.in/mgo.v2/bson"
//)

type Product struct {
	Id string
	Title string
	Price int64
	Owner string
}

func NewProduct(id string, title string, price int64, owner string) *Product {
	return &Product{id, title, price, owner}
}