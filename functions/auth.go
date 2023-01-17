package function

import (
	structure "Forum/Struct"
	dataBase "Forum/data"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gofrs/uuid"
)

func GoogleAuthLog(code string) (bool, string) {

	fmt.Printf("code: %v\n", code)

	data := url.Values{}
	data.Set("client_id", "760601264616-u9vo4s8hdistvmn6ia2goko3m6qhmff8.apps.googleusercontent.com")
	data.Set("client_secret", "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")
	data.Set("code", code)
	data.Set("redirect_uri", "https://localhost:8080/login")
	data.Set("grant_type", "authorization_code")

	responseGoogle, err := http.Post("https://oauth2.googleapis.com/token", "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		log.Fatal(err)
	}

	defer responseGoogle.Body.Close()

	var googleTokenJSON structure.AuthGoogle

	err = json.NewDecoder(responseGoogle.Body).Decode(&googleTokenJSON)
	if err != nil {
		log.Fatal(err)
	}
	a, _ := ioutil.ReadAll(responseGoogle.Body)
	fmt.Printf("ResponseGoogle: %v\n", string(a))

	fmt.Printf("googleTokenJSON: %v\n", googleTokenJSON)

	fmt.Printf("googleTokenJSON.Access_Token: %v\n", googleTokenJSON.Access_Token)
	fmt.Printf("googleTokenJSON.Scope: %v\n", googleTokenJSON.Scope)
	fmt.Printf("googleTokenJSON.Id_Token: %v\n", googleTokenJSON.Id_Token)
	fmt.Printf("googleTokenJSON.Expires_In: %v\n", googleTokenJSON.Expires_In)
	fmt.Printf("googleTokenJSON.Refresh_Token: %v\n", googleTokenJSON.Refresh_Token)
	//Rfresh_Token := googleTokenJSON.Refresh_Token
	//refresh_token := "1//03141UoOFJOiJCgYIARAAGAMSNwF-L9Irjnoum5-ga4HAMEgCNKgxA4GUcxt90qDVCa23nw0ZLZfHUDB7FJ7_JV08LIUCQSBc4r4"
	//fmt.Printf("refresh Token: %v", Rfresh_Token)
	//fmt.Printf("googleTokenJSON.Token_Type: %v\n", googleTokenJSON.Token_Type)

	googleAuthResponse, err := http.Get("https://www.googleapis.com/oauth2/v3/tokeninfo?id_token=" + googleTokenJSON.Id_Token)
	if err != nil {
		log.Fatal(err)
	}

	defer googleAuthResponse.Body.Close()
	var googleUser structure.GoogleUser
	err = json.NewDecoder(googleAuthResponse.Body).Decode(&googleUser)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("googleUser.Name: %v\n", googleUser.Name)
	fmt.Printf("googleUser.Picture: %v\n", googleUser.Picture)

	fmt.Printf("googleUserEmail.Email: %v\n", googleUser.Email)
	fmt.Printf("googleUserEmail.Email_Verified: %v\n", googleUser.Email_Verified)

	db, err := sql.Open("sqlite3", "./usersForum.db")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("UPDATE users SET NAME = ?, IMAGE = ?  WHERE email = ?", googleUser.Name, googleUser.Picture, googleUser.Email)
	if err != nil {
		fmt.Println("Error in the login Handle, sql Exec setting name, image with email:")
		fmt.Println(err)

	}

	uuidGenerated, _ := uuid.NewV4()
	uuidGoogleUser := uuidGenerated.String()

	googleUserLogged := dataBase.CheckGoogleUserLogin(googleUser.Email, googleUser.Email_Verified, uuidGoogleUser)

	fmt.Printf("googleUserLogged: %v\n", googleUserLogged)

	return googleUserLogged, uuidGoogleUser

}

func GoogleAuthRegister(code string, hashPassword string) (bool, string) {

	fmt.Printf("code: %v\n", code)

	data := url.Values{}
	data.Set("client_id", "760601264616-u9vo4s8hdistvmn6ia2goko3m6qhmff8.apps.googleusercontent.com")
	data.Set("client_secret", "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")
	data.Set("code", code)
	data.Set("redirect_uri", "https://localhost:8080/register")
	data.Set("grant_type", "authorization_code")

	responseGoogle, err := http.Post("https://oauth2.googleapis.com/token", "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		log.Fatal(err)
	}

	defer responseGoogle.Body.Close()

	var googleTokenJSON structure.AuthGoogle

	err = json.NewDecoder(responseGoogle.Body).Decode(&googleTokenJSON)
	if err != nil {
		log.Fatal(err)
	}
	a, _ := ioutil.ReadAll(responseGoogle.Body)
	fmt.Printf("ResponseGoogle: %v\n", string(a))

	fmt.Printf("googleTokenJSON: %v\n", googleTokenJSON)

	fmt.Printf("googleTokenJSON.Access_Token: %v\n", googleTokenJSON.Access_Token)
	fmt.Printf("googleTokenJSON.Scope: %v\n", googleTokenJSON.Scope)
	fmt.Printf("googleTokenJSON.Id_Token: %v\n", googleTokenJSON.Id_Token)
	fmt.Printf("googleTokenJSON.Expires_In: %v\n", googleTokenJSON.Expires_In)
	fmt.Printf("googleTokenJSON.Refresh_Token: %v\n", googleTokenJSON.Refresh_Token)
	//Rfresh_Token := googleTokenJSON.Refresh_Token
	//refresh_token := "1//03141UoOFJOiJCgYIARAAGAMSNwF-L9Irjnoum5-ga4HAMEgCNKgxA4GUcxt90qDVCa23nw0ZLZfHUDB7FJ7_JV08LIUCQSBc4r4"
	//fmt.Printf("refresh Token: %v", Rfresh_Token)
	//fmt.Printf("googleTokenJSON.Token_Type: %v\n", googleTokenJSON.Token_Type)

	googleAuthResponse, err := http.Get("https://www.googleapis.com/oauth2/v3/tokeninfo?id_token=" + googleTokenJSON.Id_Token)
	if err != nil {
		log.Fatal(err)
	}

	defer googleAuthResponse.Body.Close()
	var googleUser structure.GoogleUser
	err = json.NewDecoder(googleAuthResponse.Body).Decode(&googleUser)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("googleUser.Name: %v\n", googleUser.Name)
	fmt.Printf("googleUser.Picture: %v\n", googleUser.Picture)

	fmt.Printf("googleUserEmail.Email: %v\n", googleUser.Email)
	fmt.Printf("googleUserEmail.Email_Verified: %v\n", googleUser.Email_Verified)

	db, err := sql.Open("sqlite3", "./usersForum.db")
	if err != nil {
		fmt.Println("Erreur à l'ouverture de la base de donnée dans GoogleAuthRegister:")
		log.Fatal(err)
	}

	uuidGenerated, _ := uuid.NewV4()
	uuidGoogleUser := uuidGenerated.String()

	_, err = db.Exec("INSERT INTO users (name, image, email, uuid, password, admin) VALUES (?, ?, ?,?,?,?)", googleUser.Name, googleUser.Picture, googleUser.Email, uuidGoogleUser, hashPassword, false)
	if err != nil {
		log.Fatal(err)
	}

	googleUserLogged := dataBase.CheckGoogleUserLogin(googleUser.Email, googleUser.Email_Verified, uuidGoogleUser)

	fmt.Printf("googleUserLogged: %v\n", googleUserLogged)

	return googleUserLogged, uuidGoogleUser
}

func GitHubRegister() {

}
