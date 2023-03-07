package main

import (
	structure "Forum/Struct"
	"Forum/data"
	dataBase "Forum/data"
	function "Forum/functions"
	script "Forum/scripts"
	"crypto/tls"
	"fmt"
	"io/ioutil"
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
var userComment structure.Comment
var Posts structure.Post
var uAccount []structure.UserAccount
var homefeed []structure.HomeFeedPost
var countEnter int

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
	homefeed = data.HomeFeedPost()
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

		checkGitHubUserLoogged, GitHub_UserName, _, _ := function.GitHubLog(code)

		if checkGoogleUserLogged {
			uuidGenerated, _ := uuid.NewV4()
			uuidUser := uuidGenerated.String()
			cookie := http.Cookie{

				Value:  uuidUser,
				Name:   "session",
				MaxAge: 7200,
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
				MaxAge: 7200,
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
			http.Redirect(w, r, "/", http.StatusFound)
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

	if user.Id != 0 {
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

		if message != "" && user.Connected {
			postid := script.GeneratePostID()
			currentTime := time.Now().Format("15:04  2-Janv-2006")

			// user.Name = profil["name"]
			// user.Image = profil["image"]

			user.Post = preappendPost(structure.Post{
				PostID:    postid,
				Name:      user.Name,
				Message:   message,
				DateTime:  currentTime,
				UserImage: user.Image,
				Connected: true,
			})

			file, header, err := r.FormFile("myFile")
			imageName := ""
			if err != nil {
				if err != http.ErrMissingFile {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				//Put the message in the dataBase
				dataBase.UserPost(user.Name, message, postid, user.Image, currentTime, imageName, Posts.Count)
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

				imageSRC = "./assets/upload-image/" + imageName

				err = ioutil.WriteFile(imageSRC, fileBytes, 0o666)
				if err != nil {
					log.Fatal(err)
				}

				_, user.Post = dataBase.UserPost(user.Name, message, postid, user.Image, currentTime, imageSRC, Posts.Count)
				homefeed = dataBase.HomeFeedPost()
				file.Close()

			}
		}
		var uid, pid string
		if r.FormValue("like") != "" && user.Connected {
			postid := r.FormValue("like")
			countLike := 0
			row := data.Db.QueryRow("SELECT COUNT(*) FROM likes WHERE username = ? AND post_id = ?", uid, pid)
			err := row.Scan(&countLike)
			if err != nil {
				log.Fatal(err)
			}
			for i := range user.Post {
				fmt.Printf("i: %v\n", i)
				fmt.Printf("user.Post[i].PostID: %v\n", user.Post[i].PostID)
				if user.Post[i].PostID == postid {
					user.Post[i].Count++
					countLike = user.Post[i].Count
				}
				fmt.Printf("conteur %v\n", user.Post[i].Count)
			}
			fmt.Printf("postid: %v\n", postid)
			dataBase.AddingCountLike(countLike, postid)
			homefeed = dataBase.HomeFeedPost()
		}
	}

	// func LikeExists(uid int, pid int) bool {
	// 	count := 0
	// 	row := data.Db.QueryRow("SELECT COUNT(*) FROM likes WHERE user_id = ? AND posts_id = ?", uid, pid)
	// 	err := row.Scan(&count)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	return count != 0
	// }

	//var count int

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

		user.Connected = true
		for _, v := range user.Post {
			v.Connected = true
		}

		like := r.FormValue("like")

		if like != "" {
			fmt.Printf("click Like ! : %v\n", like)
		}

		// user.Name = profil["name"]
		// user.UUID = profil["uuid"]
		// user.Email = profil["email"]
		// user.Image = profil["userImage"]

		err = temp.ExecuteTemplate(w, "home", map[string]any{
			"User":     user,
			"HomeFeed": homefeed,
		})
		if err != nil {
			log.Fatal(err)
		}

	} else {

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

	var profil map[string]string
	profil = data.GetUserProfil()
	user.Name = profil["name"]
	user.Email = profil["email"]
	user.Image = profil["userImage"]
	user.UUID = profil["uuid"]
	if profil["admin"] == "true" {
		user.Admin = true
	} else {
		user.Admin = false
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

	fmt.Printf("len(user.Post): %v\n", len(user.Post))
	fmt.Printf("data.LenUserPost(user.Name): %v\n", data.LenUserPost(user.Name))
	if len(userHomeFeed) < data.LenUserPost(user.Name) {
		userHomeFeed = data.ProfilFeed(user.Name)
	}

	if err = temp.ExecuteTemplate(w, "profil", map[string]any{
		"user":     user,
		"UserPost": userHomeFeed,
	}); err != nil {
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
