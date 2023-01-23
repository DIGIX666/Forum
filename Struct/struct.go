package structure

type UserAccount struct {
	Id       int
	Name     string //Name of the user
	Image    string //Path src
	Email    string
	UUID     string
	Password string
	Post     []Post
	Comment  []Comment
	Like     []Like
	Admin    bool // true: the user is the Admin
}

type Comment struct {
	Name      string
	CommentID string
	Message   string
	DateTime  string
	Picture   string
	// Content   string
}

type Post struct {
	PostID   string
	Name     string
	Message  string
	DateTime string
	Picture  string
	Comment  Comment
	Like     Like
	Dislike  Dislike
}

type Like struct {
	Name     string
	Number   int
	DateTime string
}
type Dislike struct {
	Name     string
	Number   int
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

type AuthGitHub struct {
	Access_Token string `json:"access_token"`
	Scope        string `json:"scope"`
	Token_Type   string `json:"token_type"`
}

type GithubUser struct {
	Avatar_Url string `json:"avatar_url"`
	Name       string `json:"name"`
	Email      string `json:"email"`
}
