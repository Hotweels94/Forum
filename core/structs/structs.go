package structs

type User struct {
	Email       string
	Username    string
	Password    string
	isConnected bool
}

type Post struct {
	ID               string
	User             string
	Text             string
	Title            string
	ImageURL         string
	SelectedCategory int
}

type PostPage struct {
	Post       Post
	Categories []Category
}

type Comment struct {
	ID     string
	PostID string
	User   string
	Text   string
}

type PostWithComments struct {
	Post     Post
	Comments []Comment
}

type List_Post struct {
	Posts        []Post
	NameCategory string
}

type Category struct {
	ID          int
	Name        string
	Description string
}
