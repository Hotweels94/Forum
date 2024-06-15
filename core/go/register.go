package forum

import (
	"forum/core/structs"
	"html/template"
	"net/http"
	"strings"
)

type Register struct {
	user         structs.User
	ErrorMessage string
}

// ServeHTTP handles the HTTP requests for the Register struct
func (reg *Register) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// Initialisation of the database
	db, err := initDB()
	if err != nil {
		return
	}
	defer db.Close()

	var t *template.Template

	// If the user doesn't have cookie
	if !verifyCookie(r) {
		if r.Method == "POST" {
			// We get the data from user when he register
			reg.user.Username = strings.TrimSpace(r.FormValue("username"))
			reg.user.Email = strings.TrimSpace(r.FormValue("email"))
			reg.user.Password = r.FormValue("password")
			reg.user.Role = "user"

			reg.ErrorMessage = ""

			// We insert the user data in the database
			err := insertUser(db, reg.user.Email, reg.user.Username, reg.user.Password, reg.user.Role)
			if err != nil {
				reg.ErrorMessage = err.Error()
			} else {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
		}
		t, _ = template.ParseFiles("src/html/register.html")
		t.Execute(w, reg)
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
}
