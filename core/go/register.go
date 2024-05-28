package forum

import (
	"forum/core/structs"
	"html/template"
	"net/http"
	"strings"
)

type Register struct {
	user        structs.User
	isConnected bool
}

func (reg *Register) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	db, err := initDB()
	if err != nil {
		return
	}
	defer db.Close()

	var t *template.Template
	if r.URL.Path == "/register" {
		if !verifyCookie(r) {
			if r.Method == "POST" {
				reg.user.Username = strings.TrimSpace(r.FormValue("username"))
				reg.user.Email = strings.TrimSpace(r.FormValue("email"))
				reg.user.Password = r.FormValue("password")

				insertUser(db, reg.user.Email, reg.user.Username, reg.user.Password)

			}
		} else {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		t, _ = template.ParseFiles("src/html/register.html")
	}
	t.Execute(w, reg)
}
