package data

import (
	script "Forum/scripts"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func DataBaseRegister(email string, password string) {
	db, err := sql.Open("sqlite3", "./usersForum.db")
	if err != nil {
		log.Fatal(err)

	}

	uuid := ""

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY AUTOINCREMENT, email TEXT, password TEXT, UUID TEXT)")
	if err != nil {
		log.Fatal(err)
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", email).Scan(&count)
	if err != nil {
		fmt.Println("email already exist !!")
	}
	if count > 0 {

		/*log.Fatal("email adress already exist !")
		return*/

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

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY AUTOINCREMENT, email TEXT, password TEXT, UUID TEXT)")
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

	fmt.Printf("uuid: %v\n", uuid)

	if compare {
		_, err = db.Exec("UPDATE users SET UUID = ? WHERE email = ?", uuid, email)
		if err != nil {
			fmt.Println(err)

		}
		return true
	} else {
		return false
	}

}
