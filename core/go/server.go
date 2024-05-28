package forum

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
)

// http://localhost:8080/ -> Forum Jeux video

type mainInfo struct {
	PageName      string
	NumberOfVisit int
}

type web struct {
	Test string
}

func (web *web) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	web.Test = "Forum Nico Ryan"
	fmt.Fprintf(w, web.Test)
}

func (m *mainInfo) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.NumberOfVisit++
	m.PageName = "Il y a " + strconv.Itoa(m.NumberOfVisit) + " visites !!"
	t, _ := template.ParseFiles("src/html/index.html")
	t.Execute(w, m)

}

func HandleForum() {

	http.Handle("/src/", http.StripPrefix("/src/", http.FileServer(http.Dir("src"))))
	http.Handle("/databases/upload_image/", http.StripPrefix("/databases/upload_image/", http.FileServer(http.Dir("databases/upload_image"))))

	http.Handle("/", new(mainInfo))
	http.Handle("/test", new(web))
	http.Handle("/register", new(Register))
	http.Handle("/login", new(Login))
	http.Handle("/profile", new(Profil))
	http.Handle("/post", new(Posts))
	http.Handle("/category", new(Categories))
	http.Handle("/list_category", new(Categories))
	http.Handle("/list_post", new(list_Post))
	http.ListenAndServe(":8080", nil)
}
