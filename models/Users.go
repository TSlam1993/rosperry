package models

type User struct {
	Username string
	Email string
	Password []byte
	BusinessName string
	AgeOfBusiness int64
	Location string
	CreatedAt time.Time
	UpdatedAt time.Time
	LastLogged time.Time
}

func NewUser(username string, email string, password []byte) *User {
	return &User{username, email, password}
}