package handlers

import (
	"fmt"
	"net/http"
	"html/template"

	"rosperry/db/documents"

	"gopkg.in/mgo.v2"
	"github.com/gomodule/redigo/redis"
)

func UsersHandler(w http.ResponseWriter, r *http.Request, usersCollection *mgo.Collection, cache redis.Conn) {
	user := ValidateAuthentication(r, cache)

	userDocuments := []documents.UserDocument{}
	usersCollection.Find(nil).All(&userDocuments)

	var headerTemplate string
	if user != " " {
		headerTemplate = headerAuthorizedTemplate
	} else {
		headerTemplate = headerUnauthorizedTemplate
	}

	usersTemplateData := []documents.TemplateUserDocument{}
	isUser := false
	for _, doc := range userDocuments {
		isUser = false
		if user == doc.Username {
			isUser = true
		}
		user := documents.TemplateUserDocument{
			doc.Username, doc.Email, doc.Password,
			doc.BusinessName, doc.AgeOfBusiness, doc.Location,
			doc.CreatedAt.Format("01.02.2006"),
			doc.UpdatedAt.Format("01.02.2006"),
			doc.LastLogged.Format("01.02.2006"), isUser, "",
		}
		usersTemplateData = append(usersTemplateData, user)
	}

	t, err := template.ParseFiles(usersTemplate, headerTemplate, footerTemplate)
	if err != nil {
		panic(err)
	}

	t.ExecuteTemplate(w, "users", usersTemplateData)
}

func UserHandler(w http.ResponseWriter, r *http.Request, usersCollection *mgo.Collection, cache redis.Conn) {
	user := ValidateAuthentication(r, cache)
	if user == " " {
		header = headerUnauthorizedTemplate
	} else {
		header = headerAuthorizedTemplate
	}

	t, err := template.ParseFiles(userTemplate, header, footerTemplate)
	if err != nil {
		panic(err)
	}

	username := r.FormValue("username")
	userDocument := documents.UserDocument{}

	err = usersCollection.FindId(username).One(&userDocument)
	if err != nil {
		fmt.Println("error", err)
		http.Redirect(w, r, "/", 302)
		return
	}

	isUser := false
	if user == userDocument.Username {
		isUser = true
	}

	userTemplateData := documents.TemplateUserDocument{
		userDocument.Username, userDocument.Email,
		userDocument.Password, userDocument.BusinessName,
		userDocument.AgeOfBusiness, userDocument.Location,
		userDocument.CreatedAt.Format("01.02.2006"),
		userDocument.UpdatedAt.Format("01.02.2006"),
		userDocument.LastLogged.Format("01.02.2006"), isUser, "",
	}

	t.ExecuteTemplate(w, "user", userTemplateData)
}

func UserPageHandler(w http.ResponseWriter, r *http.Request, usersCollection *mgo.Collection, cache redis.Conn) {
	// TO DO: set current username in formvalue
	UserHandler(w, r, usersCollection, cache)
}