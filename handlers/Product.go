package handlers

import (
	"fmt"
	"strconv"
	"net/http"
	"html/template"

	"rosperry/models"
	"rosperry/utils"
	"rosperry/db/documents"

	"gopkg.in/mgo.v2"
	//"github.com/gomodule/redigo/redis"
)

func AddHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/add.html", "templates/header_unauthorized.html", "templates/footer.html")
	if err != nil {
		fmt.Println(w, err.Error())
		return
	}

	post := models.Product{}

	t.ExecuteTemplate(w, "add", post)
}

func EditHandler(w http.ResponseWriter, r *http.Request, productsCollection *mgo.Collection) {
	t, err := template.ParseFiles("templates/add.html", "templates/header_unauthorized.html", "templates/footer.html")
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

	product := models.Product{productDocument.Id, productDocument.Title, productDocument.Price}

	t.ExecuteTemplate(w, "add", product)
}

func SaveProductHandler(w http.ResponseWriter, r *http.Request, productsCollection *mgo.Collection) {
	id := r.FormValue("id")
	title := r.FormValue("title")
	price, _ := strconv.ParseInt(r.FormValue("price"), 0, 64)

	productDocument := documents.ProductDocument{id, title, price}

	if id != "" {
		productDocuments := []documents.ProductDocument{}
		err := productsCollection.Find(nil).All(&productDocuments)
		if err != nil {
			for _, doc := range productDocuments {
				if doc.Title == title {
					http.Redirect(w, r, "/add?message=namealreadyexists", 302)
				}
			}
		}

		productsCollection.UpdateId(id, productDocument)
	} else {
		id = utils.GenerateId()
		productDocument.Id = id

		err := productsCollection.Insert(productDocument)
		if err != nil {
			panic(err)
		}
	}

	http.Redirect(w, r, "/", 302)
}

func DeleteHandler(w http.ResponseWriter, r *http.Request, productsCollection *mgo.Collection) {
	id := r.FormValue("id")
	if id == "" {
		http.NotFound(w, r)
		return
	}

	productsCollection.RemoveId(id)
	http.Redirect(w, r, "/", 302)
}