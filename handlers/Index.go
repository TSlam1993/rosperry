package handlers

import (
	"fmt"
	"net/http"
	"html/template"

	//"rosperry/models"
	"rosperry/db/documents"

	"gopkg.in/mgo.v2"
	"github.com/gomodule/redigo/redis"
	//"go.mongodb.org/mongo-driver/bson"
)

const headerAuthorizedTemplate = "templates/landing/header_authorized.html"
const headerUnauthorizedTemplate = "templates/landing/header_unauthorized.html"
const footerTemplate = "templates/landing/footer.html"
const loginTemplate = "templates/authorization/login.html"
const registerTemplate = "templates/authorization/register.html"
const addTemplate = "templates/product/add.html"
const showTemplate = "templates/product/show.html"
const indexTemplate = "templates/landing/index_authorized.html"

func IndexHandler(w http.ResponseWriter, r *http.Request, productsCollection *mgo.Collection, cache redis.Conn) {
	RefreshHandler(w, r, cache)
	user := ValidateAuthentication(r, cache)

	productDocuments := []documents.ProductDocument{}
	productsCollection.Find(nil).All(&productDocuments)
	fmt.Println(productDocuments)

	products := []documents.TemplateProductDocument{}
	for _, prod := range productDocuments {
		product := documents.TemplateProductDocument{prod.Id, prod.Title, prod.Price,
			prod.Owner, prod.Type,
			prod.CreatedAt.Format("01-02-2006 15:04"),
			prod.UpdatedAt.Format("01-02-2006 15:04"),
			""}
		products = append(products, product)
	}


	var headerTemplate string
	if user != " " {
		headerTemplate = headerAuthorizedTemplate
	} else {
		headerTemplate = headerUnauthorizedTemplate
	}
	fmt.Println(headerTemplate)

	t, err := template.ParseFiles(indexTemplate, headerTemplate, footerTemplate)
	if err != nil {
		panic(err)
	}

	t.ExecuteTemplate(w, "index", products)
}