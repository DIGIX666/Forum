package main

import (
	data "Forum/data"
	script "Forum/scripts"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"text/template"
	"time"

	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
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

	// Create a limiter with the maximum rate of 5 requests per minute.
	lmt := tollbooth.NewLimiter(5, &limiter.ExpirableOptions{DefaultExpirationTTL: time.Minute})

	// Use the limiter as middleware for the "/" handler
	http.Handle("/", tollbooth.LimitFuncHandler(lmt, home))
	http.Handle("/login", tollbooth.LimitFuncHandler(lmt, login))
	http.Handle("/register", tollbooth.LimitFuncHandler(lmt, register))

	http.HandleFunc("/error", erreur)

	// configuration TLS en utilisant les certificats générés
	config := &tls.Config{
		MinVersion:               tls.VersionTLS12,
		CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		},
	}

	// configuration du serveur HTTP
	server := &http.Server{
		Addr:         ":8080",
		TLSConfig:    config,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
	}

	fmt.Println("Starting server at port: 8080")
	err := server.ListenAndServeTLS("Key/server.crt", "Key/server.key")
	if err != nil {
		if err != nil {
			log.Fatal(err)
		}
	}
}

func login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}
	t := template.New("login")
	t = template.Must(t.ParseFiles("./assets/login.html"))
	err := t.ExecuteTemplate(w, "login", nil)
	if err != nil {
		log.Fatal(err)
	}

	email := r.FormValue("email")
	password := r.FormValue("password")
	uuidUser := script.GenerateRandomString()
	if email != "" && password != "" {
		if data.DataBaseLogin(email, password, uuidUser) {
			http.HandleFunc("/"+uuidUser, userAccount)
		} else {
			fmt.Println("mot de passe pas bon !!")
		}
	}
}

func register(w http.ResponseWriter, r *http.Request) {

	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}

	t := template.New("register")
	t = template.Must(t.ParseFiles("./assets/register.html"))
	err := t.ExecuteTemplate(w, "register", nil)
	if err != nil {
		log.Fatal(err)
	}
	var email string
	var password string
	email = r.FormValue("email_confirm")
	password = r.FormValue("password_confirm")
	/*fmt.Printf("email: %v\n", email)
	fmt.Printf("password: %v\n", password)
	fmt.Printf("length email: %v\n", len(email))
	fmt.Printf("length password: %v\n", len(password))*/

	hashPassword := script.GenerateHash(password)

	//fmt.Printf("email: %v\n", email)
	//fmt.Printf("hashPassword: %v\n", hashPassword)

	compare := script.ComparePassword(hashPassword, password)
	fmt.Printf("compare: %v\n", compare)

	if email != "" && password != "" {
		data.DataBaseRegister(email, hashPassword)
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}
	temp := template.New("home")
	temp = template.Must(temp.ParseFiles("./assets/home.html"))
	err := temp.ExecuteTemplate(w, "home", nil)
	if err != nil {
		log.Fatal(err)
	}
}

//

func userAccount(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}

	t := template.New("userAccount")
	t = template.Must(t.ParseFiles("./assets/userAccount.html"))
	err := t.ExecuteTemplate(w, "userAccount", nil)
	if err != nil {
		log.Fatal(err)
	}
}
