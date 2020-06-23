package handlers

import (
	"time"
	"fmt"
	"strconv"
	"net/http"
	"html/template"

	"gopkg.in/mgo.v2"
	"golang.org/x/crypto/bcrypt"
	"go.mongodb.org/mongo-driver/bson"
	"github.com/gomodule/redigo/redis"

	"rosperry/db/documents"
	"rosperry/utils"
)

func UsersHandler(w http.ResponseWriter, r *http.Request, usersCollection *mgo.Collection, cache redis.Conn) {
	user := ValidateAuthentication(r, cache)

	userDocuments := []documents.UserDocument{}
	usersCollection.Find(nil).All(&userDocuments)

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

	t, err := template.ParseFiles(usersTemplate, headerAuthorizedTemplate, footerTemplate)
	if err != nil {
		panic(err)
	}

	t.ExecuteTemplate(w, "users", usersTemplateData)
}

func ShowUserHandler(w http.ResponseWriter, r *http.Request, usersCollection *mgo.Collection, cache redis.Conn) {
	user := ValidateAuthentication(r, cache)
	t, err := template.ParseFiles(showUserTemplate, headerAuthorizedTemplate, footerTemplate)
	if err != nil {
		panic(err)
	}

	username := r.FormValue("username")
	if username == "self" {
		username = user
	}
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

	t.ExecuteTemplate(w, "showUser", userTemplateData)
}

func EditUserHandler(w http.ResponseWriter, r *http.Request, usersCollection *mgo.Collection, cache redis.Conn) {
	user := ValidateAuthentication(r, cache)
	t, err := template.ParseFiles(editUserTemplate, headerAuthorizedTemplate, footerTemplate)
	if err != nil {
		panic(err)
	}

	username := r.FormValue("username")
	if username == "#" {
		username = user
	}
	userDocument := documents.UserDocument{}

	fmt.Println("user: ", username)
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

	t.ExecuteTemplate(w, "editUser", userTemplateData)
}

func SaveUserHandler(w http.ResponseWriter, r *http.Request, usersCollection *mgo.Collection, cache redis.Conn) {
	username := r.FormValue("username")
	password := r.FormValue("password")
	email := r.FormValue("email")
	businessName := r.FormValue("businessName")
	ageOfBusiness, _ := strconv.ParseInt(r.FormValue("ageOfBusiness"), 0, 64)

	ip := utils.GetIp(r)
	location := utils.GetLocation(ip)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), hashCost)
	dt := time.Now()

	existingUsers := []documents.UserDocument{}
	userDocument := documents.UserDocument{
		username, email, hashedPassword,
		businessName, ageOfBusiness,
		location, dt, dt, dt,
	}

	err = usersCollection.Find(bson.M{"_email": email}).All(&existingUsers)

	if len(existingUsers) == 0 {
		err = usersCollection.Insert(userDocument)
		if err != nil {
			panic(err)
		}

		http.Redirect(w, r, "/login?message=registersuccess", 302)
		return
	} else {
		http.Redirect(w, r, "/register?message=emailalreadyexists", 302)
	}
}

func UserCabinetHandler(w http.ResponseWriter, r *http.Request, productsCollection *mgo.Collection, cache redis.Conn) {
	RefreshHandler(w, r, cache)
	user := ValidateAuthentication(r, cache)

	productDocuments := []documents.ProductDocument{}
	productsCollection.Find(nil).All(&productDocuments)

	products := []documents.TemplateProductDocument{}
	ownsProduct := false
	for _, prod := range productDocuments {
		ownsProduct = false
		if user == prod.Owner {
			ownsProduct = true
		}

		product := documents.TemplateProductDocument{prod.Id, prod.Title, prod.Price,
			prod.Owner, prod.Type,
			prod.CreatedAt.Format("01.02.2006"),
			prod.UpdatedAt.Format("01.02.2006"), "", ownsProduct}
		if ownsProduct {
			products = append(products, product)
		}
	}

	t, err := template.ParseFiles(indexTemplate, headerAuthorizedTemplate, footerTemplate)
	if err != nil {
		panic(err)
	}

	t.ExecuteTemplate(w, "index", products)
}