package forum

import (
	"fmt"
	"net/http"
)

type web struct {
	test string
}

func (web *web) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	web.test = "Forum Nico Ryan"
	fmt.Fprintf(w, web.test)
}

func HandleTest() {
	http.Handle("/test", new(web))
	http.ListenAndServe(":8080", nil)
}
