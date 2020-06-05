package models

type User struct {
	Username string
	Email string
	Password []byte
}

func NewUser(username string, email string, password []byte) *User {
	return &User{username, email, password}
}