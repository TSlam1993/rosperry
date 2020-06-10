package handlers

import (
	"fmt"
	"time"
	"strconv"
	"net/http"
	"html/template"

	//"rosperry/models"
	"rosperry/utils"
	"rosperry/db/documents"

	"gopkg.in/mgo.v2"
	"github.com/gomodule/redigo/redis"
	"go.mongodb.org/mongo-driver/bson"
)

var header string

func AddHandler(w http.ResponseWriter, r *http.Request, cache redis.Conn) {
	user := ValidateAuthentication(r, cache)
	if user == " " {
		http.Redirect(w, r, "/", 302)
		return
	}

	t, err := template.ParseFiles("templates/product/add.html", headerAuthorizedTemplate, footerTemplate)
	if err != nil {
		panic(err)
	}

	product := documents.TemplateProductDocument{}

	t.ExecuteTemplate(w, "add", product)
}

func ShowHandler(w http.ResponseWriter, r *http.Request, productsCollection *mgo.Collection, cache redis.Conn) {
	user := ValidateAuthentication(r, cache)
	if user == " " {
		header = headerUnauthorizedTemplate
	} else {
		header = headerAuthorizedTemplate
	}

	t, err := template.ParseFiles(showTemplate, header, footerTemplate)
	if err != nil {
		panic(err)
	}

	id := r.FormValue("id")
	productDocument := documents.ProductDocument{}

	err = productsCollection.FindId(id).One(&productDocument)
	if err != nil {
		fmt.Println("error", err)
		http.Redirect(w, r, "/", 302)
		return
	}

	ownsProduct := false
	if user == productDocument.Owner {
		ownsProduct = true
	}

	product := documents.TemplateProductDocument{productDocument.Id, productDocument.Title,
		productDocument.Price, productDocument.Owner, productDocument.Type,
		productDocument.CreatedAt.Format("01.02.2006"),
		productDocument.UpdatedAt.Format("01.02.2006"), " ", ownsProduct}

	fmt.Println("showing product: ", product)

	t.ExecuteTemplate(w, "show", product)
}

func EditHandler(w http.ResponseWriter, r *http.Request, productsCollection *mgo.Collection, cache redis.Conn) {
	user := ValidateAuthentication(r, cache)
	if user == " " {
		http.Redirect(w, r, "/", 302)
		return
	}

	t, err := template.ParseFiles(addTemplate, headerAuthorizedTemplate, footerTemplate)
	if err != nil {
		fmt.Println(w, err.Error())
		return
	}

	id := r.FormValue("id")
	productDocument := documents.ProductDocument{}

	err = productsCollection.FindId(id).One(&productDocument)
	if err != nil {
		fmt.Println("error", err)
		http.Redirect(w, r, "/", 302)
	}

	ownsProduct := false
	if user == productDocument.Owner {
		ownsProduct = true
	}

	product := documents.TemplateProductDocument{productDocument.Id, productDocument.Title,
		productDocument.Price, productDocument.Owner, productDocument.Type,
		productDocument.CreatedAt.Format("01.02.2006"),
		productDocument.UpdatedAt.Format("01.02.2006"), " ", ownsProduct}

	getQuery := r.URL.Query()["message"]
	if len(getQuery) > 0 {
		message := getQuery[0]
		product.Message = message
	}

	fmt.Println("editing product: ", productDocument)

	t.ExecuteTemplate(w, "add", product)
}

func SaveProductHandler(w http.ResponseWriter, r *http.Request, productsCollection *mgo.Collection, cache redis.Conn) {
	id := r.FormValue("id")
	title := r.FormValue("title")
	price, _ := strconv.ParseInt(r.FormValue("price"), 0, 64)
	productType := r.FormValue("type")

	c, err := r.Cookie("auth_token")
	if err != nil {
		fmt.Println(err)
	}
	authToken := c.Value

	owner, err := cache.Do("GET", authToken)
	if err != nil {
		fmt.Println(err)
	}

	dt := time.Now()
	updatedAt := dt
	createdAt := dt

	productDocument := documents.ProductDocument{id, title, price, fmt.Sprintf("%s", owner), productType, createdAt, updatedAt}

	if id != "" {
		existingProduct := documents.ProductDocument{}
		err = productsCollection.FindId(id).One(&existingProduct)
		if err != nil {
			panic(err)
		}
		productDocument.CreatedAt = existingProduct.CreatedAt
		fmt.Println("now updating product: ", productDocument)

		err = productsCollection.UpdateId(id, productDocument)
		if err != nil {
			panic(err)
		}
	} else {
		id = utils.GenerateId()
		productDocument.Id = id

		existingProducts := []documents.ProductDocument{}
		err := productsCollection.Find(bson.M{"_title": title}).All(&existingProducts)

		if err != nil {
			fmt.Println(err)
			if fmt.Sprintf("%s", err) == "not found" {
				err := productsCollection.Insert(productDocument)
				if err != nil {
					panic(err)
				}
			} else {
				panic(err)
			}
		} else {
			for _, doc := range existingProducts {
				fmt.Println(doc.Owner, fmt.Sprintf("%s", owner), doc.Owner == fmt.Sprintf("%s", owner))
				if doc.Owner == fmt.Sprintf("%s", owner) {
					fmt.Println("now redirecting")
					http.Redirect(w, r, "/add?message=namealreadyexists", 302)
					return
				}
			}
			fmt.Println("now inserting product: ", productDocument)
			err := productsCollection.Insert(productDocument)
			if err != nil {
				panic(err)
			}
		}
	}

	http.Redirect(w, r, "/", 302)
}

func DeleteHandler(w http.ResponseWriter, r *http.Request, productsCollection *mgo.Collection, cache redis.Conn) {
	user := ValidateAuthentication(r, cache)
	if user == " " {
		http.Redirect(w, r, "/", 302)
		return
	}

	id := r.FormValue("id")
	if id == "" {
		http.NotFound(w, r)
		return
	}

	productsCollection.RemoveId(id)
	http.Redirect(w, r, "/", 302)
}