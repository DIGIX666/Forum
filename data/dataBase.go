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

var Db *sql.DB

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
        content TEXT
    )`)
	if err != nil {
		log.Println("erreur creation de table comments")
		log.Fatal(err)
	}

	_, err = Db.Exec(`CREATE TABLE IF NOT EXISTS posts (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        postid TEXT,
        name TEXT,
        message TEXT,
        datetime TEXT,
		picture TEXT
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
        name TEXT,
        datetime TEXT
    )`)
	if err != nil {
		log.Println("erreur creation de table likes")
		log.Fatal(err)
	}

}

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

func UserPost(userName string, message string, postID string, dateTime string, pictureURL string) bool {

	if pictureURL != "" {

		_, err := Db.Exec("INSERT INTO posts (name, message, postid, datetime,picture) VALUES (?, ?, ?,?,?)", userName, message, postID, dateTime, pictureURL)
		if err != nil {
			fmt.Println("Error Insert user Post to the dataBase:")
			log.Fatal(err)
		} else {
			return true
		}

		return false

	} else {

		_, err := Db.Exec("INSERT INTO posts (name, message, postid, datetime,picture) VALUES (?, ?, ?,?,?)", userName, message, postID, dateTime, "")
		if err != nil {
			fmt.Println("Error Insert user Post to the dataBase:")
			log.Fatal(err)
		} else {
			return true
		}

		return false

	}

}

func SetGoogleUserUUID(userEmail string) string {

	uuidGenerated, _ := uuid.NewV4()
	uuid := uuidGenerated.String()

	_, err := Db.Exec("UPDATE users SET UUID = ? WHERE email = ?", uuid, userEmail)
	if err != nil {
		fmt.Println(err)
	}
	return uuid

}

func SetGitHubUUID(userName string) string {

	uuidGenerated, _ := uuid.NewV4()
	uuid := uuidGenerated.String()

	_, err := Db.Exec("UPDATE users SET UUID = ? WHERE name = ?", uuid, userName)
	if err != nil {
		fmt.Println(err)
	}
	return uuid

}

/*func GetUserProfil(uName string)  {

	var userIdDB int



	err := Db.QueryRow("SELECT id,image, email, UUID, admin, password FROM users WHERE name = ?", uName).Scan(&userIdDB, &profil.Image, &profil.Email, &profil.UUID, &profil.Admin, &profil.Password)
	if err != nil {
		fmt.Println("Error when Selecting user profil from userForum.Db")
		log.Fatal(err)
	}

	return

}*/

func GetUserProfil() map[string]string {

	ans := make(map[string]string, 5)
	var id int
	var name, uuid, cookie string

	err := Db.QueryRow("SELECT * FROM session ORDER BY id DESC LIMIT 1").Scan(&id, &name, &uuid, &cookie)
	if err != nil {
		fmt.Println("Erreur SELECT fonction GetUserProfil dataBase:")
		// log.Fatal(err)
	}

	var userImage, userEmail, admin string

	err = Db.QueryRow("SELECT image,email,admin FROM users WHERE name = ?", name).Scan(&userImage, &userEmail, &admin)
	if err != nil {
		fmt.Println("Erreur SELECT 2 fonction GetUserProfil dataBase:")
		// log.Fatal(err)
	}

	ans["name"] = name
	ans["email"] = userEmail
	ans["userImage"] = userImage
	ans["uuid"] = uuid
	ans["admin"] = admin

	return ans

}

func GetLastPost() map[string]string {
	ans := make(map[string]string, 5)
	var id int
	var postID, message, dataTime, name, pictureURL string

	err := Db.QueryRow("SELECT * FROM posts ORDER BY id DESC LIMIT 1").Scan(&id, &postID, &name, &message, &dataTime, &pictureURL)
	if err != nil {
		fmt.Println("Erreur SELECT fonction GetLastPost dataBase:")
		log.Fatal(err)
	}
	//ans["id"] = id
	ans["postID"] = postID
	ans["userName"] = name
	ans["message"] = message
	ans["dataTime"] = dataTime
	ans["pictureURL"] = pictureURL

	return ans

}

var posts []structure.Post

func preappendPost(c structure.Post) []structure.Post {
	posts = append(posts, structure.Post{})
	copy(posts[1:], posts)
	posts[0] = c
	return posts
}

func HomeFeed() []structure.Post {

	rows, err := Db.Query("SELECT * FROM posts ORDER BY id")
	if err != nil {
		fmt.Println("Error in HomeFeed Function dataBase:")
		log.Fatal(err)
	}
	var Posts []structure.Post

	for rows.Next() {
		var id int
		var postID, userName, message, dateTime, picture string

		err = rows.Scan(&id, &postID, &userName, &message, &dateTime, &picture)
		if err != nil {
			fmt.Println("Error HomeFeed Function in rows.Scan:")
			log.Fatal(err)
		}

		Posts = preappendPost(structure.Post{
			PostID:   postID,
			Name:     userName,
			Message:  message,
			DateTime: dateTime,
			Picture:  picture,
		})

	}

	return Posts

}

func ProfilFeed(userName string) []structure.Post {

	rows, err := Db.Query("SELECT postid,message,datetime,picture FROM posts WHERE name = ?", userName)
	if err != nil {
		fmt.Println("Error in ProfilFeed Function Query didn't work in dataBase:")
		log.Fatal(err)
	}
	var Posts []structure.Post

	for rows.Next() {

		var postID, message, dateTime, picture string

		err = rows.Scan(&postID, &message, &dateTime, &picture)
		if err != nil {
			fmt.Println("Error ProfilFeed Function in rows.Scan:")
			log.Fatal(err)
		}

		Posts = preappendPost(structure.Post{
			PostID:   postID,
			Name:     userName,
			Message:  message,
			DateTime: dateTime,
			Picture:  picture,
		})

	}

	fmt.Printf("Post in dataBase: %v\n", Posts)

	return Posts

}

func ProfilFeedDelete(userName string) {

	_, err := Db.Query("DELETE postid,message,datetime,picture FROM posts WHERE name != ?", userName)
	if err != nil {
		fmt.Println("Error in ProfilFeedDelete Function Query didn't work in dataBase:")
		log.Fatal(err)
	}

}
