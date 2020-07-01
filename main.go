package main

import (
	"fmt"
	"net/http"
	"rosperry/handlers"
	//"rosperry/utils"

	"gopkg.in/mgo.v2"
	"github.com/gomodule/redigo/redis"
)

var (
	productsCollection *mgo.Collection
	usersCollection *mgo.Collection
	session *mgo.Session
	cache redis.Conn
)

func main() {
	fmt.Println("Listening on port :3000")

	initCache()
	initDB()

	defer session.Close()

	productsCollection = session.DB("test").C("products")
	usersCollection = session.DB("test").C("users")

	//utils.PrintProductCollection(productsCollection)
	//utils.PrintUserCollection(usersCollection)
	//utils.DropCollection(usersCollection)

	assetsHandle := http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets/")))

	http.Handle("/assets/", assetsHandle)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
		handlers.IndexHandler(w, r, productsCollection, usersCollection, cache)
	})
	http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		handlers.UsersHandler(w, r, usersCollection, cache)
	})
	http.HandleFunc("/user/show", func(w http.ResponseWriter, r *http.Request) {
		handlers.ShowUserHandler(w, r, usersCollection, cache)
	})
	http.HandleFunc("/user/edit", func(w http.ResponseWriter, r *http.Request) {
		handlers.EditUserHandler(w, r, usersCollection, cache)
	})
	http.HandleFunc("/user/update", func(w http.ResponseWriter, r *http.Request) {
		handlers.UpdateUserHandler(w, r, usersCollection, cache)
	})
	http.HandleFunc("/user/updateSearchInfo", func(w http.ResponseWriter, r *http.Request) {
		handlers.UpdateSearchInfoHandler(w, r, usersCollection, cache)
	})
	http.HandleFunc("/user/cabinet", func(w http.ResponseWriter, r *http.Request) {
		handlers.UserCabinetHandler(w, r, productsCollection, cache)
	})
	http.HandleFunc("/product/add", func(w http.ResponseWriter, r *http.Request) {
		handlers.AddProductHandler(w, r, cache)
	})
	http.HandleFunc("/product/show", func(w http.ResponseWriter, r *http.Request) {
		handlers.ShowProductHandler(w, r, productsCollection, usersCollection, cache)
	})
	http.HandleFunc("/product/edit", func(w http.ResponseWriter, r *http.Request) {
		handlers.EditProductHandler(w, r, productsCollection, cache)
	})
	http.HandleFunc("/product/save", func(w http.ResponseWriter, r *http.Request) {
		handlers.SaveProductHandler(w, r, productsCollection, cache)
	})
	http.HandleFunc("/product/delete", func(w http.ResponseWriter, r *http.Request) {
		handlers.DeleteProductHandler(w, r, productsCollection, cache)
	})
	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		handlers.RegisterFormHandler(w, r, cache)
	})
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		handlers.LoginFormHandler(w, r, cache)
	})
	http.HandleFunc("/signup", func(w http.ResponseWriter, r *http.Request) {
		handlers.SignUpHandler(w, r, usersCollection)
	})
	http.HandleFunc("/signin", func(w http.ResponseWriter, r *http.Request) {
		handlers.SignInHandler(w, r, usersCollection, cache)
	})
	http.HandleFunc("/signout", func(w http.ResponseWriter, r *http.Request) {
		handlers.SignOutHandler(w, r, cache)
	})
	http.HandleFunc("/welcome", func(w http.ResponseWriter, r *http.Request) {
		handlers.WelcomeHandler(w, r, cache)
	})
	http.HandleFunc("/refresh", func(w http.ResponseWriter, r *http.Request) {
		handlers.RefreshHandler(w, r, cache)
	})

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