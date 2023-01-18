package data

import (
	script "Forum/scripts"
	"Forum/structure"
	"database/sql"
	"fmt"
	"log"

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

	_, err = Db.Exec(`CREATE TABLE IF NOT EXISTS posts (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT,
        postid TEXT,
        content TEXT
    )`)
	if err != nil {
		log.Println("erreur creation de table posts")
		log.Fatal(err)
	}

	_, err = Db.Exec(`CREATE TABLE IF NOT EXISTS comments (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        commentid TEXT,
        name TEXT,
        message TEXT,
        datetime TEXT
    )`)
	if err != nil {
		log.Println("erreur creation de table comments")
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

	/*count := 0

	if count == 0 {
		_, err = Db.Exec("INSERT INTO users (name, image, email, uuid, password, admin) VALUES (?, ?, ?,?,?,?)", "none", "../assets/images/beehive-37436.svg", "none", "none", "none", false)
		if err != nil {
			log.Fatal(err)
		}
		count++
	}

	uAccount = append(uAccount, structure.UserAccount{
		Name:     "none",
		Image:    "../assets/images/beehive-37436.svg",
		Email:    "none",
		Password: "none",
		UUID:     "none",
		Admin:    false,
	})*/

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
		fmt.Println("insert")
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
		log.Fatal(err)
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

func UserPost(userName string, message string, postID string, dateTime string) bool {

	_, err := Db.Exec("INSERT INTO comments (name, message, commentid, datetime) VALUES (?, ?, ?,?)", userName, message, postID, dateTime)
	if err != nil {
		fmt.Println("Error Insert user Post to the dataBase:")
		log.Fatal(err)
	} else {
		return true
	}

	return false

}

func GetUserProfil(uName string) interface{} {
	var ans interface{}
	var userIdDB int

	var profil structure.UserAccount

	err := Db.QueryRow("SELECT id,image, email, UUID, admin, password FROM users WHERE name = ?", uName).Scan(&userIdDB, &profil.Image, &profil.Email, &profil.UUID, &profil.Admin, &profil.Password)
	if err != nil {
		fmt.Println("Error when Selecting user profil from userForum.Db")
		log.Fatal(err)
	}
	ans = profil
	return ans

}
