package handlers

import (
	//"fmt"
	"net/http"
	"html/template"

	"rosperry/db/documents"
	//"rosperry/utils"

	"gopkg.in/mgo.v2"
	"github.com/gomodule/redigo/redis"
)

const headerAuthorizedTemplate = "templates/landing/header_authorized.html"
const headerUnauthorizedTemplate = "templates/landing/header_unauthorized.html"
const footerTemplate = "templates/landing/footer.html"
const loginTemplate = "templates/authorization/login.html"
const registerTemplate = "templates/authorization/register.html"
const addProductTemplate = "templates/product/add.html"
const showProductTemplate = "templates/product/show.html"
const indexTemplate = "templates/landing/index_authorized.html"
const usersTemplate = "templates/user/users.html"
const showUserTemplate = "templates/user/show.html"
const editUserTemplate = "templates/user/edit.html"
const editSearchParametersTemplate = "templates/user/edit_search_parameters.html"

func IndexHandler(w http.ResponseWriter, r *http.Request, productsCollection *mgo.Collection, usersCollection *mgo.Collection, cache redis.Conn) {
	RefreshHandler(w, r, cache)
	user := ValidateAuthentication(r, cache)

	var headerTemplate string
	if user != " " {
		headerTemplate = headerAuthorizedTemplate
	} else {
		headerTemplate = headerUnauthorizedTemplate
	}

	productDocuments := []documents.ProductDocument{}
	productsCollection.Find(nil).All(&productDocuments)

	products := []documents.TemplateProductDocument{}
	ownsProduct := false
	for _, prod := range productDocuments {
		ownsProduct = false
		if user == prod.Owner {
			ownsProduct = true
		}

		var businessName string
		userDocument := documents.UserDocument{}
		err := usersCollection.FindId(prod.Owner).One(&userDocument)

		if err == nil {
			businessName = userDocument.BusinessName
		} else {
			businessName = prod.Owner
		}

		product := documents.TemplateProductDocument{prod.Id, prod.Title, prod.Price,
			businessName, prod.Type,
			prod.CreatedAt.Format("01.02.2006"),
			prod.UpdatedAt.Format("01.02.2006"), "", ownsProduct}
		products = append(products, product)
	}

	t, err := template.ParseFiles(indexTemplate, headerTemplate, footerTemplate)
	if err != nil {
		panic(err)
	}

	t.ExecuteTemplate(w, "index", products)
	return
}