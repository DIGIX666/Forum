package function

import (
	structure "Forum/Struct"
	dataBase "Forum/data"
	script "Forum/scripts"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

const DISCORD_CLIENT_ID = "1086324794837962795"
const DISCORD_CLIENT_SECRET = "EA-iR4RV_VfD5peX0r3DVSZr5XMXSTFe"
const DISCORD_OAUTH2_TOKEN = "https://discord.com/api/oauth2/token"
const REDIRECT_URI = "https://localhost:8080/register"

/* ----------------------------------- DISCORD AUTH REGISTER -----------------------------*/

func DiscordAuthRegister(code string, hashPassword string) {

	fmt.Printf("code: %v\n", code)

	data := url.Values{}
	data.Set("client_id", DISCORD_CLIENT_ID)
	data.Set("client_secret", DISCORD_CLIENT_SECRET)
	data.Set("code", code)
	data.Set("redirect_uri", REDIRECT_URI)
	data.Set("grant_type", "authorization_code")

	responseDiscord, err := http.Post(DISCORD_OAUTH2_TOKEN, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		log.Fatal(err)
	}

	defer responseDiscord.Body.Close()

	var discordTokenJSON structure.AuthDiscord

	err = json.NewDecoder(responseDiscord.Body).Decode(&discordTokenJSON)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("discordTokenJSON: %v\n", discordTokenJSON)

	client := &http.Client{}

	discordAuthResponse, err := http.NewRequest("GET", "https://discord.com/api/users/@me", nil)
	if err != nil {
		panic(err)
	}

	discordAuthResponse.Header.Set("Authorization", "Bot "+discordTokenJSON.Access_Token)
	resp, err := client.Do(discordAuthResponse)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var discordUser structure.DiscordUser
	err = json.NewDecoder(resp.Body).Decode(&discordUser)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("discordUser: %v\n", discordUser)
	fmt.Printf("hashPassword: %v\n", hashPassword)

	// err = dataBase.Db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", discordUser.Email).Scan(&count)
	// if err != nil {
	// 	fmt.Println("error reading database to found email !!")
	// }
	// if count > 0 {

	// 	// fmt.Println("google user already registered")

	// 	return false, discordUser.Email, ""

	// } else {

	// 	if discordUser.Email != "" {

	// 		_, err = dataBase.Db.Exec("INSERT INTO users (name, image, email, uuid, password, admin) VALUES (?, ?, ?,?,?,?)", discordUser.Name, discordUser.Picture, discordUser.Email, "", hashPassword, false)
	// 		if err != nil {
	// 			log.Fatal(err)
	// 		}
	// 		return true, discordUser.Email, discordUser.Name

	// 	} else {
	// 		return false, "", ""
	// 	}
	// }

}

/* ----------------------------------- GOOGLE AUTH LOG ---------------------------------- */
func GoogleAuthLog(code string) (bool, string, string, string) {

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
	// a, _ := ioutil.ReadAll(responseGoogle.Body)
	// fmt.Printf("ResponseGoogle: %v\n", string(a))

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

	count := 0

	err = dataBase.Db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", googleUser.Email).Scan(&count)
	if err != nil {
		fmt.Println("error reading database to found email !!")
	}
	if count == 1 {

		_, err = dataBase.Db.Exec("UPDATE users SET NAME = ?, IMAGE = ?  WHERE email = ?", googleUser.Name, googleUser.Picture, googleUser.Email)
		if err != nil {
			fmt.Println("Error in the GoogleAuthLog function, sql Exec setting name, image with email:")
			fmt.Println(err)
			return false, "", "", "false"

		}
		return true, googleUser.Name, googleUser.Email, googleUser.Email_Verified

	} else {

		return false, "", "", "false"

	}

}

/* ------------------------------ GOOGLE AUTH REGISTER -------------------------------- */
func GoogleAuthRegister(code string, hashPassword string) (bool, string, string) {

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

	//Rfresh_Token := googleTokenJSON.Refresh_Token
	//refresh_token := "1//03141UoOFJOiJCgYIARAAGAMSNwF-L9Irjnoum5-ga4HAMEgCNKgxA4GUcxt90qDVCa23nw0ZLZfHUDB7FJ7_JV08LIUCQSBc4r4"

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
	count := 0
	err = dataBase.Db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", googleUser.Email).Scan(&count)
	if err != nil {
		fmt.Println("error reading database to found email !!")
	}
	if count > 0 {

		// fmt.Println("google user already registered")

		return false, googleUser.Email, ""

	} else {

		if googleUser.Email != "" {

			_, err = dataBase.Db.Exec("INSERT INTO users (name, image, email, uuid, password, admin) VALUES (?, ?, ?,?,?,?)", googleUser.Name, googleUser.Picture, googleUser.Email, "", hashPassword, false)
			if err != nil {
				log.Fatal(err)
			}
			return true, googleUser.Email, googleUser.Name

		} else {
			return false, "", ""
		}
	}

}

/* ------------------------------ GITHUB AUTH REGISTER -------------------------------- */
func GitHubRegister(code string) (bool, string, string) {

	data := url.Values{}
	data.Set("client_id", "44fd70920b2db737a3ba")
	data.Set("client_secret", "d01537f316e411dbc710369e9f907f5b8a71cc9d")
	data.Set("code", code)
	data.Set("redirect_uri", "https://localhost:8080/register")

	responseGitHub, err := http.PostForm("https://github.com/login/oauth/access_token", data)

	if err != nil {
		log.Fatal(err)
	}

	if responseGitHub.StatusCode != http.StatusOK {
		log.Fatalf("Error: %v", responseGitHub.Status)
	}

	// read the response
	body, err := ioutil.ReadAll(responseGitHub.Body)
	if err != nil {
		log.Fatal(err)
	}

	// close the response
	responseGitHub.Body.Close()

	// parse the response
	values, err := url.ParseQuery(string(body))
	if err != nil {
		log.Fatal(err)
	}

	// get the token
	token := values.Get("access_token")

	client := &http.Client{}

	reqGitHubUser, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		log.Fatal(err)
	}

	reqGitHubUser.Header.Set("Authorization", "Bearer "+token)
	reqGitHubUser.Header.Set("Accept", "application/vnd.github+json")
	reqGitHubUser.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	// fmt.Println(reqGitHubUser.Header)

	responseGitHubUser, err := client.Do(reqGitHubUser)
	if err != nil {
		log.Fatal(err)
	}

	// fmt.Println(responseGitHubUser.Status)

	var githubUserJSONToken structure.GithubUser

	json.NewDecoder(responseGitHubUser.Body).Decode(&githubUserJSONToken)

	defer responseGitHubUser.Body.Close()

	// fmt.Println(githubUserJSONToken)

	count := 0

	err = dataBase.Db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", githubUserJSONToken.Email).Scan(&count)
	if err != nil {
		fmt.Println("error reading database to found email !!")
	}
	if count > 0 {

		// fmt.Println("Github user already register !")

		return false, "", githubUserJSONToken.Name

	} else {
		hashPassword := script.GenerateHash(script.GenerateRandomString())

		if githubUserJSONToken.Name != "" {
			_, err = dataBase.Db.Exec("INSERT INTO users (name, image, email, uuid, password, admin) VALUES (?, ?, ?,?,?,?)", githubUserJSONToken.Name, githubUserJSONToken.Avatar_Url, githubUserJSONToken.Email, "", hashPassword, false)
			if err != nil {
				fmt.Println("Erreur EXEC INSERT fonction GitHubRegister")
				log.Fatal(err)
			}

			return true, githubUserJSONToken.Email, githubUserJSONToken.Name
		} else {
			return false, "", ""
		}

	}

}

/* ------------------------------ GITHUB AUTH REGISTER -------------------------------- */
func GitHubLog(code string) (bool, string, string, string) {

	data := url.Values{}
	data.Set("client_id", "44fd70920b2db737a3ba")
	data.Set("client_secret", "d01537f316e411dbc710369e9f907f5b8a71cc9d")
	data.Set("code", code)
	data.Set("redirect_uri", "https://localhost:8080/register")

	responseGitHub, err := http.PostForm("https://github.com/login/oauth/access_token", data)

	if err != nil {
		log.Fatal(err)
	}

	if responseGitHub.StatusCode != http.StatusOK {
		log.Fatalf("Error: %v", responseGitHub.Status)
	}

	// read the response
	body, err := ioutil.ReadAll(responseGitHub.Body)
	if err != nil {
		log.Fatal(err)
	}

	// close the response
	responseGitHub.Body.Close()

	// parse the response
	values, err := url.ParseQuery(string(body))
	if err != nil {
		log.Fatal(err)
	}

	//fmt.Printf("string(body): %v\n", string(body))

	// get the token
	token := values.Get("access_token")
	fmt.Println("Token:", token)

	client := &http.Client{}

	reqGitHubUser, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		log.Fatal(err)
	}

	reqGitHubUser.Header.Set("Authorization", "Bearer "+token)
	reqGitHubUser.Header.Set("Accept", "application/vnd.github+json")
	reqGitHubUser.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	//fmt.Println(reqGitHubUser.Header)

	responseGitHubUser, err := client.Do(reqGitHubUser)
	if err != nil {
		log.Fatal(err)
	}

	//fmt.Println(responseGitHubUser.Status)
	var githubUserJSONToken structure.GithubUser

	json.NewDecoder(responseGitHubUser.Body).Decode(&githubUserJSONToken)

	defer responseGitHubUser.Body.Close()

	var userEmail, userName, userAvatar string

	// fmt.Printf("githubUserJSONToken.Email: %v\n", githubUserJSONToken.Email)

	count := 0

	err = dataBase.Db.QueryRow("SELECT COUNT(*) FROM users WHERE name = ?", githubUserJSONToken.Name).Scan(&count)
	if err != nil {
		fmt.Println("error reading database to found email !!")
	}
	if count == 1 {

		_, err = dataBase.Db.Exec("UPDATE users SET IMAGE = ? WHERE name = ?", githubUserJSONToken.Avatar_Url, githubUserJSONToken.Name)
		if err != nil {
			fmt.Println("Error in the GoogleAuthLog function, sql Exec setting name, image with email:")
			fmt.Println(err)
			return false, "", "", ""

		}

		return true, userName, userEmail, userAvatar

	} else {

		return false, "", "", ""
	}

}
