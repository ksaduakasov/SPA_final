package main

import (
	"aitu/business/data/schema"
	"aitu/foundation/database"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"io/ioutil"
	"log"
	"os"
	"time"
)

func main() {
	tokengen()
}


func migrate(){
	dbConfig := database.Config{
		User: "postgres",
		Password: "1952",
		Host: "0.0.0.0",
		Name: "postgres",
		DisableTLS: true,
	}
	db, err := database.Open(dbConfig)
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()
	if err := schema.Migrate(db); err != nil {
		log.Fatalln(err)
	}
	fmt.Println("migrations complete")
}

func tokengen() {
	privatePem, err := ioutil.ReadFile("private.pem")
	if err != nil {
		log.Fatalln(err)
	}
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privatePem)

	claims := struct {
		jwt.StandardClaims
		Roles []string `json:"roles"`
	}{

		jwt.StandardClaims{
			Issuer:    "service project",
			Subject:   "123456789",
			ExpiresAt: time.Now().Add(8760 * time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		[]string{"ADMIN"},
	}

	method := jwt.GetSigningMethod("RS256")
	tkn := jwt.NewWithClaims(method, claims)
	tkn.Header["kid"] = "54bb2165-71e1-41a6-af3e-7da4a0e1e2c1"

	str, err := tkn.SignedString(privateKey)
	if err != nil{
		log.Fatalln(err)
	}

	fmt.Printf("----BEGIN TOKEN----\n%s\n----END TOKEN----\n", str)



}

func keygen() {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalln(err)
	}

	privateFile, err := os.Create("private.pem")
	if err != nil {
		log.Fatalln(err)
	}
	defer privateFile.Close()

	privateBlock := pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	if err := pem.Encode(privateFile, &privateBlock); err != nil {
		log.Fatalln(err)
	}

	// = = = = = = =

	asn1Bytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		log.Fatalln(err)
	}

	publicFile, err := os.Create("public.pem")
	if err != nil {
		log.Fatalln(err)
	}
	defer publicFile.Close()

	publicBlock := pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: asn1Bytes,
	}

	if err := pem.Encode(publicFile, &publicBlock); err != nil {
		log.Fatalln(err)
	}

	log.Println("DONE")

}
