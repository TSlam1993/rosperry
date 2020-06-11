package handlers

import (
	"time"
	"fmt"
	"net/http"
	"html/template"

	"golang.org/x/crypto/bcrypt"
	"github.com/gomodule/redigo/redis"
	"github.com/satori/go.uuid"
	"gopkg.in/mgo.v2"
	"go.mongodb.org/mongo-driver/bson"

	"rosperry/db/documents"
)

type signParams struct {
	Message string
	Username string
	Email string
}

const hashCost = 8

func LoginFormHandler(w http.ResponseWriter, r *http.Request, cache redis.Conn) {
	user := ValidateAuthentication(r, cache)
	if user != " " {
		http.Redirect(w, r, "/", 302)
		return
	}

	t, err := template.ParseFiles(loginTemplate, headerUnauthorizedTemplate, footerTemplate)
	if err != nil {
		panic(err)
	}

	params := signParams{}

	getQuery := r.URL.Query()["message"]
	if len(getQuery) > 0 {
		message := getQuery[0]
		params = signParams{Message:message}
	}

	t.ExecuteTemplate(w, "login", params)
}

func RegisterFormHandler(w http.ResponseWriter, r *http.Request, cache redis.Conn) {
	user := ValidateAuthentication(r, cache)
	if user != " " {
		http.Redirect(w, r, "/", 302)
		return
	}

	t, err := template.ParseFiles(registerTemplate, headerUnauthorizedTemplate, footerTemplate)
	if err != nil {
		panic(err)
	}

	params := signParams{}

	getQuery := r.URL.Query()["message"]
	if len(getQuery) > 0 {
		message := getQuery[0]
		params = signParams{Message:message}
	}

	t.ExecuteTemplate(w, "register", params)
}

func SignUpHandler(w http.ResponseWriter, r *http.Request, usersCollection *mgo.Collection) {
	username := r.FormValue("username")
	password := r.FormValue("password")
	email := r.FormValue("email")

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), hashCost)
	dt := time.Now()
	existingUsers := []documents.UserDocument{}
	userDocument := documents.UserDocument{username, email, hashedPassword,
		"", 0, "", dt, dt, dt}

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

func SignInHandler(w http.ResponseWriter, r *http.Request, usersCollection *mgo.Collection, cache redis.Conn) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	userDocument := documents.UserDocument{}
	err := usersCollection.FindId(username).One(&userDocument)
	if err != nil {
		http.Redirect(w, r, "/login?message=notfound", 302)
		return
	}

	err = bcrypt.CompareHashAndPassword(userDocument.Password, []byte(password))
	if err != nil {
		http.Redirect(w, r, "/login?message=wrongpassword", 302)
		return
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

func SignOutHandler(w http.ResponseWriter, r *http.Request, cache redis.Conn) {
	// TO DO: FIX ERROR WHEN REFRESH TOKEN IS EXPIRED
	http.SetCookie(w, &http.Cookie{
		Name: "username",
		Value: "",
		Expires: time.Now(),
	})

	c, err := r.Cookie("refresh_token")
	if err != nil {
		fmt.Println("refresh token cookie error: ", err)
	}
	refreshToken := c.Value

	refreshTokenUsername, err := cache.Do("GET", refreshToken)
	if err != nil {
		fmt.Println(err)
	}
	if refreshTokenUsername != nil {
		_, err = cache.Do("DEL", refreshToken)
		if err != nil {
			fmt.Println("refresh token username error: ", err)
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
		fmt.Println("auth token caceh error: ", err)
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

func WelcomeHandler(w http.ResponseWriter, r *http.Request, cache redis.Conn) {
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

func RefreshHandler(w http.ResponseWriter, r *http.Request, cache redis.Conn) {
	c, err := r.Cookie("username")
	if err != nil {
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
		fmt.Println("refreshing tokens: error getting refresh token cache: ", err)
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
		fmt.Println("refreshing tokes: error setting new token in cache: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name: "auth_token",
		Value: newAuthToken,
		Expires: time.Now().Add(120 * time.Second),
	})
}

func ValidateAuthentication(r *http.Request, cache redis.Conn) string {
	c, err := r.Cookie("auth_token")
	if err != nil {
		return " "
	}

	authToken := c.Value

	c, err = r.Cookie("username")
	if err != nil {
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