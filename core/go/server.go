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
	http.Handle("/register", new(user))
	http.Handle("/login", new(user))
	http.Handle("/profile", new(user))
	http.Handle("/post", new(Post))
	http.Handle("/category", new(Category))
	http.Handle("/list_category", new(Category))
	http.ListenAndServe(":8080", nil)
}
