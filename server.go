package main

import (
	script "Forum/scripts"
	"fmt"
	"log"
	"net/http"
	"text/template"
)

func erreur(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" && r.URL.Path != "/register" && r.URL.Path != "/home" && r.URL.Path != "/error" {
		http.Error(w, "404 not found", http.StatusNotFound)
		return
	}

	_, err := template.ParseFiles("./assets/error.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500: Internal Server Error"))
		log.Println((http.StatusInternalServerError))
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method is not supported", http.StatusNotFound)
		return
	}

}

func main() {
	fileServer := http.FileServer(http.Dir("./assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fileServer))

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
	t := template.New("login")
	t = template.Must(t.ParseFiles("./assets/login.html"))
	err := t.ExecuteTemplate(w, "login", nil)
	if err != nil {
		log.Fatal(err)
	}

}

func register(w http.ResponseWriter, r *http.Request) {

	t := template.New("register")
	t = template.Must(t.ParseFiles("./assets/register.html"))
	err := t.ExecuteTemplate(w, "register", nil)
	if err != nil {
		log.Fatal(err)
	}

	email := r.FormValue("email_confirm")
	password := r.FormValue("password_confirm")
	fmt.Printf("email: %v\n", email)
	fmt.Printf("password: %v\n", password)

	//mixedPassword := password + script.GenerateRandomString()

	hashPassword := script.GenerateHash(password)

	fmt.Printf("email: %v\n", email)
	fmt.Printf("hashPassword: %v\n", hashPassword)

	compare := script.ComparePassword(hashPassword, password)
	fmt.Printf("compare: %v\n", compare)

	//data.DataBase(email, password)

}

func home(w http.ResponseWriter, r *http.Request) {
	//fonction a compléter
}
