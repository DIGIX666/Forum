package data

import (
	script "Forum/scripts"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func CreateDataBase() {

	db, err := sql.Open("sqlite3", "./usersForum.db")
	if err != nil {
		log.Fatal(err)

	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users 
        (id INTEGER PRIMARY KEY AUTOINCREMENT,
        pseudo TEXT,
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

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS posts (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        pseudo TEXT,
        postid TEXT,
        content TEXT
    )`)
	if err != nil {
		log.Println("erreur creation de table posts")
		log.Fatal(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS comments (
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

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS likes (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT,
        datetime TEXT
    )`)
	if err != nil {
		log.Println("erreur creation de table likes")
		log.Fatal(err)
	}

}

func DataBaseRegister(email string, password string) {
	db, err := sql.Open("sqlite3", "./usersForum.db")
	if err != nil {
		log.Fatal(err)

	}

	uuid := ""

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", email).Scan(&count)
	if err != nil {
		fmt.Println("email already exist !!")
	}
	if count > 0 {

		fmt.Println("email adress already exist !")

	} else {
		_, err = db.Exec("INSERT INTO users (email, password, UUID) VALUES (?, ?, ?)", email, password, uuid)
		if err != nil {
			log.Fatal(err)
		}
	}

	rows, err := db.Query("SELECT id, email, password FROM users")
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

}

func DataBaseLogin(email string, password string, uuid string) bool {
	db, err := sql.Open("sqlite3", "./usersForum.db")
	if err != nil {
		log.Fatal(err)

	}

	defer db.Close()

	var hashpassword string
	err = db.QueryRow("SELECT password FROM users WHERE email = ?", email).Scan(&hashpassword)
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

func CheckUserLogin(email string, password string, uuid string) bool {
	db, err := sql.Open("sqlite3", "./usersForum.db")
	if err != nil {
		log.Fatal(err)

	}

	var hashpassword string
	err = db.QueryRow("SELECT password FROM users WHERE email = ?", email).Scan(&hashpassword)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("email: %v\n", email)
	fmt.Printf("hashpassword: %v\n", hashpassword)
	compare := script.ComparePassword(hashpassword, password)

	fmt.Printf("password of user in the dataBase? %v\n", compare)

	_, err = db.Exec("UPDATE users SET UUID = ? WHERE email = ?", uuid, email)
	if err != nil {
		fmt.Println(err)
		return false

	} else {
		return true
	}

}
