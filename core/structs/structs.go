package structs

type User struct {
	Email       string
	Username    string
	Password    string
	IsConnected bool
	Role        string
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
	Post        Post
	Categories  []Category
	User        User
	IsConnected bool
}

type Comment struct {
	ID     string
	PostID string
	User   string
	Text   string
}

type PostWithComments struct {
	Post        Post
	Comments    []Comment
	User        User
	IsConnected bool
	Likes       int
	Dislikes    int
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

type Like struct {
	ID     int
	PostID string
	User   string
	IsLike bool
}
