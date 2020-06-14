package utils

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"net"
	"encoding/json"
	"io/ioutil"

	"gopkg.in/mgo.v2"

	"rosperry/db/documents"
)

var (
	apiAccessKey = "32c29d07f044e91126f02fc6feb42f74"
)

func GenerateId() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func DropCollection(collection *mgo.Collection) {
	collection.RemoveAll(nil)
}

func PrintUserCollection(collection *mgo.Collection) {
	collectionDocuments := []documents.UserDocument{}
	collection.Find(nil).All(&collectionDocuments)

	for _, doc := range collectionDocuments {
		fmt.Println(doc)
	}
}

func PrintProductCollection(collection *mgo.Collection) {
	collectionDocuments := []documents.ProductDocument{}
	collection.Find(nil).All(&collectionDocuments)

	for _, doc := range collectionDocuments {
		fmt.Println(doc)
	}
}

func GetLocation(ip string) string {
	//my ip for dev, comment out to use ip from GetIp
	ip = "188.232.181.218"
	ipstackUrl := "http://api.ipstack.com/"+ip+"?access_key="+apiAccessKey

	response, err := http.Get(ipstackUrl)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	bodyString := string(body)

	var result map[string]interface{}
	json.Unmarshal([]byte(bodyString), &result)
	country := result["country_name"]

	return fmt.Sprintf("%v", country)
}

func GetIp(r *http.Request) string {
	if ipProxy := r.Header.Get("X-FORWARDED-FOR"); len(ipProxy) > 0 {
		return ipProxy
	}
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)

	userIP := net.ParseIP(ip)
	if userIP == nil {
		fmt.Println("userip: is not IP:port", r.RemoteAddr)
		return ""
	}
	return userIP.String()
}