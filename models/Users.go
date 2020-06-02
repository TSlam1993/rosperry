package models

type User struct {
	Username string
	Email string
	Password string
}

func NewUser(username string, email string, password string) *User {
	return &User{username, email, password}
}