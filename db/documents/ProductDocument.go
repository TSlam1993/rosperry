package documents

type ProductDocument struct {
	Id string  `bson:"_id,omitempty"`
	Title string
	Price int64
	Owner string
}
