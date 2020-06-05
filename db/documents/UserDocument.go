package documents

type UserDocument struct {
	Username string `bson:"_id,omitempty"`
	Email string
	Password []byte
}