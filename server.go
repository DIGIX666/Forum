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
var user structure.UserAccount
var like structure.Likes
var dislike structure.Dislikes
var userComment structure.Comment
var Posts structure.Post
var uAccount []structure.UserAccount

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
	user.Post = data.HomeFeed()
	defer data.Db.Close()

	fileServer := http.FileServer(http.Dir("./assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fileServer))

	// Create a limiter with the maximum rate of 5 requests per minute.
	lmt := tollbooth.NewLimiter(100, &limiter.ExpirableOptions{DefaultExpirationTTL: time.Minute})

	// Use the limiter as middleware for the "/" handler
	http.Handle("/", tollbooth.LimitFuncHandler(lmt, home))
	http.Handle("/profil", tollbooth.LimitFuncHandler(lmt, profil))
	http.Handle("/comment", tollbooth.LimitFuncHandler(lmt, comment))
	http.Handle("/login", tollbooth.LimitFuncHandler(lmt, login))
	http.Handle("/register", tollbooth.LimitFuncHandler(lmt, register))
	//http.Handle("/userAccount", tollbooth.LimitFuncHandler(lmt, userAccount))

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

/***************************** FUNCTION LOGIN *****************************/

func login(w http.ResponseWriter, r *http.Request) {

	if r.FormValue("login") != "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	if r.FormValue("code") != "" {

		code := r.FormValue("code")

		checkGoogleUserLogged, uName, uEmail, _ := function.GoogleAuthLog(code)
		fmt.Printf("checkGoogleUserLogged: %v\n", checkGoogleUserLogged)
		checkGitHubUserLoogged, GitHub_UserName, _, _ := function.GitHubLog(code)
		fmt.Printf("checkGitHubUserLoogged: %v\n", checkGitHubUserLoogged)

		if checkGoogleUserLogged {
			uuidGenerated, _ := uuid.NewV4()
			uuidUser := uuidGenerated.String()
			cookie := http.Cookie{

				Value:  uuidUser,
				Name:   "session",
				MaxAge: 10000,
			}
			http.SetCookie(w, &cookie)

			user.Connected = true
			Posts.Connected = true
			userComment.Connected = true

			data.SetGoogleUserUUID(uEmail)
			dataBase.AddSession(uName, uuidUser, cookie.Value)
			http.Redirect(w, r, "/profil", http.StatusFound)
			return
		}
		if checkGitHubUserLoogged {

			uuidGenerated, _ := uuid.NewV4()
			uuidGithubUser := uuidGenerated.String()
			cookie := http.Cookie{

				Value:  uuidGithubUser,
				Name:   "session",
				MaxAge: 10000,
			}
			http.SetCookie(w, &cookie)
			user.Connected = true
			Posts.Connected = true
			userComment.Connected = true
			dataBase.AddSession(GitHub_UserName, uuidGithubUser, cookie.Value)

			http.Redirect(w, r, "/profil", http.StatusFound)
			return
		}
		http.Redirect(w, r, "/register", http.StatusFound)
		return

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

				}

				cookie := http.Cookie{
					Value:  uuidUser,
					Name:   "session",
					MaxAge: 10000,
				}
				http.SetCookie(w, &cookie)

				var uName, uEmail, uPassword string
				var uAdmin bool
				var uImage string
				err := data.Db.QueryRow("SELECT name, image, email, password, admin FROM users WHERE email = ?", email).Scan(&uName, &uImage, &uEmail, &uPassword, &uAdmin)
				if err != nil {
					log.Println("Erreur dans la selection des parametres utilisateur dans la fonction login: ")
					log.Fatal(err)
				}
				user.Connected = true
				data.AddSession(uName, userSession, cookie.Value)

				Posts.Connected = true
				userComment.Connected = true

				http.Redirect(w, r, "/", http.StatusFound)
				return

			} else {
				http.Redirect(w, r, "/register", http.StatusFound)
				return
			}

		} else {
			fmt.Println("email empty && password empty!")
			return
		}
	} else if r.Method == "GET" {
		t := template.New("login")
		t = template.Must(t.ParseFiles("./assets/login.html"))
		err := t.ExecuteTemplate(w, "login", nil)
		if err != nil {
			log.Fatal(err)
			return
		}
		return

	}
}

/***************************** FUNCTION LOGOUT *****************************/
// func logout(w http.ResponseWriter, r *http.Request) {
// 	// Supprime les informations de session de l'utilisateur
// 	session, _ := store.Get(r, "session-name")
// 	session.Options.MaxAge = -1
// 	session.Save(r, w)

// 	// Redirige l'utilisateur vers la page de connexion
// 	http.Redirect(w, r, "/home", http.StatusFound)
// }

// func logout(w http.ResponseWriter, r *http.Request) {
// 	cookie, err := r.Cookie("session")
// 	if err != nil {
// 		http.Redirect(w, r, "/profil", http.StatusFound)
// 		return
// 	}

// 	dataBase.DeleteSession(cookie.Value)

// 	cookie = &http.Cookie{
// 		Name:   "session",
// 		Value:  "",
// 		MaxAge: -1,
// 	}
// 	http.SetCookie(w, cookie)
// 	http.Redirect(w, r, "/profil", http.StatusFound)
// }

/*************************** FUNCTION REGISTER **********************************/

func register(w http.ResponseWriter, r *http.Request) {

	//gitHub client Secret: d01537f316e411dbc710369e9f907f5b8a71cc9d

	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}

	if r.FormValue("code") != "" {
		fmt.Printf("Code receive: %v\n", r.FormValue("code"))
		checkGitHub_User_Registered, _, userGitHubName := function.GitHubRegister(r.FormValue("code"))
		hashPassword := script.GenerateHash(script.GenerateRandomString())

		checkGoogleUserRegistered, googleUserEmail, userGoogleName := function.GoogleAuthRegister(r.FormValue("code"), hashPassword)
		fmt.Printf("checkGoogleUserRegistered: %v\n", checkGoogleUserRegistered)

		if checkGitHub_User_Registered || userGitHubName != "" {

			uuidGitHubUser := data.SetGitHubUUID(userGitHubName)
			cookie := http.Cookie{
				Value:  uuidGitHubUser,
				Name:   "session",
				MaxAge: 100000,
			}

			http.SetCookie(w, &cookie)
			data.AddSession(userGitHubName, uuidGitHubUser, cookie.Value)
			user.Connected = true
			Posts.Connected = true
			userComment.Connected = true
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		if checkGoogleUserRegistered && googleUserEmail != "" {
			uuidGoogleUser := data.SetGoogleUserUUID(googleUserEmail)
			cookie := http.Cookie{
				Value:  uuidGoogleUser,
				Name:   "session",
				MaxAge: 100000,
			}
			http.SetCookie(w, &cookie)

			data.AddSession(userGoogleName, uuidGoogleUser, cookie.Value)

			user.Connected = true
			Posts.Connected = true
			userComment.Connected = true

			http.Redirect(w, r, "/", http.StatusFound)
			return
		} else if googleUserEmail != "" {
			http.Redirect(w, r, "/login", http.StatusFound)
			return

		} else {
			fmt.Println("Error register Google User !")
			return
		}
	} else {
		fmt.Println("Receive no code !")

		if r.Method == "GET" {
			t := template.New("register")
			t = template.Must(t.ParseFiles("./assets/register.html"))
			err := t.ExecuteTemplate(w, "register", nil)
			if err != nil {
				log.Fatal(err)
			}
		}

		var email string
		var password string
		email = r.FormValue("email_confirm")
		password = r.FormValue("password_confirm")

		hashPassword := script.GenerateHash(password)

		if email != "" && password != "" {

			checkRegister := dataBase.DataBaseRegister(email, hashPassword)
			if checkRegister {
				if r.Method == "POST" {
					uuidGenerated, _ := uuid.NewV4()
					uuidUser := uuidGenerated.String()
					cookie := http.Cookie{
						Value:  uuidUser,
						Name:   "session",
						MaxAge: 100000,
					}
					user.Connected = true
					Posts.Connected = true
					userComment.Connected = true
					http.SetCookie(w, &cookie)
					data.AddSession("none", uuidUser, cookie.Value)
					_, err := data.Db.Exec("UPDATE users SET UUID = ? WHERE email = ?", uuidUser, email)
					if err != nil {
						fmt.Println("Erreur modifie la valeur de UUID de la fonction register:")
						log.Fatal(err)
						return
					}
					http.Redirect(w, r, "/", http.StatusFound)
					return
				}
			} else {
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}
		}
	}
}

/*************************** FUNCTION HOME **********************************/

var posts []structure.Post

func preappendPost(c structure.Post) []structure.Post {
	posts = append(posts, structure.Post{})
	copy(posts[1:], posts)
	posts[0] = c
	return posts
}

func home(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}

	//var count int
	var profil map[string]string

	temp, err := template.ParseFiles("./assets/Home/home.html")
	if err != nil {
		log.Println("Error parsing template:", err)
		return
	}

	if user.Connected {
		_, err = r.Cookie("session")
		if err != nil {
			user.Connected = false
			for _, v := range user.Post {
				v.Connected = false
			}
			data.DeleteSession(user.Name)
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		profil = data.GetUserProfil()
		user.Connected = true
		for _, v := range user.Post {
			v.Connected = true
		}

		user.Name = profil["name"]
		user.Email = profil["email"]
		user.Image = profil["userImage"]
		user.UUID = profil["uuid"]
		if profil["admin"] == "true" {
			user.Admin = true
		} else {
			user.Admin = false
		}

		picture := r.FormValue("picture")
		message := r.FormValue("message")
		postid := script.GeneratePostID()

		if message != "" {
			currentTime := time.Now().Format("15:04  2-Janv-2006")
			user.Post = preappendPost(structure.Post{
				PostID:          postid,
				Name:            user.Name,
				Message:         message,
				DateTime:        currentTime,
				Picture:         picture,
				NumberOfComment: data.NumberOfComment(postid),
				Connected:       user.Connected,
				UserImage:       user.Image,
			})

			fmt.Printf("data.NumberOfComment(postid): %v\n", data.NumberOfComment(postid))

			//Put the message in the dataBase
			dataBase.UserPost(user.Name, message, postid, user.Image, currentTime, picture)
		}
	}

	err = temp.ExecuteTemplate(w, "home", user)
	if err != nil {
		log.Fatal(err)
	}

}

/*************************** FUNCTION PROFIL **********************************/
func profil(w http.ResponseWriter, r *http.Request) {

	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}

	if r.FormValue("logout") != "" {

		cookie := http.Cookie{
			Value:  "",
			Name:   "session",
			MaxAge: -1,
		}
		http.SetCookie(w, &cookie)
		data.DeleteSession(user.Name)
		user.Connected = false
		Posts.Connected = false
		userComment.Connected = false
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	for _, v := range user.Post {
		v.Connected = true
	}

	temp, err := template.ParseFiles("./assets/Profil/profil.html")
	if err != nil {
		log.Println("Error parsing template:", err)
		return
	}

	if user.Connected {
		_, err = r.Cookie("session")
		if err != nil {
			user.Connected = false
			for _, v := range user.Post {
				v.Connected = false
			}
			data.DeleteSession(user.Name)
			Posts.Connected = false
			userComment.Connected = false
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
	}

	profil := data.GetUserProfil()

	user.Name = profil["name"]
	user.Email = profil["email"]
	user.Image = profil["userImage"]
	user.UUID = profil["uuid"]
	if profil["admin"] == "true" {
		user.Admin = true
	} else {
		user.Admin = false
	}

	message := r.FormValue("message")
	picture := r.FormValue("picture")
	postid := script.GeneratePostID()
	if message != "" {
		currentTime := time.Now().Format("15:04  2-Janv-2006")
		user.Post = preappendPost(structure.Post{
			PostID:   postid,
			Name:     user.Name,
			Message:  message,
			DateTime: currentTime,
			Picture:  picture,
		})
		//Put the message in the dataBase
		dataBase.UserPost(user.Name, message, postid, user.Image, currentTime, picture)
	}

	if err = temp.ExecuteTemplate(w, "profil", user); err != nil {
		log.Println("Error executing template:", err)
		return
	}

}

/*************************** FUNCTION COMMENT **********************************/

func comment(w http.ResponseWriter, r *http.Request) {
	temp, err := template.ParseFiles("./assets/Commentaire/comment.html")
	if err != nil {
		log.Println("Error parsing template:", err)
		return
	}

	if r.Method == "POST" {

		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}

		message := r.FormValue("message")
		postID := r.FormValue("Post_values")
		fmt.Printf("postID: %v\n", postID)
		//postid, _ := strconv.Atoi(postID)
		//fmt.Printf("postid: %v\n", postid)

		currentTime := time.Now().Format("15:04  2-Jan-2006")

		//Put the message in the dataBase
		v := dataBase.UserComment(user.Name, message, script.GenerateCommentID(), currentTime, postID)
		fmt.Printf("User Comment: %v\n", v)

		http.Redirect(w, r, "/comment?postid="+postID, http.StatusSeeOther)

	} else if r.Method == "GET" {

		postid := r.URL.Query().Get("postid")

		if err := temp.ExecuteTemplate(w, "comment", map[string]any{
			"PostID":   postid,
			"Comments": data.GetPostComment(postid),
		}); err != nil {
			log.Println("Error executing template:", err)
			return
		}

	}
}
