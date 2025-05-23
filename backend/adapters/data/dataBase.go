package data

import (
	structure "Forum/backend/infrastructures/Struct"
	script "Forum/backend/infrastructures/cryptage"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gofrs/uuid"
	_ "github.com/mattn/go-sqlite3"
)

var uAccount []structure.UserAccount
var user structure.UserAccount
var posts []structure.Post
var users []structure.UserAccount

// var Notif structure.Notification
//var Notifs []structure.Notification

func preappendPost(c structure.Post) []structure.Post {
	posts = append(posts, structure.Post{})
	copy(posts[1:], posts)
	posts[0] = c
	return posts
}

func prependNotif(x []structure.Notification, y structure.Notification) []structure.Notification {
	x = append(x, structure.Notification{})
	copy(x[1:], x)
	x[0] = y
	return x
}

func preappendUser(c structure.UserAccount) []structure.UserAccount {
	users = append(users, structure.UserAccount{})
	copy(users[1:], users)
	users[0] = c
	return users
}

// func preappendNotif(c structure.Notification) []structure.Notification {
// 	Notifs = append(Notifs, structure.Notification{})
// 	copy(Notifs[1:], Notifs)
// 	Notifs[0] = c
// 	return Notifs
// }

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
        name NOT NULL,
        image NOT NULL,
        email NOT NULL,
        uuid NOT NULL,
        password NOT NULL,
		moderateur BOOLEAN DEFAULT FALSE,
        admin BOOLEAN DEFAULT FALSE
        )`)
	if err != nil {
		log.Println("erreur creation de table users")
		log.Fatal(err)
	}

	_, err = Db.Exec(`CREATE TABLE IF NOT EXISTS notif_admin
        (id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT DEFAULT '',
        avatar TEXT DEFAULT '',
        notifid TEXT DEFAULT '',
        date TEXT DEFAULT ''
        )`)
	if err != nil {
		log.Println("erreur creation de table notif_admin")
		log.Fatal(err)
	}

	_, err = Db.Exec(`CREATE TABLE IF NOT EXISTS comments (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT DEFAULT '',
        commentid TEXT DEFAULT '',
        content TEXT DEFAULT '',
        date TEXT DEFAULT '',
        post_id TEXT DEFAULT '',
        commentLike INTEGER DEFAULT 0,
        commentDislike INTEGER DEFAULT 0
    )`)
	if err != nil {
		log.Println("erreur creation de table comments")
		log.Fatal(err)
	}

	_, err = Db.Exec(`CREATE TABLE IF NOT EXISTS posts (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        postid NOT NULL,
		image NOT NULL,
        name NOT NULL,
        message NOT NULL,
        datetime NOT NULL,
		picture NOT NULL,
		countComment INTEGER DEFAULT 0,
		countLikes INTEGER DEFAULT 0,
		countDislikes INTEGER DEFAULT 0,
		categories TEXT NOT NULL,
		categories2 TEXT NOT NULL
    )`)

	if err != nil {
		log.Println("erreur creation de table posts")
		log.Fatal(err)
	}

	_, err = Db.Exec(`CREATE TABLE IF NOT EXISTS session (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name NOT NULL,
		uuid NOT NULL,
		cookie NOT NULL
		)`)
	if err != nil {
		fmt.Println("erreur creation de table session")
		log.Fatal(err)
	}

	_, err = Db.Exec(`CREATE TABLE IF NOT EXISTS likes (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL,
        datetime TEXT NOT NULL,
		post_id INTEGER,
		comment_id INTEGER,
		FOREIGN KEY (post_id) REFERENCES posts(postid),
		FOREIGN KEY (comment_id) REFERENCES comments(commentid)

    )`)
	if err != nil {
		log.Println("erreur creation de table likes")
		log.Fatal(err)
	}

	_, err = Db.Exec(`CREATE TABLE IF NOT EXISTS dislikes (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
		username NOT NULL,
        datetime NOT NULL,
		post_id INTEGER,
		comment_id INTEGER,
		FOREIGN KEY (post_id) REFERENCES posts(postid),
		FOREIGN KEY (comment_id) REFERENCES comments(commentid)

    )`)
	if err != nil {
		log.Println("erreur creation de table dislikes")
		log.Fatal(err)
	}
	//table notification
	_, err = Db.Exec(`CREATE TABLE IF NOT EXISTS notification (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		notifid TEXT DEFAULT '',
		username TEXT DEFAULT '',	
		avatar TEXT DEFAULT '',
		datetime TEXT DEFAULT '',
		post_id TEXT DEFAULT '',
		comment_id TEXT DEFAULT '',
		like_post BOOLEAN DEFAULT FALSE,
		dislike_post BOOLEAN DEFAULT FALSE,
		like_comment BOOLEAN DEFAULT FALSE,
		dislike_comment BOOLEAN DEFAULT FALSE,
		added_comment BOOLEAN DEFAULT FALSE,
		action TEXT DEFAULT ''
	)`)
	if err != nil {
		log.Fatal(err)
	}
}

/*************************** ADD SESSION **********************************/
func AddSession(name string, uuid string, cookie string) {

	_, err := Db.Exec(`CREATE TABLE IF NOT EXISTS session (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name NOT NULL,
		uuid NOT NULL,
		cookie NOT NULL
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
		_, err = Db.Exec("INSERT INTO users (name, image, email, uuid, password, admin) VALUES (?, ?, ?,?,?,?)", email, "../frontend/images/beehive-37436.svg", email, uuid, password, false)
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

func AddingAdminUser() {
	count := 0
	err := Db.QueryRow("SELECT COUNT(*) FROM users WHERE admin = ?", 1).Scan(&count)
	if err != nil {
		fmt.Println("error reading users database function AddingAdminUser")
	}

	if count == 0 {
		_, err = Db.Exec("INSERT INTO users (name, image, email, uuid, password, admin) VALUES (?, ?, ?,?,?,?)", "admin", "../frontend/images/beehive-37436.svg", "admin", "admin", script.GenerateHash("adminadmin"), true)
		if err != nil {
			log.Fatal(err)
		}
	}

}

func GetAllUsers() []structure.UserAccount {
	rows, err := Db.Query("SELECT id,name,image,email,uuid,password,admin FROM users ORDER BY id")
	if err != nil {
		fmt.Println("Erreur GetAllUsers rows in dataBase:")
		log.Fatal(err)
	}

	userName := ""
	userImage := ""
	email := ""
	uuid := ""
	password := ""
	id := 0
	admin := false

	var allUsers []structure.UserAccount

	for rows.Next() {

		err = rows.Scan(&id, &userName, &userImage, &email, &uuid, &password, &admin)
		if err != nil {
			fmt.Println("Erreur")
			fmt.Printf("err: %v\n", err)
		}

		allUsers = preappendUser(structure.UserAccount{
			Id:       id,
			Name:     userName,
			Image:    userImage,
			Email:    email,
			UUID:     uuid,
			Password: password,
			Admin:    admin,
		})
	}

	return allUsers

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

/************************* USER POST **********************************/
func UserPost(userName string, message string, postID string, image string, dateTime string, pictureURL string, countComment int, countLikes int, countDislikes int, categories string, categories2 string) (bool, []structure.Post) {

	// NumberOfComment := 0

	fmt.Printf("image: %v", pictureURL)
	fmt.Println("")

	_, err := Db.Exec("INSERT INTO posts (name, message, postid,image, datetime,picture,countComment, countLikes, countDislikes, categories, categories2) VALUES (?, ?, ?,?,?,?,?,?,?,?,?)", userName, message, postID, image, dateTime, pictureURL, countComment, countLikes, countDislikes, categories, categories2)
	if err != nil {
		fmt.Println("Error Insert user Post to the dataBase:")
		log.Fatal(err)
		return false, user.Post
	} else {

		user.Post = preappendPost(structure.Post{
			PostID:      postID,
			Name:        userName,
			Message:     message,
			DateTime:    dateTime,
			Picture:     pictureURL,
			Connected:   true,
			CountCom:    countComment,
			Count:       countLikes,
			CountDis:    countDislikes,
			Categories:  categories,
			Categories2: categories2,
			// Admin:       admin,
		})

		return true, user.Post
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

/*************************** GET USER PROFIL **********************************/
func GetUserProfil() map[string]string {

	ans := make(map[string]string, 5)

	var id int
	var name, uuid, cookie string

	err := Db.QueryRow("SELECT * FROM session ORDER BY id DESC LIMIT 1").Scan(&id, &name, &uuid, &cookie)
	if err != nil {
		fmt.Println("Erreur SELECT fonction GetUserProfil dataBase:")
		fmt.Println(err)
	}

	var userImage, userEmail string

	var admin, moderateur bool

	fmt.Printf("NAME USER: %v\n", name)

	err = Db.QueryRow("SELECT image,email,moderateur,admin  FROM users WHERE name = ?", name).Scan(&userImage, &userEmail, &moderateur, &admin)
	if err != nil {
		fmt.Println("Erreur SELECT #2 fonction GetUserProfil dataBase:")
		fmt.Println(err)
	}

	fmt.Printf("MODERATEUR: %v\n", moderateur)
	fmt.Printf("ADMIN: %v\n", admin)

	ans["name"] = name
	ans["email"] = userEmail
	ans["userImage"] = userImage
	ans["uuid"] = uuid

	// boolAdmin, _ := strconv.Atoi(admin)
	// boolModo, _ := strconv.Atoi(moderateur)

	if admin {
		ans["admin"] = "true"
	} else {
		ans["admin"] = "false"
	}
	if moderateur {
		ans["moderateur"] = "true"
	} else {
		ans["moderateur"] = "false"
	}

	return ans
}

/*************************** GET USER POST **********************************/

func GetUserPosts(name string) []structure.Post {
	rows, err := Db.Query("SELECT  postid, name, message, image, datetime,picture, countComment, countLikes, countDislikes, categories, categories2 FROM posts WHERE name = ?", name)
	if err != nil {
		fmt.Println("Erreur SELECT #3 fonction GetUserProfil dataBase:")
		log.Fatal(err)
	}

	postID := ""
	userName := ""
	message := ""
	userImage := ""
	dateTime := ""
	pictureURL := ""
	categories := ""
	categories2 := ""
	// NumberOfComment := 0

	var userPosts []structure.Post

	for rows.Next() {

		err = rows.Scan(&postID, &userName, &message, &userImage, &dateTime, &pictureURL, &categories, &categories2)
		if err != nil {
			log.Fatal(err)
		}

		userPosts = preappendPost(structure.Post{
			PostID:      postID,
			Name:        userName,
			Message:     message,
			UserImage:   userImage,
			DateTime:    dateTime,
			Picture:     pictureURL,
			Categories:  categories,
			Categories2: categories2,
			// NumberOfComment: NumberOfComment,
			Connected: true,
		})
	}

	return userPosts

}

/*************************** PREPEND HOME FEED POST **********************************/
func prependHomeFeedPost(x []structure.HomeFeedPost, y structure.HomeFeedPost) []structure.HomeFeedPost {
	x = append(x, structure.HomeFeedPost{})
	copy(x[1:], x)
	x[0] = y
	return x
}

/*************************** HOME FEED POST **********************************/
func HomeFeedPost() []structure.HomeFeedPost {

	rows, err := Db.Query("SELECT * FROM posts ORDER BY id")
	if err != nil {
		fmt.Println("Error in HomeFeed Function dataBase:")
		log.Fatal(err)
	}
	var Posts []structure.HomeFeedPost
	var id, NumberOfComment, NumberOfLikes, NumberOfDislikes int
	var postID, userName, message, image, dateTime, picture, categories, categories2 string

	for rows.Next() {

		err := rows.Scan(&id, &postID, &image, &userName, &message, &dateTime, &picture, &NumberOfComment, &NumberOfLikes, &NumberOfDislikes, &categories, &categories2)
		if err != nil {
			fmt.Println("Error HomeFeedPost Function in rows.Scan:")
			log.Fatal(err)
		}

		Posts = prependHomeFeedPost(Posts, structure.HomeFeedPost{
			PostID:           postID,
			Name:             userName,
			UserImage:        image,
			Message:          message,
			DateTime:         dateTime,
			Picture:          picture,
			NumberOfComment:  LenUserComment(postID),
			NumberOfLikes:    NumberOfLikes,
			NumberOfDislikes: NumberOfDislikes,
			Categories:       categories,
			Categories2:      categories2,
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
	var commentLike, commentDislike int

	for rows.Next() {

		err := rows.Scan(&_id, &name, &commentid, &content, &date, &post_id, &commentLike, &commentDislike)
		if err != nil {
			fmt.Println("Error GetPostComment Function in rows.Scan:")
			log.Fatal(err)
		}

		ans = prependComment(ans, structure.Comment{
			Message:        content,
			Name:           name,
			DateTime:       date,
			CommentID:      commentid,
			PostID:         post_id,
			CommentLike:    commentLike,
			CommentDislike: commentDislike,
			Connected:      true,
		})
	}

	return ans

}

func LenUserComment(postID string) int {

	var NumberComment int
	err := Db.QueryRow("SELECT COUNT (*) FROM comments WHERE post_id = ?", postID).Scan(&NumberComment)
	if err != nil {
		fmt.Println("Error SELECT From LenUserPost dataBase:")
		log.Fatal(err)
	}

	return NumberComment
}

func GetComment(postID string) []structure.Comment {

	rows, err := Db.Query("SELECT * FROM comments WHERE post_id = ?", postID)
	if err != nil {
		fmt.Println("Error in GetPostCommnet Query didn't work:")
		log.Fatal(err)
	}

	var ans []structure.Comment

	var _id int

	var name, commentid, content, date, post_id string
	var commentLike, commentDislike int

	for rows.Next() {

		err := rows.Scan(&_id, &name, &commentid, &content, &date, &post_id, &commentLike, &commentDislike)
		if err != nil {
			fmt.Println("Error GetPostComment Function in rows.Scan:")
			log.Fatal(err)
		}

		ans = prependComment(ans, structure.Comment{
			Message:        content,
			Name:           name,
			DateTime:       date,
			CommentID:      commentid,
			PostID:         post_id,
			CommentLike:    commentLike,
			CommentDislike: commentDislike,
			Connected:      true,
		})
	}

	return ans

}

/******************************************************************************************************/

func preappendUserFeed(x []structure.UserFeedPost, y structure.UserFeedPost) []structure.UserFeedPost {
	x = append(x, structure.UserFeedPost{})
	copy(x[1:], x)
	x[0] = y
	return x
}

func preappendCategorie1FeedPost(x []structure.Categorie1FeedPost, y structure.Categorie1FeedPost) []structure.Categorie1FeedPost {
	x = append(x, structure.Categorie1FeedPost{})
	copy(x[1:], x)
	x[0] = y
	return x
}
func preappendCategorie2FeedPost(x []structure.Categorie2FeedPost, y structure.Categorie2FeedPost) []structure.Categorie2FeedPost {
	x = append(x, structure.Categorie2FeedPost{})
	copy(x[1:], x)
	x[0] = y
	return x
}
func preappendCategorie3FeedPost(x []structure.Categorie3FeedPost, y structure.Categorie3FeedPost) []structure.Categorie3FeedPost {
	x = append(x, structure.Categorie3FeedPost{})
	copy(x[1:], x)
	x[0] = y
	return x
}

/*************************** PROFIL FEED **********************************/
func ProfilFeed(userName string) []structure.UserFeedPost {

	rows, err := Db.Query("SELECT * FROM posts WHERE name = ?", userName)
	if err != nil {
		fmt.Println("Error in ProfilFeed Function Query didn't work in dataBase:")
		log.Fatal(err)
	}
	var Posts []structure.UserFeedPost

	for rows.Next() {

		var id int
		var postID, name, message, dateTime, image, picture, categories, categories2 string
		var NumberOfComment, NumberOfLikes, NumberOfDislikes int

		err := rows.Scan(&id, &postID, &image, &name, &message, &dateTime, &picture, &NumberOfComment, &NumberOfLikes, &NumberOfDislikes, &categories, &categories2)
		if err != nil {
			fmt.Println("Error ProfilFeed Function in rows.Scan:")
			log.Fatal(err)
		}

		fmt.Printf("NumberOfComment: %v\n", NumberOfComment)

		Posts = preappendUserFeed(Posts, structure.UserFeedPost{
			PostID:           postID,
			UserImage:        image,
			Name:             userName,
			Message:          message,
			DateTime:         dateTime,
			Picture:          picture,
			NumberOfComment:  LenUserComment(postID),
			NumberOfLikes:    NumberOfLikes,
			NumberOfDislikes: NumberOfDislikes,
			Categories:       categories,
			Categories2:      categories2,
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

func LenUserPost(nameUser string) int {

	var NumberPost int
	err := Db.QueryRow("SELECT COUNT (*) FROM posts WHERE name = ?", nameUser).Scan(&NumberPost)
	if err != nil {
		fmt.Println("Error SELECT From LenUserPost dataBase:")
		log.Fatal(err)
	}

	return NumberPost
}

func LenLikeUserPost(nameUser string) int {

	var NumberPost int
	err := Db.QueryRow("SELECT COUNT (*) FROM likes WHERE username = ?", nameUser).Scan(&NumberPost)
	if err != nil {
		fmt.Println("Error SELECT From LenUserPost dataBase:")
		log.Fatal(err)
	}
	return NumberPost
}

func LenCategories1UserPost() int {

	var NumberPost int
	err := Db.QueryRow("SELECT COUNT (*) FROM posts WHERE categories = ? OR categories2 = ?", "cat1", "cat1").Scan(&NumberPost)
	if err != nil {
		fmt.Println("Error SELECT From LenUserPost dataBase:")
		log.Fatal(err)
	}

	return NumberPost
}

func LenUser() int {

	var NumberUser int
	err := Db.QueryRow("SELECT COUNT (*) FROM users").Scan(&NumberUser)
	if err != nil {
		fmt.Println("Error SELECT From LenUserPost dataBase:")
		log.Fatal(err)
	}

	return NumberUser
}

/*************************** PROFIL LIKE FEED **********************************/
func ProfilLikeFeed(userName string) []structure.UserFeedPost {
	var tabpostid []string
	var likeFeed []structure.UserFeedPost
	var postid string
	rows, err := Db.Query("SELECT post_id FROM likes WHERE username = ?", userName)
	if err != nil {
		fmt.Println("Error in ProfilLikeFeed Function Query didn't work in dataBase:")
		log.Fatal(err)
	}
	for rows.Next() {
		rows.Scan(&postid)
		tabpostid = append(tabpostid, postid)
	}

	for _, ps := range tabpostid {
		rows2, err := Db.Query("SELECT * FROM posts WHERE postid = ?", ps)
		if err != nil {
			fmt.Println("Error in ProfilLikeFeed Function Query didn't work in dataBase:")
			log.Fatal(err)
		}
		for rows2.Next() {
			var id int
			var postID, name, message, image, picture, dateTime, categories, categories2 string
			var NumberOfComment, NumberOfLikes, NumberOfDislikes int

			err := rows2.Scan(&id, &postID, &image, &name, &message, &dateTime, &picture, &NumberOfComment, &NumberOfLikes, &NumberOfDislikes, &categories, &categories2)
			if err != nil {
				fmt.Println("Error ProfilLikeFeed Function in rows2.Scan:")
				log.Fatal(err)
			}

			likeFeed = preappendUserFeed(likeFeed, structure.UserFeedPost{
				PostID:           postID,
				Name:             name,
				UserImage:        image,
				Message:          message,
				DateTime:         dateTime,
				Picture:          picture,
				NumberOfComment:  LenUserComment(postID),
				NumberOfLikes:    NumberOfLikes,
				NumberOfDislikes: NumberOfDislikes,
				Categories:       categories,
				Categories2:      categories2,
			})

		}
	}

	return likeFeed
}

/*************************** ADDING COUNT POST **********************************/

func AddingCountLike(postID, username, currentTime string) {
	var count int
	fmt.Printf("username: %v\n", username)
	_, err := Db.Exec("INSERT INTO likes (username, datetime, post_id) VALUES (?,?,?)", username, currentTime, postID)
	if err != nil {
		fmt.Println("Error function AddingCount Insert countLikes Posts to the dataBase:")
		fmt.Printf("err: %v\n", err)
		panic(err)
	}

	row := Db.QueryRow("SELECT COUNT (*) FROM likes WHERE post_id = ?", postID)
	err = row.Scan(&count)
	if err != nil {
		panic(err)
	}
	_, err = Db.Exec("UPDATE posts SET countLikes = ? WHERE postid=?", count, postID)
	if err != nil {
		fmt.Println("Error function AddingCountLike Insert countLikes Posts to the dataBase:")
		fmt.Printf("err: %v\n", err)
		panic(err)
	}
}

func AddingCountDislike(postID, username, currentTime string) {
	var CountDis int
	fmt.Printf("username: %v\n", username)
	_, err := Db.Exec("INSERT INTO dislikes (username, datetime, post_id) VALUES (?,?,?)", username, currentTime, postID)
	if err != nil {
		fmt.Println("Error function AddingCountDislike Insert countDislikes Posts to the dataBase:")
		fmt.Printf("err: %v\n", err)
		panic(err)
	}

	row := Db.QueryRow("SELECT COUNT (*) FROM dislikes WHERE post_id = ?", postID)
	err = row.Scan(&CountDis)
	if err != nil {
		panic(err)
	}
	_, err = Db.Exec("UPDATE posts SET countDislikes = ? WHERE postid=?", CountDis, postID)
	if err != nil {
		fmt.Println("Error function AddingCountDislikes Insert countDislikes Posts to the dataBase:")
		fmt.Printf("err: %v\n", err)
		panic(err)
	}
}

func AddingCountComment(postID, username string) {
	var CountComment int
	fmt.Printf("username: %v\n", username)

	row := Db.QueryRow("SELECT COUNT (*) FROM comments WHERE post_id = ?", postID)
	err := row.Scan(&CountComment)
	if err != nil {
		panic(err)
	}
	_, err = Db.Exec("UPDATE posts SET countComment = ? WHERE postid=?", CountComment, postID)
	if err != nil {
		fmt.Println("Error function AddingCountComment Insert countComment Posts to the dataBase:")
		fmt.Printf("err: %v\n", err)
		panic(err)
	}
}

/*************************** ADDING COUNT COMMENT **********************************/
func AddingCommentLike(commentid string, countLike int, userName string, currentTime string) {

	_, err := Db.Exec("INSERT INTO likes (username,datetime,comment_id) VALUES (?,?,?)", userName, currentTime, commentid)
	if err != nil {
		fmt.Println("Error function AddingCommentLike Insert commentLike,date Comments to the dataBase:")
		fmt.Printf("err: %v\n", err)
	}

	count := 0
	row := Db.QueryRow("SELECT COUNT (*) FROM likes WHERE comment_id = ?", commentid)
	err = row.Scan(&count)
	if err != nil {
		fmt.Println("Error Select Count in AddingCommentLike in Database:")
		fmt.Println(err)

	}
	fmt.Printf("count in database: %v\n", count)

	_, err = Db.Exec("UPDATE comments SET commentLike = ? WHERE commentid = ?", count, commentid)
	if err != nil {
		fmt.Println("Error function AddingCountLike Insert countLikes Posts to the dataBase:")
		fmt.Printf("err: %v\n", err)

	}

}

func AddingCommentDisLike(commentid string, countLike int, userName string, currentTime string) {

	_, err := Db.Exec("INSERT INTO dislikes (username,datetime,comment_id) VALUES (?,?,?)", userName, currentTime, commentid)
	if err != nil {
		fmt.Println("Error function AddingCommentLike Insert commentLike,date Comments to the dataBase:")
		fmt.Printf("err: %v\n", err)
	}

	count := 0
	row := Db.QueryRow("SELECT COUNT (*) FROM dislikes WHERE comment_id = ?", commentid)
	err = row.Scan(&count)
	if err != nil {
		fmt.Println("Error Select Count in AddingCommentLike in Database:")
		fmt.Println(err)

	}
	fmt.Printf("count in database: %v\n", count)

	_, err = Db.Exec("UPDATE comments SET commentDislike = ? WHERE commentid = ?", count, commentid)
	if err != nil {
		fmt.Println("Error function AddingCountLike Insert countLikes Posts to the dataBase:")
		fmt.Printf("err: %v\n", err)

	}

}

/*************************** CATEGORIE 1 FEED POST **********************************/
func Categorie1FeedPost(userName string) []structure.Categorie1FeedPost {

	rows, err := Db.Query("SELECT * FROM posts WHERE categories = ? OR categories2 = ?", "cat1", "cat1")
	if err != nil {
		fmt.Println("Error in ProfilFeed Function Query didn't work in dataBase:")
		log.Fatal(err)
	}
	var Posts []structure.Categorie1FeedPost

	var id int

	var postID, name, message, dateTime, image, picture, categories, categories2 string
	var NumberOfComment, NumberOfLikes, NumberOfDislikes int
	for rows.Next() {

		err := rows.Scan(&id, &postID, &image, &name, &message, &dateTime, &picture, &NumberOfComment, &NumberOfLikes, &NumberOfDislikes, &categories, &categories2)
		if err != nil {
			fmt.Println("Error function AddingCommentLike Insert commentLike,date Comments to the dataBase:")
			fmt.Printf("err: %v\n", err)
		}

		Posts = preappendCategorie1FeedPost(Posts, structure.Categorie1FeedPost{
			PostID:           postID,
			Name:             userName,
			UserImage:        image,
			Message:          message,
			DateTime:         dateTime,
			Picture:          picture,
			NumberOfComment:  LenUserComment(postID),
			NumberOfLikes:    NumberOfLikes,
			NumberOfDislikes: NumberOfDislikes,
			Categories:       categories,
			Categories2:      categories2,
		})
		fmt.Printf("Posts: %v\n", Posts)
	}
	return Posts
}

/*************************** CATEGORIE 2 FEED POST **********************************/
func Categorie2FeedPost(userName string) []structure.Categorie2FeedPost {

	rows, err := Db.Query("SELECT * FROM posts WHERE categories = ? OR categories2 = ?", "cat2", "cat2")
	if err != nil {
		fmt.Println("Error in ProfilFeed Function Query didn't work in dataBase:")
		log.Fatal(err)
	}
	var Posts []structure.Categorie2FeedPost
	var id int

	var postID, name, message, dateTime, image, picture, categories, categories2 string
	var NumberOfComment, NumberOfLikes, NumberOfDislikes int
	for rows.Next() {

		err := rows.Scan(&id, &postID, &image, &name, &message, &dateTime, &picture, &NumberOfComment, &NumberOfLikes, &NumberOfDislikes, &categories, &categories2)
		if err != nil {
			fmt.Println("Error ProfilFeed Function in rows.Scan:")
			log.Fatal(err)
		}

		fmt.Printf("NumberOfComment: %v\n", NumberOfComment)

		Posts = preappendCategorie2FeedPost(Posts, structure.Categorie2FeedPost{
			PostID:           postID,
			Name:             userName,
			UserImage:        image,
			Message:          message,
			DateTime:         dateTime,
			Picture:          picture,
			NumberOfComment:  LenUserComment(postID),
			NumberOfLikes:    NumberOfLikes,
			NumberOfDislikes: NumberOfDislikes,
			Categories:       categories,
			Categories2:      categories2,
		})
		fmt.Printf("Posts: %v\n", Posts)
	}

	return Posts
}

/*************************** CATEGORIE 3 FEED POST **********************************/
func Categorie3FeedPost(userName string) []structure.Categorie3FeedPost {

	rows, err := Db.Query("SELECT * FROM posts WHERE categories = ? OR categories2 = ?", "cat3", "cat3")
	if err != nil {
		fmt.Println("Error in ProfilFeed Function Query didn't work in dataBase:")
		log.Fatal(err)
	}
	var Posts []structure.Categorie3FeedPost
	var id int

	var postID, name, message, dateTime, image, picture, categories, categories2 string
	var NumberOfComment, NumberOfLikes, NumberOfDislikes int
	for rows.Next() {

		err := rows.Scan(&id, &postID, &image, &name, &message, &dateTime, &picture, &NumberOfComment, &NumberOfLikes, &NumberOfDislikes, &categories, &categories2)
		if err != nil {
			fmt.Println("Error ProfilFeed Function in rows.Scan:")
			log.Fatal(err)
		}

		fmt.Printf("NumberOfComment: %v\n", NumberOfComment)

		Posts = preappendCategorie3FeedPost(Posts, structure.Categorie3FeedPost{
			PostID:           postID,
			Name:             userName,
			UserImage:        image,
			Message:          message,
			DateTime:         dateTime,
			Picture:          picture,
			NumberOfComment:  LenUserComment(postID),
			NumberOfLikes:    NumberOfLikes,
			NumberOfDislikes: NumberOfDislikes,
			Categories:       categories,
			Categories2:      categories2,
		})
		fmt.Printf("Posts: %v\n", Posts)
	}

	return Posts
}

func prependAdminFeedPost(x []structure.AdminFeedPost, y structure.AdminFeedPost) []structure.AdminFeedPost {
	x = append(x, structure.AdminFeedPost{})
	copy(x[1:], x)
	x[0] = y
	return x
}

func AdminFeedPost() []structure.AdminFeedPost {
	rows, err := Db.Query("SELECT * FROM notif_admin ORDER BY id")
	if err != nil {
		fmt.Println("Error in AdminFeedPost Function dataBase:")
		log.Fatal(err)
	}
	var Posts []structure.AdminFeedPost
	var id int
	var userName, userImage, notifID, date string

	for rows.Next() {

		err := rows.Scan(&id, &userName, &userImage, &notifID, &date)
		if err != nil {
			fmt.Println("Error AdminFeedPost Function in rows.Scan:")
			log.Fatal(err)
		}

		Posts = prependAdminFeedPost(Posts, structure.AdminFeedPost{
			Name:      userName,
			UserImage: userImage,
			NotifID:   notifID,
			Date:      date,
		})
	}

	return Posts

}

func AddingModoRequest(userName string, avatar string, notifid string, date string) {
	count := 0
	row := Db.QueryRow("SELECT COUNT (*) FROM notif_admin WHERE name = ?", userName)
	err := row.Scan(&count)
	if err != nil {
		fmt.Println("Error Select Count in AddingModoRequest in Database:")
		fmt.Println(err)

	}
	if count == 0 {
		_, err = Db.Exec("INSERT INTO notif_admin (name, avatar, notifid,date) VALUES (?,?,?,?)", userName, avatar, notifid, date)
		if err != nil {
			fmt.Println("Error function AddingModoRequest Insert name to the dataBase:")
			fmt.Printf("err: %v\n", err)
		}
	}

}

func DeleteAdminRequest(name string, notifID string) {

	_, err := Db.Exec("DELETE FROM notif_admin WHERE name = ? AND notifid=?", name, notifID)
	if err != nil {
		fmt.Println("Erreur lors de la suppression de la session dans la base de données, func DeleteSession:")
		log.Fatal(err)
	}
}

func SelectUserForModo(notifID string) map[string]string {

	ans := make(map[string]string)

	var userName, userImage, date string

	rows, err := Db.Query("SELECT name,avatar,date FROM notif_admin WHERE notifid = ?", notifID)
	if err != nil {
		fmt.Println("Error SELECT in SelectUserForModo")
		log.Fatal(err)
	}

	for rows.Next() {

		err := rows.Scan(&userName, &userImage, &date)
		if err != nil {
			fmt.Println("Error in the rows.Scan in the SelectUserForModo function")
			log.Fatal(err)
		}
	}

	ans["name"] = userName
	ans["avatar"] = userImage
	ans["notifID"] = notifID
	ans["date"] = date

	return ans

}

func AddingUser2Modo(userName string) {

	_, err := Db.Exec("UPDATE users SET moderateur = ? WHERE name = ?", 1, userName)
	if err != nil {
		fmt.Println("Error function AddingUser2Modo Insert modo USERS to the dataBase:")
		fmt.Printf("err: %v\n", err)
	}

}

func LenUserModerateur() int {

	var NumberModerateur int
	err := Db.QueryRow("SELECT COUNT (*) FROM users WHERE moderateur = ?", 1).Scan(&NumberModerateur)
	if err != nil {
		fmt.Println("Error SELECT From LenUserPost dataBase:")
		log.Fatal(err)
	}

	return NumberModerateur
}

func GetAllModerateur() map[string]string {

	res := make(map[string]string)

	rows, err := Db.Query("SELECT name FROM users WHERE moderateur = ?", 1)
	if err != nil {
		fmt.Println("Error SELECT From LenUserPost dataBase:")
		log.Fatal(err)
	}

	var name string
	i := 0

	for rows.Next() {
		err := rows.Scan(&name)
		if err != nil {
			fmt.Println("Error rows.Scan() function GetAllModerateur")
			log.Fatal(err)
		}

		i++

		res["moderateur"+strconv.Itoa(i)] = name
	}

	return res

}

func DeleteModerateur(name string) {

	_, err := Db.Exec("DELETE FROM users WHERE name = ?", name)
	if err != nil {
		fmt.Println("Erreur lors de la suppression de la session dans la base de données, func DeleteSession:")
		log.Fatal(err)
	}

	_, err = Db.Exec("UPDATE users SET moderateur = ? WHERE name = ?", 0, name)
	if err != nil {
		fmt.Println("Error function DeleteModerateur Update moderateur:")
		fmt.Printf("err: %v\n", err)

	}
}

func DeletePost(postID string) {

	_, err := Db.Exec("DELETE FROM posts WHERE postid = ?", postID)
	if err != nil {
		fmt.Println("Erreur lors de la suppression de la session dans la base de données, func DeleteSession:")
		log.Fatal(err)
	}
}

func NotifComment(postID string) {

	rows, err := Db.Query("SELECT commentid FROM comments WHERE post_id=?", postID)
	if err != nil {
		panic(err)
	}
	var commentID string
	for rows.Next() {

		err := rows.Scan(&commentID)
		if err != nil {
			panic(err)
		}

	}
	rows2, err := Db.Query("SELECT name FROM comments WHERE commentid = ?", commentID)
	if err != nil {
		panic(err)

	}
	var name string
	for rows2.Next() {
		err := rows2.Scan(&name)
		if err != nil {
			panic(err)

		}
	}
	rows3, err := Db.Query("SELECT image FROM users WHERE name=?", name)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Name of the user who commented : %v\n", name)

	var picture string

	for rows3.Next() {
		err := rows3.Scan(&picture)
		if err != nil {
			panic(err)
		}
	}

	r, err := Db.Query("SELECT name FROM posts WHERE postid=?", postID)
	if err != nil {
		panic(err)
	}
	var namePost string
	for r.Next() {
		err := r.Scan(&namePost)
		if err != nil {
			panic(err)
		}
	}
	fmt.Printf("Name of the post : %v\n", namePost)

	dateTime := time.Now().Format("1-Janv-2006 15:04")
	_, err = Db.Exec("INSERT INTO notification (notifid,username, avatar, datetime, post_id, comment_id, added_comment) VALUES (?,?,?,?,?,?,?)", script.GenerateRandomString(), namePost, picture, dateTime, postID, commentID, true)
	if err != nil {
		fmt.Println("Error function NotifComment dataBase:")
		fmt.Printf("err: %v\n", err)
	}
}

func NotifLike(postID string) {

	rows, err := Db.Query("SELECT username FROM likes WHERE post_id=?", postID)
	if err != nil {
		panic(err)
	}
	var name string
	for rows.Next() {

		err := rows.Scan(&name)
		if err != nil {
			panic(err)
		}

	}
	rows2, err := Db.Query("SELECT image FROM users WHERE name=?", name)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Name user who liked the post: %v\n", name)

	var picture string

	for rows2.Next() {
		err := rows2.Scan(&picture)
		if err != nil {
			panic(err)
		}
	}

	r, err := Db.Query("SELECT name FROM posts WHERE postid=?", postID)
	if err != nil {

		panic(err)
	}
	var namePost string
	for r.Next() {
		err := r.Scan(&namePost)
		if err != nil {
			panic(err)
		}
	}
	fmt.Printf("Name of the post : %v\n", namePost)

	dateTime := time.Now().Format("1-Janv-2006 15:04")
	_, err = Db.Exec("INSERT INTO notification (notifid,username, avatar, datetime, post_id, like_post) VALUES (?,?,?,?,?,?)", script.GenerateRandomString(), namePost, picture, dateTime, postID, true)
	if err != nil {
		fmt.Println("Error function NotifLike dataBase:")
		fmt.Printf("err: %v\n", err)
	}
}

// username, avatar, datetime, post_id, comment_id, like_post, dislike_post, like_comment, dislike_comment, added_comment
func NotifLikeComment(commentID string) {

	rows, err := Db.Query("SELECT username FROM likes WHERE comment_id=?", commentID)
	if err != nil {
		panic(err)
	}
	var name string
	for rows.Next() {

		err := rows.Scan(&name)
		if err != nil {
			panic(err)
		}

	}
	rows2, err := Db.Query("SELECT image FROM users WHERE name=?", name)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Name of the user who liked the comment: %v\n", name)

	var picture string

	for rows2.Next() {
		err := rows2.Scan(&picture)
		if err != nil {
			panic(err)
		}
	}

	r, err := Db.Query("SELECT name FROM comments WHERE commentid=?", commentID)
	if err != nil {
		panic(err)
	}
	var namePost string
	for r.Next() {
		err := r.Scan(&namePost)
		if err != nil {
			panic(err)
		}
	}
	fmt.Printf("Name of the post : %v\n", namePost)

	dateTime := time.Now().Format("1-Janv-2006 15:04")
	_, err = Db.Exec("INSERT INTO notification (notifid,username, avatar, datetime, comment_id,like_comment) VALUES (?,?,?,?,?,?)", script.GenerateRandomString(), namePost, picture, dateTime, commentID, true)
	if err != nil {
		fmt.Println("Error function NotifLikeCommment dataBase:")
		fmt.Printf("err: %v\n", err)
	}
}

func NotifDisLike(postID string) {

	rows, err := Db.Query("SELECT username FROM dislikes WHERE post_id=?", postID)
	if err != nil {
		panic(err)
	}
	var name string
	for rows.Next() {

		err := rows.Scan(&name)
		if err != nil {
			panic(err)
		}
	}

	rows2, err := Db.Query("SELECT image FROM users WHERE name=?", name)
	if err != nil {
		panic(err)
	}

	var picture string

	for rows2.Next() {
		err := rows2.Scan(&picture)
		if err != nil {
			panic(err)
		}
	}

	r, err := Db.Query("SELECT name FROM posts WHERE postid=?", postID)
	if err != nil {
		panic(err)
	}
	var namePost string
	for r.Next() {
		err := r.Scan(&namePost)
		if err != nil {
			panic(err)
		}
	}
	fmt.Printf("Name of the post : %v\n", namePost)

	dateTime := time.Now().Format("1-Janv-2023 15:04")
	_, err = Db.Exec("INSERT INTO notification (notifid,username, avatar, datetime, post_id, dislike_post) VALUES (?,?,?,?,?,?)", script.GenerateRandomString(), namePost, picture, dateTime, postID, true)
	if err != nil {
		fmt.Println("Error function NotifDislike dataBase:")
		fmt.Printf("err: %v\n", err)
	}

}
func NotifDisLikeComment(commentID string) {

	rows, err := Db.Query("SELECT username FROM dislikes WHERE comment_id=?", commentID)
	if err != nil {
		panic(err)
	}
	var name string
	for rows.Next() {

		err := rows.Scan(&name)
		if err != nil {
			panic(err)
		}
	}

	rows2, err := Db.Query("SELECT image FROM users WHERE name=?", name)
	if err != nil {
		panic(err)
	}

	var picture string

	for rows2.Next() {
		err := rows2.Scan(&picture)
		if err != nil {
			panic(err)
		}
	}
	dateTime := time.Now().Format("1-Janv-2023 15:04")
	var dislikeComment bool = true

	r, err := Db.Query("SELECT username FROM notification WHERE comment_id=?", commentID)
	if err != nil {
		panic(err)
	}
	var nameNotif string
	for r.Next() {
		err := r.Scan(&nameNotif)
		if err != nil {
			panic(err)
		}
	}

	_, err = Db.Exec("INSERT INTO notification (notifid,username,avatar, datetime,comment_id,dislike_comment) VALUES (?,?,?,?,?,?)", script.GenerateRandomString(), nameNotif, picture, dateTime, commentID, dislikeComment)
	if err != nil {
		fmt.Println("Error function NotifDislikeComment dataBase:")
		fmt.Printf("err: %v\n", err)
	}
}

func reverseApppend(s []structure.Notification, e structure.Notification) []structure.Notification {
	s = append(s, e)
	copy(s[1:], s)
	s[0] = e
	return s
}

// a function that return the type of notification
func GetTypeNotif(notif structure.Notification) string {
	action := ""
	if notif.LikePost {
		action = "liked your Post"

	} else if notif.DislikePost {
		action = "disliked your Post"

	} else if notif.LikeComment {
		action = "liked your Comment"

	} else if notif.DislikeComment {
		action = "disliked your Comment"

	} else if notif.AddedComment {
		action = "added a Comment on your Post"

	}
	return action
}

func GetUserNotif() []structure.Notification {
	rows, err := Db.Query("SELECT * FROM notification WHERE username=? ORDER BY datetime DESC", GetUserProfil()["name"])
	if err != nil {
		panic(err)
	}
	var Notifs []structure.Notification

	var notifID, name, picture, date, postID, commentID, action string
	var likePost, dislikePost, likeComment, dislikeComment, addedComment bool

	for rows.Next() {

		var id int
		err := rows.Scan(&id, &notifID, &name, &picture, &date, &postID, &commentID, &likePost, &dislikePost, &likeComment, &dislikeComment, &addedComment, &action)
		if err != nil {
			panic(err)
		}

		Notifs = reverseApppend(Notifs, structure.Notification{
			NotifID:        notifID,
			UserName:       name,
			UserAvatar:     picture,
			Date:           date,
			PostID:         postID,
			CommentID:      commentID,
			LikePost:       likePost,
			DislikePost:    dislikePost,
			LikeComment:    likeComment,
			DislikeComment: dislikeComment,
			AddedComment:   addedComment,
			Action: GetTypeNotif(structure.Notification{
				LikePost:       likePost,
				DislikePost:    dislikePost,
				LikeComment:    likeComment,
				DislikeComment: dislikeComment,
				AddedComment:   addedComment,
			}),
		})
	}
	return Notifs
}

func LenUserNotif(username string) int {
	var count int
	row := Db.QueryRow("SELECT COUNT(*) FROM notification WHERE username = ?", username)
	err := row.Scan(&count)
	if err != nil {
		panic(err)
	}
	return count
}

func DeleteNotif(notifID string) {
	_, err := Db.Exec("DELETE FROM notification WHERE notifid=?", notifID)
	if err != nil {
		fmt.Println("Error function DeleteNotif dataBase:")
		fmt.Printf("err: %v\n", err)
	}
}
