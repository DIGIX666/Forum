package data

import (
	structure "Forum/Struct"
	script "Forum/scripts"
	"database/sql"
	"fmt"
	"log"

	"github.com/gofrs/uuid"
	_ "github.com/mattn/go-sqlite3"
)

var uAccount []structure.UserAccount
var user structure.UserAccount
var posts []structure.Post
var Db *sql.DB

/*************************** CREATE DATA BASE **********************************/
func CreateDataBase() {

	var err error
	Db, err = sql.Open("sqlite3", "./usersForum.db")
	if err != nil {
		fmt.Println("Erreur ouverture de la base de donnée à la creation de la table:")
		log.Fatal(err)
	}

	_, err = Db.Exec(`CREATE TABLE IF NOT EXISTS users
        (id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT,
        image TEXT,
        email TEXT,
        uuid TEXT,
        password TEXT,
        admin BOOLEAN
        )`)
	if err != nil {
		log.Println("erreur creation de table users")
		log.Fatal(err)
	}

	_, err = Db.Exec(`CREATE TABLE IF NOT EXISTS comments (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT,
        commentid TEXT,
        content TEXT,
		date TEXT,
		post_id TEXT
    )`)
	if err != nil {
		log.Println("erreur creation de table comments")
		log.Fatal(err)
	}

	_, err = Db.Exec(`CREATE TABLE IF NOT EXISTS posts (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        postid TEXT,
		image TEXT,
        name TEXT,
        message TEXT,
        datetime TEXT,
		picture TEXT
		count INTEGER DEFAULT 0

    )`)

	if err != nil {
		log.Println("erreur creation de table posts")
		log.Fatal(err)
	}

	_, err = Db.Exec(`CREATE TABLE IF NOT EXISTS session (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		uuid TEXT,
		cookie TEXT
		)`)
	if err != nil {
		fmt.Println("erreur creation de table session")
		log.Fatal(err)
	}

	_, err = Db.Exec(`CREATE TABLE IF NOT EXISTS likes (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT,
        datetime TEXT,
		post_id INTEGER,
		count INTEGER DEFAULT 0,
		FOREIGN KEY (post_id) REFERENCES posts(postid)

    )`)
	if err != nil {
		log.Println("erreur creation de table likes")
		log.Fatal(err)
	}

	_, err = Db.Exec(`CREATE TABLE IF NOT EXISTS dislikes (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT,
        datetime TEXT,
		post_id INTEGER,
		FOREIGN KEY (post_id) REFERENCES posts(postid)
    )`)
	if err != nil {
		log.Println("erreur creation de table dislikes")
		log.Fatal(err)
	}
}

/*************************** ADD SESSION **********************************/
func AddSession(name string, uuid string, cookie string) {

	_, err := Db.Exec(`CREATE TABLE IF NOT EXISTS session (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		uuid TEXT,
		cookie TEXT
		)`)
	if err != nil {
		fmt.Println("erreur creation de table session")
		log.Fatal(err)
	}
	fmt.Printf("uuidUser: %v\n", uuid)
	fmt.Printf("userName: %v\n", name)
	fmt.Printf("cookie.Value: %v\n", cookie)

	if name != "" || uuid != "" || cookie != "" {

		_, err := Db.Exec("INSERT INTO session (name, uuid, cookie) VALUES (?, ?, ?)", name, uuid, cookie)
		if err != nil {
			fmt.Println("Erreur à l'insertion de donnée dans session, func AddSession:")
			log.Fatal(err)
		}
	} else {
		fmt.Println("name uuid cookie vide !")
	}

}

/*************************** DELETE SESSION **********************************/
func DeleteSession(name string) {

	_, err := Db.Exec("DELETE FROM session WHERE name = ?", name)
	if err != nil {
		fmt.Println("Erreur lors de la suppression de la session dans la base de données, func DeleteSession:")
		log.Fatal(err)
	}
}

/*************************** DATA BASE REGISTER **********************************/
func DataBaseRegister(email string, password string) bool {

	uuid := ""
	checkExisting := false

	var count int
	err := Db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", email).Scan(&count)
	if err != nil {
		fmt.Println("error reading database to found email !!")
	}
	if count > 0 {

		fmt.Println("email adress already exist !")
		checkExisting = true

	} else {
		_, err = Db.Exec("INSERT INTO users (name, image, email, uuid, password, admin) VALUES (?, ?, ?,?,?,?)", "none", "../assets/images/beehive-37436.svg", email, uuid, password, false)
		if err != nil {
			log.Fatal(err)
		}
	}

	rows, err := Db.Query("SELECT id, email, password FROM users")
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		var id int
		var email string
		var password string

		err = rows.Scan(&id, &email, &password)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("id: %v\n", id)
		fmt.Printf("name: %v\n", email)
		fmt.Printf("password: %v\n", password)

	}

	if !checkExisting {
		return true
	} else {
		return false
	}

}

/*************************** DATA BASE LOGIN **********************************/
func DataBaseLogin(email string, password string, uuid string) bool {
	var hashpassword string
	err := Db.QueryRow("SELECT password FROM users WHERE email = ?", email).Scan(&hashpassword)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("email: %v\n", email)
	fmt.Printf("hashpassword: %v\n", hashpassword)

	compare := script.ComparePassword(hashpassword, password)

	fmt.Printf("password of user in the dataBase? %v\n", compare)

	fmt.Printf("uuid: %v\n", uuid)

	if compare {

		return true
	} else {
		fmt.Println("pas le bon mot de passe")
		return false
	}

}

/*************************** CHECK GOOGLE USER LOGIN **********************************/
func CheckGoogleUserLogin(email string, email_verified string, uuid string) bool {
	if email_verified == "false" {

		return false

	} else {
		_, err := Db.Exec("UPDATE users SET UUID = ? WHERE email = ?", uuid, email)
		if err != nil {
			fmt.Println(err)
			return false

		} else {
			return true
		}
	}
}

/*************************** CHECK USER LOGIN **********************************/
func CheckUserLogin(email string, password string, uuid string) bool {

	var hashpassword string
	err := Db.QueryRow("SELECT password FROM users WHERE email = ?", email).Scan(&hashpassword)
	if err != nil {
		fmt.Println("Erreur SELECT fonction checkUserLogin: ")
		fmt.Println(err)
		return false
	}

	fmt.Printf("email: %v\n", email)
	fmt.Printf("hashpassword: %v\n", hashpassword)
	compare := script.ComparePassword(hashpassword, password)

	fmt.Printf("password of user in the dataBase? %v\n", compare)

	_, err = Db.Exec("UPDATE users SET UUID = ? WHERE email = ?", uuid, email)
	if err != nil {
		fmt.Println(err)
		return false

	} else {
		return true
	}

}

/*************************** PREAPPEND POST **********************************/
func preappendPost(c structure.Post) []structure.Post {
	posts = append(posts, structure.Post{})
	copy(posts[1:], posts)
	posts[0] = c
	return posts
}

/************************* USER POST **********************************/
func UserPost(userName string, message string, postID string, image string, dateTime string, pictureURL string) (bool, []structure.Post) {

	// _, err := Db.Exec("INSERT INTO posts (name, message, postid,image, datetime,picture) VALUES (?, ?, ?,?,?,?)", userName, message, postID, image, dateTime, pictureURL)
	// if err != nil {
	// 	fmt.Println(" Insert user Post to the dataBase:")
	// 	log.Fatal(err)
	// } else {
	// 	return true
	// }
	// return false
	NumberOfComment := 0

	fmt.Printf("image: %v", pictureURL)
	fmt.Println("")
	if pictureURL != "" {

		_, err := Db.Exec("INSERT INTO posts (name, message, postid,image, datetime,picture) VALUES (?, ?, ?,?,?,?)", userName, message, postID, image, dateTime, pictureURL)
		if err != nil {
			fmt.Println("Error Insert user Post to the dataBase:")
			log.Fatal(err)
		} else {

			user.Post = preappendPost(structure.Post{
				PostID:          postID,
				Name:            userName,
				Message:         message,
				DateTime:        dateTime,
				Picture:         pictureURL,
				NumberOfComment: NumberOfComment,
				Connected:       true,
			})

			return true, user.Post
		}

		return false, user.Post
	} else {
		_, err := Db.Exec("INSERT INTO posts (name, message, postid,image, datetime,picture) VALUES (?, ?, ?,?,?,?)", userName, message, postID, image, dateTime, "")
		if err != nil {
			fmt.Println("Error Insert user Post to the dataBase:")
			log.Fatal(err)
		} else {

			user.Post = preappendPost(structure.Post{
				PostID:          postID,
				Name:            userName,
				Message:         message,
				DateTime:        dateTime,
				Picture:         "",
				NumberOfComment: NumberOfComment,
				Connected:       true,
			})

			return true, user.Post
		}
		return false, user.Post
	}
}

/*************************** USER COMMENT **********************************/
func UserComment(userName string, message string, CommentID string, dateTime string, postID string) bool {
	_, err := Db.Exec("INSERT INTO comments (name, content, commentid, date,post_id) VALUES (?, ?, ?, ?,?)", userName, message, CommentID, dateTime, postID)
	if err != nil {
		fmt.Println("Error Insert user Comment to the dataBase:")
		log.Fatal(err)
	} else {
		return true
	}
	return false
}

/*************************** SET GOOGLE USER UUID **********************************/
func SetGoogleUserUUID(userEmail string) string {

	uuidGenerated, _ := uuid.NewV4()
	uuid := uuidGenerated.String()

	_, err := Db.Exec("UPDATE users SET UUID = ? WHERE email = ?", uuid, userEmail)
	if err != nil {
		fmt.Println(err)
	}
	return uuid
}

/*************************** SET GIT HUB UUID **********************************/
func SetGitHubUUID(userName string) string {

	uuidGenerated, _ := uuid.NewV4()
	uuid := uuidGenerated.String()

	_, err := Db.Exec("UPDATE users SET UUID = ? WHERE name = ?", uuid, userName)
	if err != nil {
		fmt.Println("Error function SetGitUUID dataBase:")
		fmt.Println(err)
	}
	return uuid

}

/*************************** ADD LIKES **********************************/
// func AddLikes(userName string, postID string, dateTime string) {

// 	_, err := Db.Exec("INSERT INTO likes (username,numberlikes,postid,datetime) VALUES (?,?,?,?)", userName, postID, dateTime)
// 	if err != nil {
// 		fmt.Println("Error function AddLikes dataBase:")
// 		log.Fatal(err)
// 	}
// }

/*************************** NUMBER OF LIKES POST **********************************/
// cette fonction permet d'obtenir des likes total d'un post.

// func NumberOFLikesPost(postID string) int {

// 	var numberLikes int
// 	err := Db.QueryRow("COUNT (*) FROM likes").Scan(&numberLikes)
// 	if err != nil {
// 		fmt.Println("Error SELECT From NumberODLikes dataBase:")
// 		log.Fatal(err)
// 	}

// 	return numberLikes
// }

/*************************** ADD DISLIKES **********************************/
// func AddDisLikes() {

// }

/*************************** GET USER PROFIL **********************************/
func GetUserProfil() map[string]string {

	ans := make(map[string]string, 5)
	var id int
	var name, uuid, cookie string

	err := Db.QueryRow("SELECT * FROM session ORDER BY id DESC LIMIT 1").Scan(&id, &name, &uuid, &cookie)
	if err != nil {
		fmt.Println("Erreur SELECT fonction GetUserProfil dataBase:")
		log.Fatal(err)
	}

	var userImage, userEmail, admin string

	err = Db.QueryRow("SELECT image,email,admin FROM users WHERE name = ?", name).Scan(&userImage, &userEmail, &admin)
	if err != nil {
		fmt.Println("Erreur SELECT #2 fonction GetUserProfil dataBase:")
		log.Fatal(err)
	}

	fmt.Printf("User IMAGE: %v\n", userImage)

	ans["name"] = name
	ans["email"] = userEmail
	ans["userImage"] = userImage
	ans["uuid"] = uuid
	ans["admin"] = admin

	fmt.Printf("ans: %v\n", ans)
	return ans
}

/*************************** GET POST **********************************/
// func GetPost() []structure.Post {
// 	ans := make(map[string]string)
// 	var id int
// 	var postID, message, dataTime, name, pictureURL, image string

// 	//var Post []structure.Post

// 	err := Db.QueryRow("SELECT * FROM posts ORDER BY id DESC LIMIT 1").Scan(&id, &postID, &name, &message, &image, &dataTime, &pictureURL)
// 	if err != nil {
// 		fmt.Println("Erreur SELECT fonction GetLastPost dataBase:")
// 		log.Fatal(err)
// 	}
// 	//ans["id"] = id
// 	ans["postID"] = postID
// 	ans["image"] = image
// 	ans["userName"] = name
// 	ans["message"] = message
// 	ans["dataTime"] = dataTime
// 	ans["pictureURL"] = pictureURL

// 	/*Post = preappendPost(structure.Post{
// 		PostID:    postID,
// 		Name:      name,
// 		UserImage: image,
// 		Message:   message,
// 		DateTime:  dataTime,
// 		Picture:   pictureURL,
// 	})*/

// 	return ans
// }

/*************************** HOME FEED **********************************/
func HomeFeed() []structure.Post {

	rows, err := Db.Query("SELECT * FROM posts ORDER BY id")
	if err != nil {
		fmt.Println("Error in HomeFeed Function dataBase:")
		log.Fatal(err)
	}
	var Posts []structure.Post
	var id int
	var postID, userName, message, image, dateTime, picture string

	for rows.Next() {

		err := rows.Scan(&id, &postID, &image, &userName, &message, &dateTime, &picture)
		if err != nil {
			fmt.Println("Error HomeFeed Function in rows.Scan:")
			log.Fatal(err)
		}
		Posts = preappendPost(structure.Post{
			PostID:    postID,
			Name:      userName,
			UserImage: image,
			Message:   message,
			DateTime:  dateTime,
			Picture:   picture,
		})

	}
	return Posts

}

/*************************** PREPEND COMMENT **********************************/
func prependComment(x []structure.Comment, y structure.Comment) []structure.Comment {
	x = append(x, structure.Comment{})
	copy(x[1:], x)
	x[0] = y
	return x
}

/*************************** GET POST COMMENT **********************************/
func GetPostComment(postID string) []structure.Comment {

	rows, err := Db.Query("SELECT * FROM comments WHERE post_id = ?", postID)
	if err != nil {
		fmt.Println("Error in GetPostCommnet Query didn't work:")
		log.Fatal(err)
	}

	var ans []structure.Comment

	var _id int

	var name, commentid, content, date, post_id string

	for rows.Next() {

		err := rows.Scan(&_id, &name, &commentid, &content, &date, &post_id)
		if err != nil {
			fmt.Println("Error GetPostComment Function in rows.Scan:")
			log.Fatal(err)
		}

		ans = prependComment(ans, structure.Comment{
			Message:   content,
			Name:      name,
			DateTime:  date,
			CommentID: commentid,
			PostID:    post_id,
			Connected: false,
		})
	}

	return ans

}

/*************************** PROFIL FEED **********************************/
func ProfilFeed(userName string) []structure.Post {

	rows, err := Db.Query("SELECT id,postid,image,message,datetime,picture FROM posts WHERE name = ?", userName)
	if err != nil {
		fmt.Println("Error in ProfilFeed Function Query didn't work in dataBase:")
		log.Fatal(err)
	}
	var Posts []structure.Post

	for rows.Next() {

		var id int

		var postID, message, dateTime, image, picture string

		err := rows.Scan(&id, &postID, &image, &message, &dateTime, &picture)
		if err != nil {
			fmt.Println("Error ProfilFeed Function in rows.Scan:")
			log.Fatal(err)
		}

		Posts = preappendPost(structure.Post{
			PostID:    postID,
			Name:      userName,
			UserImage: image,
			Message:   message,
			DateTime:  dateTime,
			Picture:   picture,
		})

	}
	return Posts
}

/*************************** PROFIL FEED DELETE **********************************/
func ProfilFeedDelete(userName string) {

	_, err := Db.Query("DELETE postid,message,datetime,picture FROM posts WHERE name != ?", userName)
	if err != nil {
		fmt.Println("Error in ProfilFeedDelete Function Query didn't work in dataBase:")
		log.Fatal(err)
	}
}
