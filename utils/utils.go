package utils

import (
	"crypto/rand"
	"fmt"
	"gopkg.in/mgo.v2"
	"rosperry/db/documents"
)

func GenerateId() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func DropCollection(collection *mgo.Collection) {
	collection.RemoveAll(nil)
}

func PrintCollection(collection *mgo.Collection) {
	collectionDocuments := []documents.UserDocument{}
	collection.Find(nil).All(&collectionDocuments)

	for _, doc := range collectionDocuments {
		fmt.Println(doc)
	}
}