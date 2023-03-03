package script

import (
	"log"
	"math/rand"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!/?#@"

func GenerateHash(password string) string {

	var mpCrypt []byte
	cost := 11
	mpCrypt, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		log.Fatal(err)
	}

	return string(mpCrypt)
}

func ComparePassword(hashedPassword string, password string) bool {

	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))

	if err != nil {
		log.Fatal(err)
		return false
	} else {
		return true
	}
}

func GenerateRandomString() string {

	lengthString := 10
	rand.Seed(time.Now().UnixNano())
	randomString := RandStringBytes(lengthString)
	return randomString

}

func GeneratePostID() string {
	lengthString := 7
	rand.Seed(time.Now().UnixNano())
	randomString := RandStringBytes(lengthString)
	return randomString
}

func GenerateCommentID() string {
	lengthString := 7
	rand.Seed(time.Now().UnixNano())
	randomString := RandStringBytes(lengthString)
	return randomString
}

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
