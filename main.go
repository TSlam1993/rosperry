package main

import (
	"time"
	"fmt"
	"strconv"
	"net/http"
	"html/template"

	"rosperry/models"
	"rosperry/utils"
	"rosperry/db/documents"

	"gopkg.in/mgo.v2"
	"golang.org/x/crypto/bcrypt"
	"github.com/gomodule/redigo/redis"
	"github.com/satori/go.uuid"
)

var productsCollection *mgo.Collection
var usersCollection *mgo.Collection
var session *mgo.Session
var cache redis.Conn
const hashCost = 8

func indexHandler(w http.ResponseWriter, r *http.Request) {
	refreshHandler(w, r)
	user := validateAuthentication(r)

	productDocuments := []documents.ProductDocument{}
	productsCollection.Find(nil).All(&productDocuments)

	products := []models.Product{}
	for _, prod := range productDocuments {
		product := models.Product{prod.Id, prod.Title, prod.Price}
		products = append(products, product)
	}

	var headerTemplate string
	if user != " " {
		headerTemplate = "templates/header_authorized.html"
	} else {
		headerTemplate = "templates/header_unauthorized.html"
	}

	t, err := template.ParseFiles("templates/index.html", headerTemplate, "templates/footer.html")
	if err != nil {
		fmt.Println(w, err.Error())
		return
	}

	t.ExecuteTemplate(w, "index", products)
}

func addHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/add.html", "templates/header_unauthorized.html", "templates/footer.html")
	if err != nil {
		fmt.Println(w, err.Error())
		return
	}

	post := models.Product{}

	t.ExecuteTemplate(w, "add", post)
}

func editHandler(w http.ResponseWriter, r *http.Request) {
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

func saveProductHandler(w http.ResponseWriter, r *http.Request) {
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

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		http.NotFound(w, r)
		return
	}

	productsCollection.RemoveId(id)
	http.Redirect(w, r, "/", 302)
}

type signParams struct {
	Message string
	Username string
	Email string
}

func loginFormHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/login.html", "templates/header_unauthorized.html", "templates/footer.html")

	if err != nil {
		fmt.Println(w, err.Error())
		return
	}

	params := signParams{}

	getQuery := r.URL.Query()["message"]
	if len(getQuery) > 0 {
		message := getQuery[0]
		params = signParams{Message:message}
	}

	t.ExecuteTemplate(w, "login", params)
}

func registerFormHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/register.html", "templates/header_unauthorized.html", "templates/footer.html")

	if err != nil {
		fmt.Println(w, err.Error())
		return
	}

	params := signParams{}

	getQuery := r.URL.Query()["message"]
	if len(getQuery) > 0 {
		message := getQuery[0]
		params = signParams{Message:message}
	}

	t.ExecuteTemplate(w, "register", params)
}

func signUpHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")
	email := r.FormValue("email")

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), hashCost)

	userDocument := documents.UserDocument{username, email, hashedPassword}
	err = usersCollection.FindId(username).One(&userDocument)
	if err != nil {
		userDocuments := []documents.UserDocument{}
		usersCollection.Find(nil).All(&userDocuments)
		for _, doc := range userDocuments {
			if doc.Email == email {
				http.Redirect(w, r, "/register?message=emailalreadyexists", 302)
			}
		}

		err := usersCollection.Insert(userDocument)
		if err != nil {
			panic(err)
		}

		http.Redirect(w, r, "/login?message=registersuccess", 302)
	}

	http.Redirect(w, r, "/register?message=namealreadyexists", 302)
}

func signInHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")
	fmt.Println(password)

	userDocument := documents.UserDocument{}
	err := usersCollection.FindId(username).One(&userDocument)
	if err != nil {
		http.Redirect(w, r, "/login?message=notfound", 302)
	}

	err = bcrypt.CompareHashAndPassword(userDocument.Password, []byte(password))
	if err != nil {
		http.Redirect(w, r, "/login?message=wrongpassword", 302)
	}

	authToken := uuid.Must(uuid.NewV4()).String()
	_, err = cache.Do("SETEX", authToken, "120", username)
	if err != nil {
		panic(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name: "auth_token",
		Value: authToken,
		Expires: time.Now().Add(120 * time.Second),
	})

	http.SetCookie(w, &http.Cookie{
		Name: "username",
		Value: username,
		Expires: time.Now().Add(120 * time.Hour),
	})

	refreshToken := uuid.Must(uuid.NewV4()).String()
	_, err = cache.Do("SETEX", refreshToken, "432000", username)
	if err != nil {
		panic(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name: "refresh_token",
		Value: refreshToken,
		Expires: time.Now().Add(120 * time.Hour),
	})

	http.Redirect(w, r, "/", 302)
}

func signOutHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("signing out, BRO :)))))))))))))00")

	http.SetCookie(w, &http.Cookie{
		Name: "username",
		Value: "",
		Expires: time.Now(),
	})

	c, err := r.Cookie("refresh_token")
	if err != nil {
		fmt.Println(err)
	}
	refreshToken := c.Value

	refreshTokenUsername, err := cache.Do("GET", refreshToken)
	if err != nil {
		fmt.Println(err)
	}
	if refreshTokenUsername != nil {
		_, err = cache.Do("DEL", refreshToken)
		if err != nil {
			fmt.Println(err)
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name: "refresh_token",
		Value: "",
		Expires: time.Now(),
	})


	c, err = r.Cookie("auth_token")
	if err != nil {
		fmt.Println(err)
	}
	authToken := c.Value

	authTokenUsername, err := cache.Do("GET", authToken)
	if err != nil {
		fmt.Println(err)
	}
	if authTokenUsername != nil {
		_, err = cache.Do("DEL", authToken)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name: "auth_token",
		Value: "",
		Expires: time.Now(),
	})

	http.Redirect(w, r, "/", 302)
}

func welcomeHandler(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("auth_token")
	if err != nil {
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	authToken := c.Value

	response, err := cache.Do("GET", authToken)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if response == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	w.Write([]byte(fmt.Sprintf("Welcome %s!", response)))
}

func refreshHandler(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("username")
	if err != nil {
		fmt.Println(err)
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	cookieUsername := c.Value

	newAuthToken := uuid.Must(uuid.NewV4()).String()

	c, err = r.Cookie("refresh_token")
	if err != nil {
		fmt.Println(err)
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	refreshToken := c.Value

	refreshTokenUsername, err := cache.Do("GET", refreshToken)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if refreshTokenUsername == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if fmt.Sprintf("%s", refreshTokenUsername) != cookieUsername {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	_, err = cache.Do("SETEX", newAuthToken, "120", cookieUsername)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name: "auth_token",
		Value: newAuthToken,
		Expires: time.Now().Add(120 * time.Second),
	})

	fmt.Println("DATA: ", refreshToken, newAuthToken, fmt.Sprintf("%s", refreshTokenUsername), cookieUsername)
}

func validateAuthentication(r *http.Request) string {
	c, err := r.Cookie("auth_token")
	if err != nil {
		fmt.Println(err)
		return " "
	}

	authToken := c.Value

	c, err = r.Cookie("username")
	if err != nil {
		fmt.Println(err)
		return " "
	}

	cookieUsername := c.Value

	tokenUsername, err := cache.Do("GET", authToken)
	if err != nil {
		panic(err)
	}

	if fmt.Sprintf("%s", tokenUsername) == cookieUsername {
		return cookieUsername
	} else {
		return " "
	}
}

func main() {
	fmt.Println("Listening on port :3000")

	initCache()
	initDB()

	defer session.Close()

	productsCollection = session.DB("test").C("products")
	usersCollection = session.DB("test").C("users")

	//printCollection(usersCollection)
	//dropCollection(usersCollection)

	assetsHandle := http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets/")))

	http.Handle("/assets/", assetsHandle)
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/add", addHandler)
	http.HandleFunc("/edit", editHandler)
	http.HandleFunc("/saveProduct", saveProductHandler)
	http.HandleFunc("/delete", deleteHandler)
	http.HandleFunc("/register", registerFormHandler)
	http.HandleFunc("/login", loginFormHandler)
	http.HandleFunc("/signup", signUpHandler)
	http.HandleFunc("/signin", signInHandler)
	http.HandleFunc("/signout", signOutHandler)
	http.HandleFunc("/welcome", welcomeHandler)
	http.HandleFunc("/refresh", refreshHandler)

	http.ListenAndServe(":3000", nil)
}

func initDB() {
	localSession, err := mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}

	session = localSession
}

func initCache() {
	conn, err := redis.DialURL("redis://localhost")
	if err != nil {
		panic(err)
	}
	cache = conn
}

func dropCollection(collection *mgo.Collection) {
	collection.RemoveAll(nil)
}

func printCollection(collection *mgo.Collection) {
	collectionDocuments := []documents.UserDocument{}
	collection.Find(nil).All(&collectionDocuments)

	for _, doc := range collectionDocuments {
		fmt.Println(doc)
	}
}