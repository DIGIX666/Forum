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
	"os"
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
	uAccount = data.GetAllUsers()
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

			dataBase.AddSession(uName, uuidUser, cookie.Value)
			user.Connected = true
			Posts.Connected = true
			userComment.Connected = true
			data.SetGoogleUserUUID(uEmail)
			uAccount = data.GetAllUsers()
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
			dataBase.AddSession(GitHub_UserName, uuidGithubUser, cookie.Value)
			http.SetCookie(w, &cookie)
			user.Connected = true
			Posts.Connected = true
			userComment.Connected = true
			uAccount = data.GetAllUsers()
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
		// uAccount = data.GetAllUsers()
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
			maxImageSize := 20 * 1024 * 1024 // 20Mo en octets
			if err != nil {
				if err != http.ErrMissingFile {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				//Put the message in the dataBase
				dataBase.UserPost(user.Name, message, postid, user.Image, currentTime, imageName, Posts.Count, Posts.CountDis, Posts.CountCom)
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
				// fileIMAGE, _ := os.Open(imageSRC)
				// fileStat, _ := fileIMAGE.Stat()
				// fileStat.Size()

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
					_, user.Post = dataBase.UserPost(user.Name, message, postid, user.Image, currentTime, imageSRC, Posts.Count, Posts.CountDis, Posts.CountCom)
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
			// for i := range user.Post {
			// 	fmt.Printf("i: %v\n", i)
			// 	fmt.Printf("user.Post[i].PostID: %v\n", user.Post[i].PostID)
			// 	if user.Post[i].PostID == postid && countLike < 1 {
			// 		user.Post[i].Count++
			// 		countLike = user.Post[i].Count
			// 	}
			// 	fmt.Printf("conteur %v\n", user.Post[i].Count)
			// }
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
			// for i := range user.Post {
			// 	fmt.Printf("i: %v\n", i)
			// 	fmt.Printf("user.Post[i].PostID: %v\n", user.Post[i].PostID)
			// 	if user.Post[i].PostID == postid && countLike < 1 {
			// 		user.Post[i].Count++
			// 		countLike = user.Post[i].Count
			// 	}
			// 	fmt.Printf("conteur %v\n", user.Post[i].Count)
			// }
			fmt.Printf("postid: %v\n", postid)
			homefeed = dataBase.HomeFeedPost()
		}

	}

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
			userComment.Connected = false
			data.DeleteSession(user.Name)
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		user.Connected = true
		for _, v := range user.Post {
			v.Connected = true
		}

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

		Posts.Connected = false
		userComment.Connected = false
		data.DeleteSession(user.Name)
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
		if user.Connected && postid != "" {
			currentTime := time.Now().Format("15:04  2-Janv-2006")

			countComment := 0
			// row := data.Db.QueryRow("SELECT countLikes FROM posts WHERE name = ? AND postid = ?", user.Name, postid)
			row := data.Db.QueryRow("SELECT COUNT (*) FROM comments WHERE post_id = ?", postid)
			err := row.Scan(&countComment)
			if err != nil {
				panic(err)
			}
			if countComment >= 0 {
				fmt.Printf("countComment: %v\n", countComment)
				dataBase.AddingCountComment(postid, user.Name, currentTime)
			}
			fmt.Printf("postid: %v\n", postid)
			homefeed = dataBase.HomeFeedPost()
		}
	}

}
