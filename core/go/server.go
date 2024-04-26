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
	m.PageName = "Forum Jeux video !! There are : " + strconv.Itoa(m.NumberOfVisit) + " visits !!"
	t, _ := template.ParseFiles("src/html/index.html")
	t.Execute(w, m)

}

func HandleForum() {
	http.Handle("/", new(mainInfo))
	http.Handle("/test", new(web))
	http.Handle("/login", new(user))
	http.Handle("/register", new(user))
	http.ListenAndServe(":8080", nil)
}
