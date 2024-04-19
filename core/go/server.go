package forum

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
)

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

func HandleTest() {
	http.Handle("/", new(mainInfo))
	http.Handle("/test", new(web))
	http.ListenAndServe(":8080", nil)
}
