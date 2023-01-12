package structure

type UserAccount struct {
	Name     string //Name of the user
	Image    string //Path src
	Email    string
	UUID     string
	Password string
	Admin    bool // true: the user is Admin
}

type Post struct {
	Name    string //
	PostID  string //
	Content string //
}

type Comment struct {
	CommentID string
	Name      string
	Message   string
	DateTime  string
}

type Like struct {
	Name     string
	DateTime string
}
