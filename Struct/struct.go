package structure

type UserAccount struct {
	Id       int
	Name     string //Name of the user
	Image    string //Path src
	Email    string
	UUID     string
	Password string
	Admin    bool // true: the user is Admin
}

type Comment struct {
	Name      string
	CommentID string
	Content   string
}

type Post struct {
	PostID   string
	Name     string
	Message  string
	DateTime string
	Picture  string
	UUID     string
}

type Like struct {
	Name     string
	DateTime string
}

type AuthGoogle struct {
	Access_Token  string `json:"access_token"`
	Expires_In    int    `json:"expires_in"`
	Refresh_Token string `json:"refresh_token"`
	Id_Token      string `json:"id_token"`
	Scope         string `json:"scope"`
	Token_Type    string `json:"token_type"`
}

type GoogleUser struct {
	Name           string `json:"name"`
	Picture        string `json:"picture"`
	Email          string `json:"email"`
	Email_Verified string `json:"email_verified"`
}
