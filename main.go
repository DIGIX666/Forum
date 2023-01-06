package main

import (
	data "Forum/data"
	"fmt"
	"log"
)

func main() {
	email := identificationEmail()
	password := identificationPassword()
	data.DataBase(email, password)

}

func identificationEmail() string {
	fmt.Println("enter you email:")
	var email string

	_, err := fmt.Scanln(&email)
	if err != nil {
		log.Fatal(err)
	}
	return email
}

func identificationPassword() string {
	fmt.Println("enter your password:")
	var password string

	_, err := fmt.Scanln(&password)
	if err != nil {
		log.Fatal(err)
	}
	return password
}
