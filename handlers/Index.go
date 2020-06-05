package handlers

import (
	"fmt"
	"net/http"
	"html/template"

	"rosperry/models"
	"rosperry/db/documents"

	"gopkg.in/mgo.v2"
	"github.com/gomodule/redigo/redis"
)

func IndexHandler(w http.ResponseWriter, r *http.Request, productsCollection *mgo.Collection, cache redis.Conn) {
	RefreshHandler(w, r, cache)
	user := ValidateAuthentication(r, cache)

	productDocuments := []documents.ProductDocument{}
	productsCollection.Find(nil).All(&productDocuments)

	products := []models.Product{}
	for _, prod := range productDocuments {
		product := models.Product{prod.Id, prod.Title, prod.Price, prod.Owner}
		products = append(products, product)
	}

	var headerTemplate string
	var indexTemplate string
	if user != " " {
		headerTemplate = "templates/header_authorized.html"
		indexTemplate = "templates/index_authorized.html"
	} else {
		headerTemplate = "templates/header_unauthorized.html"
		indexTemplate = "templates/index_unauthorized.html"
	}

	t, err := template.ParseFiles(indexTemplate, headerTemplate, "templates/footer.html")
	if err != nil {
		fmt.Println(w, err.Error())
		return
	}

	t.ExecuteTemplate(w, "index", products)
}