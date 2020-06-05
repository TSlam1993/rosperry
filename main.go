package main

import (
	"fmt"
	"net/http"
	"rosperry/handlers"

	"gopkg.in/mgo.v2"
	"github.com/gomodule/redigo/redis"
)

var productsCollection *mgo.Collection
var usersCollection *mgo.Collection
var session *mgo.Session
var cache redis.Conn

func main() {
	fmt.Println("Listening on port :3000")

	initCache()
	initDB()

	defer session.Close()

	productsCollection = session.DB("test").C("products")
	usersCollection = session.DB("test").C("users")

	//utils.PrintCollection(usersCollection)
	//utils.DropCollection(usersCollection)

	assetsHandle := http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets/")))

	http.Handle("/assets/", assetsHandle)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
		handlers.IndexHandler(w, r, productsCollection, cache)
	})
	http.HandleFunc("/add", handlers.AddHandler)
	http.HandleFunc("/edit", func(w http.ResponseWriter, r *http.Request){
		handlers.EditHandler(w, r, productsCollection)
	})
	http.HandleFunc("/saveProduct", func(w http.ResponseWriter, r *http.Request){
		handlers.SaveProductHandler(w, r, productsCollection)
	})
	http.HandleFunc("/delete", func(w http.ResponseWriter, r *http.Request){
		handlers.DeleteHandler(w, r, productsCollection)
	})
	http.HandleFunc("/register", handlers.RegisterFormHandler)
	http.HandleFunc("/login", handlers.LoginFormHandler)
	http.HandleFunc("/signup", func(w http.ResponseWriter, r *http.Request) {
		handlers.SignUpHandler(w, r, usersCollection)
	})
	http.HandleFunc("/signin", func(w http.ResponseWriter, r *http.Request) {
		handlers.SignInHandler(w, r, usersCollection, cache)
	})
	http.HandleFunc("/signout", func(w http.ResponseWriter, r *http.Request){
		handlers.SignOutHandler(w, r, cache)
	})
	http.HandleFunc("/welcome", func(w http.ResponseWriter, r *http.Request){
		handlers.WelcomeHandler(w, r, cache)
	})
	http.HandleFunc("/refresh", func(w http.ResponseWriter, r *http.Request){
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