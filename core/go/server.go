package forum

import (
	"forum/core/structs"
	"html/template"
	"net/http"
	"strconv"
)

// http://localhost:8080/ -> Forum Jeux video

type mainInfo struct {
	PageName      string
	NumberOfVisit int
	User          structs.User
	IsConnected   bool
}

func (m *mainInfo) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	db, err := initDB()
	if err != nil {
		return
	}
	defer db.Close()

	if verifyCookie(r) {
		m.IsConnected = true
		m.User = userSession
	} else {
		m.IsConnected = false
	}

	m.NumberOfVisit++
	m.PageName = "Il y a " + strconv.Itoa(m.NumberOfVisit) + " visites !!"
	t, _ := template.ParseFiles("src/html/index.html")
	t.Execute(w, m)

}

func HandleForum() {

	http.Handle("/src/", http.StripPrefix("/src/", http.FileServer(http.Dir("src"))))
	http.Handle("/databases/upload_image/", http.StripPrefix("/databases/upload_image/", http.FileServer(http.Dir("databases/upload_image"))))

	http.Handle("/", new(mainInfo))
	http.Handle("/register", new(Register))
	http.Handle("/login", new(Login))
	http.Handle("/profile", new(Profil))
	http.Handle("/post", new(Posts))
	http.Handle("/category", new(Categories))
	http.Handle("/list_category", new(Categories))
	http.Handle("/list_post", new(list_Post))
	http.Handle("/panel_admin", new(Admin))
	http.Handle("/report", new(Posts))
	http.Handle("/user_posts", new(list_Post))
	http.ListenAndServe(":8080", nil)
}
