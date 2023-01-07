package main

import (
	"fmt"
	"log"
	"net/http"
)

func erreur(w http.ResponseWriter, r *http.Request) {

}

func main() {
	fileServer := http.FileServer(http.Dir("../assets"))
	http.Handle("../assets/", http.StripPrefix("../assets/", fileServer))

	http.HandleFunc("/", login)
	http.HandleFunc("/register", register)
	http.HandleFunc("/home", home)

	http.HandleFunc("/error", erreur)

	fmt.Println("Starting server at port: 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}

}

func login(w http.ResponseWriter, r *http.Request) {

}

func register(w http.ResponseWriter, r *http.Request) {

}

func home(w http.ResponseWriter, r *http.Request) {

}
