package main

import (
	structure "Forum/Struct"
	"Forum/data"
	dataBase "Forum/data"
	function "Forum/functions"
	script "Forum/scripts"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"

	"text/template"
	"time"

	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	"github.com/gofrs/uuid"
)

/****************************** FUNCTION ERREUR *******************************/

func erreur(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" && r.URL.Path != "/register" && r.URL.Path != "/home" && r.URL.Path != "/error" && r.URL.Path != "/userAccount" {
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

	/*if r.Method != "POST" {
		http.Error(w, "Method is not supported", http.StatusNotFound)
		return
	}*/
}

/****************************** FUNCTION MAIN ********************************/

func main() {
	dataBase.CreateDataBase()
	defer data.Db.Close()

	fileServer := http.FileServer(http.Dir("./assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fileServer))

	// Create a limiter with the maximum rate of 5 requests per minute.
	lmt := tollbooth.NewLimiter(100, &limiter.ExpirableOptions{DefaultExpirationTTL: time.Minute})

	// Use the limiter as middleware for the "/" handler
	http.Handle("/", tollbooth.LimitFuncHandler(lmt, home))
	http.Handle("/profil", tollbooth.LimitFuncHandler(lmt, profil))
	http.Handle("/login", tollbooth.LimitFuncHandler(lmt, login))
	http.Handle("/register", tollbooth.LimitFuncHandler(lmt, register))
	http.Handle("/userAccount", tollbooth.LimitFuncHandler(lmt, userAccount))

	http.Handle("/error", tollbooth.LimitFuncHandler(lmt, erreur))

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

var uAccount []structure.UserAccount

/***************************** FUNCTION LOGIN *****************************/

func login(w http.ResponseWriter, r *http.Request) {

	if r.FormValue("code") != "" {

		code := r.FormValue("code")
		checkUserLogged, uuidUser := function.GoogleAuthLog(code)
		if checkUserLogged {
			http.Redirect(w, r, "/profil/"+uuidUser, http.StatusFound)
		} else {
			fmt.Println("Error to logIn the Google User")
			return
		}

	}

	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}

		email := r.FormValue("email")
		password := r.FormValue("password")
		uuidGenerated, _ := uuid.NewV4()
		uuidUser := uuidGenerated.String()
		fmt.Printf("uuidUser: %v\n", uuidUser)

		if email != "" && password != "" {
			checkLogin := dataBase.CheckUserLogin(email, password, uuidUser)
			if checkLogin {
				uAccount = append(uAccount, structure.UserAccount{
					UUID: uuidUser,
				})
				var userSession string
				for range uAccount {

					err := data.Db.QueryRow("SELECT uuid FROM users WHERE email = ?", email).Scan(&userSession)
					if err != nil {
						log.Println("Erreur dans la QueryRow dans la fonction login pour userSession")
						log.Fatal(err)
					}
					//fmt.Println("l'UUID: " + v.UUID)

				}

				cookie := http.Cookie{
					Expires: time.Now().Add(time.Hour),
					Value:   uuidUser,
					Name:    "session",
				}
				http.SetCookie(w, &cookie)

				var uName, uEmail, uPassword string
				var uAdmin bool
				var uImage string
				err := data.Db.QueryRow("SELECT name, image, email, password, admin FROM users WHERE email = ?", email).Scan(&uName, &uImage, &uEmail, &uPassword, &uAdmin)
				if err != nil {
					log.Println(" Erreur dans la selection des parametres utilisateur dans la fonction login: ")
					log.Fatal(err)
				}

				uAccount = append(uAccount, structure.UserAccount{
					Name:     uName,
					Image:    uImage,
					Email:    uEmail,
					Password: uPassword,
					Admin:    uAdmin,
				})
				data.AddSession(uName, userSession, cookie.Value)
				for _, v := range uAccount {
					t := template.New("userAccount")
					t = template.Must(t.ParseFiles("./assets/userAccount.html"))
					err = t.ExecuteTemplate(w, "userAccount", v)
					if err != nil {
						log.Fatal(err)
					}
				}

			}

		} else {
			fmt.Println("email empty && password empty!")
		}
	} else if r.Method == "GET" {
		t := template.New("login")
		t = template.Must(t.ParseFiles("./assets/login.html"))
		err := t.ExecuteTemplate(w, "login", nil)
		if err != nil {
			log.Fatal(err)
		}

	}

}

/*************************** FUNCTION REGISTER **********************************/

func register(w http.ResponseWriter, r *http.Request) {

	//gitHub client Secret: d01537f316e411dbc710369e9f907f5b8a71cc9d

	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}

	if r.FormValue("code") != "" {
		checkUserRegistered, uuidUser, userName := function.GitHubRegister(r.FormValue("code"))

		if checkUserRegistered {
			cookie := http.Cookie{
				Expires: time.Now().Add(time.Hour),
				Value:   uuidUser,
				Name:    "session",
			}
			http.SetCookie(w, &cookie)
			//fmt.Printf("uuidUser: %v\n", uuidUser)
			//fmt.Printf("userName: %s\n", userName)
			//fmt.Printf("cookie.Value: %v\n", cookie.Value)

			data.AddSession(userName, uuidUser, cookie.Value)
			http.Redirect(w, r, "/profil/"+uuidUser, http.StatusFound)

		} else {
			fmt.Println("Error to register the GitHub user !")
			return
		}
		return
	} else {
		fmt.Println("Receive nothing from github")
	}

	if r.FormValue("code") != "" {
		code := r.FormValue("code")
		hashPassword := script.GenerateHash(script.GenerateRandomString())

		checkUserRegistered, uuidUser, userName := function.GoogleAuthRegister(code, hashPassword)

		if checkUserRegistered {

			cookie := http.Cookie{
				Expires: time.Now().Add(time.Hour),
				Value:   uuidUser,
				Name:    "session",
			}
			http.SetCookie(w, &cookie)
			data.AddSession(userName, uuidUser, cookie.Value)

			http.Redirect(w, r, "/profil/"+uuidUser, http.StatusFound)
		} else {
			fmt.Println("Error to Register the Google User !")
			return
		}

	} else {

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

		hashPassword := script.GenerateHash(password)

		//fmt.Printf("email: %v\n", email)
		//fmt.Printf("hashPassword: %v\n", hashPassword)

		compare := script.ComparePassword(hashPassword, password)
		fmt.Printf("compare: %v\n", compare)

		if email != "" && password != "" {
			checkRegister := dataBase.DataBaseRegister(email, hashPassword)

			if checkRegister {
				uAccount = append(uAccount, structure.UserAccount{

					Email:    email,
					Password: password,
				})

			} else {
				fmt.Println("problem to Register ! maybe email already exist !")

			}
		}

	}

}

/*************************** FUNCTION HOME **********************************/

var posts []structure.Post

//var templatePost []structure.Post

func preappendPost(c structure.Post) {
	posts = append(posts, structure.Post{})
	copy(posts[1:], posts)
	posts[0] = c
}

func home(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}
	/*temp, err := template.ParseFiles("./assets/Home/home.html")
	if err != nil {
		log.Println("Error parsing template:", err)
		return
	}

	//var user structure.UserAccount

	/*name := r.FormValue("name")
	//name := "Mirio Togata"
	var userIdDB int

	var profil structure.UserAccount

	err = data.Db.QueryRow("SELECT id,image, email, UUID, admin, password FROM users WHERE name = ?", name).Scan(&userIdDB, &profil.Image, &profil.Email, &profil.UUID, &profil.Admin, &profil.Password)
	if err != nil {
		fmt.Println("Error when Selecting user profil from userForum.db")
		log.Fatal(err)
	}

	message := r.FormValue("message")
	if message != "" {
		currentTime := time.Now().Format("15:04  11.janv.2006")
		preappendPost(structure.Post{
			PostID:   script.GeneratePostID(),
			Name:     name,
			Message:  message,
			DateTime: currentTime,
		})

		//Put the message in the dataBase
		fmt.Printf("profil.Name: %v", profil.Name)
		dataBase.UserPost(profil.Name, message, script.GeneratePostID(), currentTime)

	}

	if err := temp.ExecuteTemplate(w, "home", posts); err != nil {
		log.Println("Error executing template:", err)
		return
	}*/

}

/*************************** FUNCTION PROFIL **********************************/
func profil(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}

	temp, err := template.ParseFiles("./assets/Profil/profil.html")
	if err != nil {
		log.Println("Error parsing template:", err)
		return
	}

	name := r.FormValue("name")
	message := r.FormValue("message")
	if message != "" {
		currentTime := time.Now().Format("15:04  11.janv.2006")
		preappendPost(structure.Post{Name: name, Message: message, DateTime: currentTime})
	}

	for _, v := range posts {
		fmt.Printf("v.DateTime: %v\n", v.DateTime)
		fmt.Printf("v.Message: %v\n", v.Message)
	}
	if err := temp.ExecuteTemplate(w, "profil", posts); err != nil {
		log.Println("Error executing template:", err)
		return
	}
}

/*************************** FUNCTION USER ACCOUNT **********************************/

func userAccount(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}

	var userAccount []structure.UserAccount

	t := template.New("userAccount")
	t = template.Must(t.ParseFiles("./assets/userAccount.html"))

	for _, v := range userAccount {
		err := t.ExecuteTemplate(w, "userAccount", v)
		if err != nil {
			log.Fatal(err)
		}
	}

}
