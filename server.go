package main

import (
	structure "Forum/backend/infrastructures/Struct"
	"Forum/backend/adapters/data"
	dataBase "Forum/backend/adapters/data"
	function "Forum/backend/domains"
	script "Forum/backend/infrastructures/cryptage"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	"github.com/gofrs/uuid"
	"github.com/joho/godotenv"
)

var user structure.UserAccount
var userComment structure.Comment
var Posts structure.Post
var uAccount []structure.UserAccount
var homefeed []structure.HomeFeedPost
var comments []structure.Comment
var adminfeed []structure.AdminFeedPost

// Fonction for generated CSRF token
func GenerateCSRFToken() (string, error) {
	token := make([]byte, 32)
	_, err := rand.Read(token)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(token), nil
}

// Middleware for token verification CSRF
func csrfMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			sessionToken := r.FormValue("csrf_token")
			cookie, err := r.Cookie("csrf_token")
			if err != nil || sessionToken != cookie.Value {
				http.Error(w, "Invalid CSRF token", http.StatusForbidden)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

/****************************** FUNCTION ERREUR *******************************/

func erreur(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" && r.URL.Path != "/register" && r.URL.Path != "/home" && r.URL.Path != "/error" && r.URL.Path != "/userAccount" {
		http.Error(w, "404 not found", http.StatusNotFound)
		return
	}

	_, err := template.ParseFiles("./frontend/error.html")
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
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Utilisation des variables d'environnement

	dataBase.CreateDataBase()
	homefeed = data.HomeFeedPost()
	data.AddingAdminUser()
	uAccount = data.GetAllUsers()
	defer data.Db.Close()

	fileServer := http.FileServer(http.Dir("./frontend"))
	http.Handle("/frontend/", http.StripPrefix("/frontend/", fileServer))

	// Create a limiter with the maximum rate of 5 requests per minute.
	lmt := tollbooth.NewLimiter(100, &limiter.ExpirableOptions{DefaultExpirationTTL: time.Minute})

	// Use the limiter as middleware for the "/" handler
	http.Handle("/", tollbooth.LimitFuncHandler(lmt, home))
	http.Handle("/moderateur", tollbooth.LimitFuncHandler(lmt, moderateur))
	http.Handle("/admin", tollbooth.LimitFuncHandler(lmt, admin))
	http.Handle("/categorie1", tollbooth.LimitFuncHandler(lmt, categorie1))
	http.Handle("/categorie2", tollbooth.LimitFuncHandler(lmt, categorie2))
	http.Handle("/categorie3", tollbooth.LimitFuncHandler(lmt, categorie3))
	http.Handle("/profil", tollbooth.LimitFuncHandler(lmt, profil))
	http.Handle("/comment", tollbooth.LimitFuncHandler(lmt, comment))
	http.Handle("/login", tollbooth.LimitFuncHandler(lmt, login))
	http.Handle("/register", csrfMiddleware(tollbooth.LimitFuncHandler(lmt, register)))
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
		ReadTimeout:  5 * time.Second,
	}

	fmt.Println("Starting server at port: 8080")
	err = server.ListenAndServeTLS("./backend/infrastructures/Key/server.crt", "./backend/infrastructures/Key/server.key")
	if err != nil {
		log.Fatal(err)
	}
}

/***************************** FUNCTION LOGIN *****************************/

func login(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}

		// Vérification du token CSRF
		sessionToken := r.FormValue("csrf_token")
		cookie, err := r.Cookie("csrf_token")
		if err != nil || sessionToken != cookie.Value {
			http.Error(w, "Invalid CSRF token", http.StatusForbidden)
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
					MaxAge: 7200,
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

				http.Redirect(w, r, "/profil", http.StatusFound)
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
		// Generate a new CSRF token
		csrfToken, err := GenerateCSRFToken()
		if err != nil {
			http.Error(w, "Erreur lors de la génération du token CSRF", http.StatusInternalServerError)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:     "csrf_token",
			Value:    csrfToken,
			HttpOnly: true,
			Path:     "/",
			MaxAge:   3600,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		})

		t := template.New("login")
		t = template.Must(t.ParseFiles("./frontend/login.html"))
		// Give the CSRF token to the template
		err = t.ExecuteTemplate(w, "login", map[string]interface{}{
			"CSRFToken": csrfToken,
		})
		if err != nil {
			log.Fatal(err)
			return
		}
		return
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

		checkGitHub_User_Registered, _, userGitHubName := function.GitHubRegister(r.FormValue("code"))
		hashPassword := script.GenerateHash(script.GenerateRandomString())

		checkGoogleUserRegistered, googleUserEmail, userGoogleName := function.GoogleAuthRegister(r.FormValue("code"), hashPassword)
		function.DiscordAuthRegister(r.FormValue("code"), hashPassword)

		if checkGitHub_User_Registered || userGitHubName != "" {

			uuidGitHubUser := data.SetGitHubUUID(userGitHubName)
			cookie := http.Cookie{
				Value:  uuidGitHubUser,
				Name:   "session",
				MaxAge: 120,
			}

			http.SetCookie(w, &cookie)
			data.AddSession(userGitHubName, uuidGitHubUser, cookie.Value)
			user.Connected = true
			Posts.Connected = true
			userComment.Connected = true
			http.Redirect(w, r, "/profil", http.StatusFound)
			return
		}

		if checkGoogleUserRegistered && googleUserEmail != "" {
			uuidGoogleUser := data.SetGoogleUserUUID(googleUserEmail)
			cookie := http.Cookie{
				Value:  uuidGoogleUser,
				Name:   "session",
				MaxAge: 120,
			}
			http.SetCookie(w, &cookie)

			data.AddSession(userGoogleName, uuidGoogleUser, cookie.Value)

			user.Connected = true
			Posts.Connected = true
			userComment.Connected = true

			http.Redirect(w, r, "/profil", http.StatusFound)
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
			csrfToken, err := GenerateCSRFToken()
			if err != nil {
				http.Error(w, "Erreur lors de la génération du token CSRF", http.StatusInternalServerError)
				return
			}
			// Stocker le token dans un cookie
			http.SetCookie(w, &http.Cookie{
				Name:     "csrf_token",
				Value:    csrfToken,
				HttpOnly: true,
				Path:     "/",
				MaxAge:   3600,
				Secure:   true,
				SameSite: http.SameSiteStrictMode,
			})

			t := template.New("register")
			t = template.Must(t.ParseFiles("./frontend/register.html"))
			err = t.ExecuteTemplate(w, "register", map[string]interface{}{
				"CSRFToken": csrfToken,
			})
			if err != nil {
				log.Fatal(err)
			}
			return
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
						MaxAge: 190,
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
					http.Redirect(w, r, "/profil", http.StatusFound)
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

	if len(uAccount) < data.LenUser() {
		uAccount = data.GetAllUsers()
	}

	var notification []structure.Notification

	if len(uAccount) > 0 && user.Connected {

		profil := data.GetUserProfil()
		user.Name = profil["name"]
		user.Email = profil["email"]
		user.Image = profil["userImage"]
		user.UUID = profil["uuid"]
		if profil["admin"] == "true" {
			user.Admin = true
		} else if profil["admin"] == "false" {
			user.Admin = false
		}
		if profil["moderateur"] == "true" {
			user.Moderateur = true
		} else if profil["moderateur"] == "false" {
			user.Moderateur = false
		}
	}
	var imageSRC string
	if r.Method == "POST" {

		message := r.FormValue("message")
		Posts.Categories = r.FormValue("categories")
		Posts.Categories2 = r.FormValue("categories2")
		// notif = r.FormValue("selectnone")
		// user.Admin = r.FormValue("admin") == "true"

		fmt.Printf("Posts.Categories: %v\n", Posts.Categories)
		if message != "" && user.Connected {
			postid := script.GeneratePostID()
			currentTime := time.Now().Format("15:04  2-Janv-2006")

			user.Post = preappendPost(structure.Post{
				PostID:      postid,
				Name:        user.Name,
				Message:     message,
				DateTime:    currentTime,
				UserImage:   user.Image,
				CountCom:    data.LenUserComment(postid),
				Categories:  Posts.Categories,
				Categories2: Posts.Categories2,
				Admin:       user.Admin,
				Connected:   true,
			})
			file, header, err := r.FormFile("myFile")

			imageName := ""
			maxImageSize := 20 * 1024 * 1024 // 20Mo en octets
			if err != nil {
				if err != http.ErrMissingFile {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				//Put the message in the dataBase
				dataBase.UserPost(user.Name, message, postid, user.Image, currentTime, imageName, Posts.Count, Posts.CountDis, Posts.CountCom, Posts.Categories, Posts.Categories2)
				homefeed = dataBase.HomeFeedPost()

			} else {
				imageName = header.Filename
				fmt.Printf("Uploaded File: %+v\n", header.Filename)

				// read all of the contents of our uploaded file into a
				// byte array
				fileBytes, err := ioutil.ReadAll(file)

				if err != nil {
					fmt.Println(err)
				}

				imageSRC = "./frontend/upload-image/" + imageName

				err = ioutil.WriteFile(imageSRC, fileBytes, 0o666)
				if err != nil {
					log.Fatal(err)
				}

				fileIMAGE, err := os.Open(imageSRC)
				if err != nil {
					fmt.Println(err)
				}
				fileStat, err := fileIMAGE.Stat()
				if err != nil {
					fmt.Println(err)
				}
				if fileStat.Size() > int64(maxImageSize) {
					os.Remove(imageSRC)
				} else {
					_, user.Post = dataBase.UserPost(user.Name, message, postid, user.Image, currentTime, imageSRC, Posts.Count, Posts.CountDis, Posts.CountCom, Posts.Categories, Posts.Categories2)
					homefeed = dataBase.HomeFeedPost()
					file.Close()

				}
			}
		}

		if r.FormValue("like") != "" && user.Connected {
			currentTime := time.Now().Format("15:04  2-Janv-2006")

			postid := r.FormValue("like")
			countLike := 0
			row := data.Db.QueryRow("SELECT COUNT (*) FROM likes WHERE username = ? AND post_id = ?", user.Name, postid)
			err := row.Scan(&countLike)
			if err != nil {
				panic(err)
			}
			if countLike == 0 {
				dataBase.AddingCountLike(postid, user.Name, currentTime)
			}

			data.NotifLike(postid)
			homefeed = dataBase.HomeFeedPost()
		}

		if r.FormValue("dislike") != "" && user.Connected {
			currentTime := time.Now().Format("15:04  2-Janv-2006")

			postid := r.FormValue("dislike")
			countDislike := 0
			row := data.Db.QueryRow("SELECT COUNT (*) FROM dislikes WHERE username = ? AND post_id = ?", user.Name, postid)
			err := row.Scan(&countDislike)
			if err != nil {
				panic(err)
			}
			if countDislike == 0 {
				data.AddingCountDislike(postid, user.Name, currentTime)
			}

			data.NotifDisLike(postid)
			homefeed = dataBase.HomeFeedPost()
		}
	}

	temp, err := template.ParseFiles("./frontend/Home/home.html")
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
			userComment.Connected = false
			data.DeleteSession(user.Name)
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		user.Connected = true
		for _, v := range user.Post {
			v.Connected = true
		}

		if r.FormValue("delete-notif") != "" {
			data.DeleteNotif(r.FormValue("delete-notif"))

		}
		notification = data.GetUserNotif()

		homefeed = data.HomeFeedPost()

		err = temp.ExecuteTemplate(w, "home", map[string]any{
			"User":          user,
			"HomeFeed":      homefeed,
			"Notifications": notification,
		})
		if err != nil {
			log.Fatal(err)
		}

	} else {

		homefeed = data.HomeFeedPost()

		err = temp.ExecuteTemplate(w, "home", map[string]any{
			"User":     user,
			"HomeFeed": homefeed,
		})
		if err != nil {
			log.Fatal(err)
		}
	}
}

/*************************** FUNCTION PROFIL **********************************/
func profil(w http.ResponseWriter, r *http.Request) {
	var userLikeFeed []structure.UserFeedPost
	var notification []structure.Notification

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
	if profil["moderateur"] == "true" {
		user.Moderateur = true
	} else {
		user.Moderateur = false
	}

	var userHomeFeed []structure.UserFeedPost

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

		user.Connected = false
		Posts.Connected = false
		userComment.Connected = false
		data.DeleteSession(user.Name)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	for _, v := range user.Post {
		v.Connected = true
	}

	temp, err := template.ParseFiles("./frontend/Profil/profil.html")
	if err != nil {
		log.Println("Error parsing template:", err)
		return
	}

	_, err = r.Cookie("session")
	if err != nil {
		user.Connected = false
		for _, v := range user.Post {
			v.Connected = false
		}

		Posts.Connected = false
		userComment.Connected = false
		data.DeleteSession(user.Name)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	if len(userHomeFeed) < data.LenUserPost(user.Name) {
		userHomeFeed = data.ProfilFeed(user.Name)
	}

	if len(userLikeFeed) < data.LenLikeUserPost(user.Name) {
		userLikeFeed = data.ProfilLikeFeed(user.Name)
	}

	notif := r.FormValue("notif")
	if notif == "notif_moderateur" {

		data.AddingModoRequest(user.Name, user.Image, script.GeneratePostID(), time.Now().Format("15:04  2-Janv-2006"))

	}
	if r.FormValue("delete-notif") != "" {
		data.DeleteNotif(r.FormValue("delete-notif"))

	}

	notification = data.GetUserNotif()

	if err = temp.ExecuteTemplate(w, "profil", map[string]any{
		"user":          user,
		"UserPost":      userHomeFeed,
		"UserLike":      userLikeFeed,
		"Notifications": notification,
	}); err != nil {
		log.Println("Error executing template:", err)
		return

	}
}

/********************************************************************************/

/*************************** FUNCTION CATEGORIE 1 **********************************/
func categorie1(w http.ResponseWriter, r *http.Request) {

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

	// var categorie1Feed []structure.Categorie1FeedPost

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

		user.Connected = false
		Posts.Connected = false
		userComment.Connected = false
		data.DeleteSession(user.Name)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	for _, v := range user.Post {
		v.Connected = true
	}

	temp, err := template.ParseFiles("./frontend/Categories/Cat-1/cat1.html")
	if err != nil {
		log.Println("Error parsing template:", err)
		return
	}

	_, err = r.Cookie("session")
	if err != nil {
		user.Connected = false
		for _, v := range user.Post {
			v.Connected = false
		}

		Posts.Connected = false
		userComment.Connected = false
		data.DeleteSession(user.Name)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	// data.Categorie1FeedPost(user.Name)
	fmt.Printf("user.Name: %v\n", user.Name)

	if err = temp.ExecuteTemplate(w, "categorie1", map[string]any{
		"user":       user,
		"categories": data.Categorie1FeedPost(user.Name),
		// "categries2": data.Categorie1FeedPost(user.Name),
	}); err != nil {
		log.Println("Error executing template:", err)
		return
	}
	// fmt.Printf("categorie1Feed: %v\n", categorie1Feed)
}

/********************************************************************************/

/*************************** FUNCTION CATEGORIE 2 **********************************/
func categorie2(w http.ResponseWriter, r *http.Request) {

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

	// var categorie2Feed []structure.Categorie2FeedPost

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

		user.Connected = false
		Posts.Connected = false
		userComment.Connected = false
		data.DeleteSession(user.Name)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	for _, v := range user.Post {
		v.Connected = true
	}

	temp, err := template.ParseFiles("./frontend/Categories/Cat-2/cat2.html")
	if err != nil {
		log.Println("Error parsing template:", err)
		return
	}

	_, err = r.Cookie("session")
	if err != nil {
		user.Connected = false
		for _, v := range user.Post {
			v.Connected = false
		}

		Posts.Connected = false
		userComment.Connected = false
		data.DeleteSession(user.Name)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	data.Categorie2FeedPost(user.Name)
	fmt.Printf("user.Name: %v\n", user.Name)

	if err = temp.ExecuteTemplate(w, "categorie2", map[string]any{
		"user":        user,
		"categories2": data.Categorie2FeedPost(user.Name),
	}); err != nil {
		log.Println("Error executing template:", err)
		return
	}
	fmt.Printf("data.Categorie2FeedPost(): %v\n", data.Categorie2FeedPost(user.Name))
}

/********************************************************************************/

/*************************** FUNCTION CATEGORIE 3 **********************************/
func categorie3(w http.ResponseWriter, r *http.Request) {

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

	var categorie3Feed []structure.Categorie3FeedPost

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

		user.Connected = false
		Posts.Connected = false
		userComment.Connected = false
		data.DeleteSession(user.Name)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	for _, v := range user.Post {
		v.Connected = true
	}

	temp, err := template.ParseFiles("./frontend/Categories/Cat-3/cat3.html")
	if err != nil {
		log.Println("Error parsing template:", err)
		return
	}

	_, err = r.Cookie("session")
	if err != nil {
		user.Connected = false
		for _, v := range user.Post {
			v.Connected = false
		}

		Posts.Connected = false
		userComment.Connected = false
		data.DeleteSession(user.Name)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	data.Categorie3FeedPost(user.Name)
	fmt.Printf("user.Name: %v\n", user.Name)

	if err = temp.ExecuteTemplate(w, "categorie3", map[string]any{
		"user":        user,
		"categories3": data.Categorie3FeedPost(user.Name),
	}); err != nil {
		log.Println("Error executing template:", err)
		return
	}
	fmt.Printf("categorie3Feed: %v\n", categorie3Feed)
}

/********************************************************************************/

/*************************** FUNCTION COMMENT **********************************/

func comment(w http.ResponseWriter, r *http.Request) {
	temp, err := template.ParseFiles("./frontend/Commentaire/comment.html")
	if err != nil {
		log.Println("Error parsing template:", err)
		return
	}

	userProfil := data.GetUserProfil()
	user.Name = userProfil["name"]
	user.Image = userProfil["userImage"]

	if r.Method == "POST" {

		currentTime := time.Now().Format("15:04  2-Jan-2006")
		postID := r.FormValue("Post_values")

		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}

		message := r.FormValue("message")
		fmt.Printf("message: %v\n", message)

		if message != "" && user.Connected {
			dataBase.UserComment(user.Name, message, script.GenerateCommentID(), currentTime, postID)
			data.NotifComment(postID)
			homefeed = data.HomeFeedPost()
		}
		var commentid, pressed string
		buttonLike := r.FormValue("like")
		buttonDislike := r.FormValue("dislike")
		if buttonLike != "" {

			cutLike := strings.Split(buttonLike, "+")
			fmt.Printf("cutLike: %v\n", cutLike)
			commentid = cutLike[0]
			pressed = cutLike[1]
		}

		if buttonDislike != "" {
			cutDisLike := strings.Split(buttonDislike, "+")
			fmt.Printf("cutDisLike: %v\n", cutDisLike)
			commentid = cutDisLike[0]
			pressed = cutDisLike[1]
		}

		fmt.Printf("commentid: %v\n", commentid)
		fmt.Printf("pressed: %v\n", pressed)

		if commentid != "" && postID != "" && pressed == "like" {

			countLike := 0
			row := data.Db.QueryRow("SELECT COUNT (*) FROM likes WHERE username = ? AND comment_id = ?", user.Name, commentid)
			err := row.Scan(&countLike)
			if err != nil {
				panic(err)
			}

			fmt.Printf("countLike: %v\n", countLike)
			if countLike == 0 {
				data.AddingCommentLike(commentid, countLike, user.Name, currentTime)
				data.NotifLikeComment(commentid)

			}

			homefeed = dataBase.HomeFeedPost()
			comments = data.GetComment(postID)

		}

		if commentid != "" && postID != "" && pressed == "dislike" {

			fmt.Println("Enter the Dislike condition")

			countLike := 0
			row := data.Db.QueryRow("SELECT COUNT (*) FROM dislikes WHERE username = ? AND comment_id = ?", user.Name, commentid)
			err := row.Scan(&countLike)
			if err != nil {
				panic(err)
			}

			if countLike == 0 {
				data.AddingCommentDisLike(commentid, countLike, user.Name, currentTime)
				data.NotifDisLikeComment(commentid)

			}

			homefeed = dataBase.HomeFeedPost()
			comments = data.GetComment(postID)

		}

		homefeed = dataBase.HomeFeedPost()
		comments = data.GetComment(postID)
		if err := temp.ExecuteTemplate(w, "comment", map[string]any{
			"PostID":    postID,
			"UserImage": user.Image,
			"UserName":  user.Name,
			"Comments":  comments,
		}); err != nil {
			log.Println("Error executing template:", err)
			return
		}

		if postID != "" && message != "" && (r.FormValue("like") == "" && r.FormValue("dislike") == "") {
			http.Redirect(w, r, "/comment?postid="+postID, http.StatusSeeOther)
			return

		}

	} else if r.Method == "GET" {

		postid := r.URL.Query().Get("postid")

		fmt.Printf("postid: %v\n", postid)
		if user.Connected && postid != "" {
			// currentTime := time.Now().Format("15:04  2-Janv-2006")

			countComment := 0

			row := data.Db.QueryRow("SELECT COUNT (*) FROM comments WHERE post_id = ?", postid)
			err := row.Scan(&countComment)
			if err != nil {
				panic(err)
			}
			if countComment >= 0 {

				dataBase.AddingCountComment(postid, user.Name)
			}
			fmt.Printf("postid: %v\n", postid)
		}

		if r.Method == "POST" && (r.FormValue("like") != "" || r.FormValue("dislike") != "") {
			http.Redirect(w, r, "/comment", http.StatusNotFound)
			return

		}

		homefeed = data.HomeFeedPost()
		if err := temp.ExecuteTemplate(w, "comment", map[string]any{
			"PostID":    postid,
			"UserImage": user.Image,
			"UserName":  user.Name,
			"Comments":  data.GetPostComment(postid),
		}); err != nil {
			log.Println("Error executing template:", err)
			return
		}
	}
}

/*************************** FUNCTION MODERATION **********************************/

func moderateur(w http.ResponseWriter, r *http.Request) {

	if len(uAccount) < data.LenUser() {
		uAccount = data.GetAllUsers()
	}

	fmt.Printf("len(uAccount): %v\n", len(uAccount))

	if len(uAccount) > 0 && user.Connected {
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
	}
	var imageSRC string
	if r.Method == "POST" {

		message := r.FormValue("message")
		Posts.Categories = r.FormValue("categories")
		Posts.Categories2 = r.FormValue("categories2")

		fmt.Printf("Posts.Categories: %v\n", Posts.Categories)
		if message != "" && user.Connected {
			postid := script.GeneratePostID()
			currentTime := time.Now().Format("15:04  2-Janv-2006")

			user.Post = preappendPost(structure.Post{
				PostID:      postid,
				Name:        user.Name,
				Message:     message,
				DateTime:    currentTime,
				UserImage:   user.Image,
				CountCom:    data.LenUserComment(postid),
				Categories:  Posts.Categories,
				Categories2: Posts.Categories2,
				Connected:   true,
			})
			file, header, err := r.FormFile("myFile")

			imageName := ""
			maxImageSize := 20 * 1024 * 1024 // 20Mo en octets
			if err != nil {
				if err != http.ErrMissingFile {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				//Put the message in the dataBase
				dataBase.UserPost(user.Name, message, postid, user.Image, currentTime, imageName, Posts.Count, Posts.CountDis, Posts.CountCom, Posts.Categories, Posts.Categories2)
				homefeed = dataBase.HomeFeedPost()

			} else {
				imageName = header.Filename
				fmt.Printf("Uploaded File: %+v\n", header.Filename)

				// read all of the contents of our uploaded file into a
				// byte array
				fileBytes, err := ioutil.ReadAll(file)

				if err != nil {
					fmt.Println(err)
				}

				imageSRC = "./frontend/upload-image/" + imageName

				err = ioutil.WriteFile(imageSRC, fileBytes, 0o666)
				if err != nil {
					log.Fatal(err)
				}

				fileIMAGE, err := os.Open(imageSRC)
				if err != nil {
					fmt.Println(err)
				}
				fileStat, err := fileIMAGE.Stat()
				if err != nil {
					fmt.Println(err)
				}
				if fileStat.Size() > int64(maxImageSize) {
					os.Remove(imageSRC)
				} else {
					_, user.Post = dataBase.UserPost(user.Name, message, postid, user.Image, currentTime, imageSRC, Posts.Count, Posts.CountDis, Posts.CountCom, Posts.Categories, Posts.Categories2)
					homefeed = dataBase.HomeFeedPost()
					file.Close()

				}

			}
		}
		if r.FormValue("like") != "" && user.Connected {
			currentTime := time.Now().Format("15:04  2-Janv-2006")

			postid := r.FormValue("like")
			countLike := 0
			// row := data.Db.QueryRow("SELECT countLikes FROM posts WHERE name = ? AND postid = ?", user.Name, postid)
			row := data.Db.QueryRow("SELECT COUNT (*) FROM likes WHERE username = ? AND post_id = ?", user.Name, postid)
			err := row.Scan(&countLike)
			if err != nil {
				panic(err)
			}
			if countLike == 0 {
				fmt.Printf("countLike: %v\n", countLike)
				dataBase.AddingCountLike(postid, user.Name, currentTime)
			}

			fmt.Printf("postid: %v\n", postid)
			homefeed = dataBase.HomeFeedPost()
		}

		if r.FormValue("dislike") != "" && user.Connected {
			currentTime := time.Now().Format("15:04  2-Janv-2006")

			postid := r.FormValue("dislike")
			countDislike := 0
			// row := data.Db.QueryRow("SELECT countLikes FROM posts WHERE name = ? AND postid = ?", user.Name, postid)
			row := data.Db.QueryRow("SELECT COUNT (*) FROM dislikes WHERE username = ? AND post_id = ?", user.Name, postid)
			err := row.Scan(&countDislike)
			if err != nil {
				panic(err)
			}
			if countDislike == 0 {
				fmt.Printf("countDislike: %v\n", countDislike)
				data.AddingCountDislike(postid, user.Name, currentTime)
			}

			fmt.Printf("postid: %v\n", postid)
			homefeed = dataBase.HomeFeedPost()
		}
	}
	if r.FormValue("delete") != "" && user.Connected {
		data.DeletePost(r.FormValue("delete"))
	}

	temp, err := template.ParseFiles("./frontend/Moderateur/moderateur.html")
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
			userComment.Connected = false
			data.DeleteSession(user.Name)
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		user.Connected = true
		for _, v := range user.Post {
			v.Connected = true
		}

		homefeed = data.HomeFeedPost()

		err = temp.ExecuteTemplate(w, "moderateur", map[string]any{
			"User":     user,
			"HomeFeed": homefeed,
		})
		if err != nil {
			log.Fatal(err)
		}

	} else {

		homefeed = data.HomeFeedPost()

		err = temp.ExecuteTemplate(w, "moderateur", map[string]any{
			"User":     user,
			"HomeFeed": homefeed,
		})
		if err != nil {
			log.Fatal(err)
		}
	}
}

/*************************** FUNCTION ADMIN **********************************/

func admin(w http.ResponseWriter, r *http.Request) {

	var ModerateurRequestParam map[string]string

	if len(uAccount) < data.LenUser() {
		uAccount = data.GetAllUsers()
	}

	// if len(adminfeed) < len(data.AdminFeedPost()) {
	// 	adminfeed = data.AdminFeedPost()
	// }

	fmt.Printf("len(uAccount): %v\n", len(uAccount))

	if len(uAccount) > 0 && user.Connected {
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
	}
	var imageSRC string
	if r.Method == "POST" {

		message := r.FormValue("message")
		Posts.Categories = r.FormValue("categories")
		Posts.Categories2 = r.FormValue("categories2")

		fmt.Printf("Posts.Categories: %v\n", Posts.Categories)
		if message != "" && user.Connected {
			postid := script.GeneratePostID()
			currentTime := time.Now().Format("15:04  2-Janv-2006")

			// user.Name = profil["name"]
			// user.Image = profil["image"]

			user.Post = preappendPost(structure.Post{
				PostID:      postid,
				Name:        user.Name,
				Message:     message,
				DateTime:    currentTime,
				UserImage:   user.Image,
				CountCom:    data.LenUserComment(postid),
				Categories:  Posts.Categories,
				Categories2: Posts.Categories2,
				Connected:   true,
			})
			file, header, err := r.FormFile("myFile")

			imageName := ""
			maxImageSize := 20 * 1024 * 1024 // 20Mo en octets
			if err != nil {
				if err != http.ErrMissingFile {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				//Put the message in the dataBase
				dataBase.UserPost(user.Name, message, postid, user.Image, currentTime, imageName, Posts.Count, Posts.CountDis, Posts.CountCom, Posts.Categories, Posts.Categories2)
				homefeed = dataBase.HomeFeedPost()

			} else {
				imageName = header.Filename
				fmt.Printf("Uploaded File: %+v\n", header.Filename)

				// read all of the contents of our uploaded file into a
				// byte array
				fileBytes, err := ioutil.ReadAll(file)

				if err != nil {
					fmt.Println(err)
				}

				imageSRC = "./frontend/upload-image/" + imageName

				err = ioutil.WriteFile(imageSRC, fileBytes, 0o666)
				if err != nil {
					log.Fatal(err)
				}

				fileIMAGE, err := os.Open(imageSRC)
				if err != nil {
					fmt.Println(err)
				}
				fileStat, err := fileIMAGE.Stat()
				if err != nil {
					fmt.Println(err)
				}
				if fileStat.Size() > int64(maxImageSize) {
					os.Remove(imageSRC)
				} else {
					_, user.Post = dataBase.UserPost(user.Name, message, postid, user.Image, currentTime, imageSRC, Posts.Count, Posts.CountDis, Posts.CountCom, Posts.Categories, Posts.Categories2)
					homefeed = dataBase.HomeFeedPost()
					file.Close()

				}
			}
		}

		if r.FormValue("accepted") != "" && user.Connected {

			ModerateurRequestParam = data.SelectUserForModo(r.FormValue("accepted"))
			modoName := ModerateurRequestParam["name"]
			data.AddingUser2Modo(modoName)
			data.DeleteAdminRequest(modoName, r.FormValue("accepted"))

		}

		if r.FormValue("refused") != "" && user.Connected {

			modoRequestName := ""
			modoRequestID := ""

			for _, v := range data.AdminFeedPost() {
				modoRequestID = v.NotifID
				modoRequestName = v.Name
			}

			fmt.Printf("modoRequestID: %v\n", modoRequestID)
			fmt.Printf("modoRequestName: %v\n", modoRequestName)

			data.DeleteAdminRequest(modoRequestName, modoRequestID)

		}
		if r.FormValue("delete") != "" && user.Connected {
			data.DeleteModerateur(r.FormValue("delete"))
		}
	}
	if r.FormValue("delete") != "" && user.Connected {
		data.DeletePost(r.FormValue("delete"))
	}

	temp, err := template.ParseFiles("./frontend/Admin/admin.html")
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
			userComment.Connected = false
			data.DeleteSession(user.Name)
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		user.Connected = true
		for _, v := range user.Post {
			v.Connected = true
		}

		homefeed = data.HomeFeedPost()
		adminfeed = data.AdminFeedPost()

		err = temp.ExecuteTemplate(w, "admin", map[string]any{
			"User":        user,
			"HomeFeed":    homefeed,
			"AdminFeed":   adminfeed,
			"Moderateur1": data.GetAllModerateur()["moderateur1"],
			"Moderateur2": data.GetAllModerateur()["moderateur2"],
			"Moderateur3": data.GetAllModerateur()["moderateur3"],
		})
		if err != nil {
			log.Fatal(err)
		}

	} else {

		homefeed = data.HomeFeedPost()

		err = temp.ExecuteTemplate(w, "admin", map[string]any{
			"User":     user,
			"HomeFeed": homefeed,
		})
		if err != nil {
			log.Fatal(err)
		}
	}
}
