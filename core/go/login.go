package forum

import (
	"forum/core/structs"
	"html/template"
	"net/http"
	"strings"

	"github.com/gofrs/uuid"
)

type Login struct {
	user        structs.User
	isConnected bool
}

var userSessions = make(map[string]structs.User)

func (l *Login) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	db, err := initDB()
	if err != nil {
		return
	}
	defer db.Close()

	var t *template.Template

	if r.URL.Path == "/login" {
		if !verifyCookie(r) {
			if r.Method == "POST" {
				l.user.Username = strings.TrimSpace(r.FormValue("username or email"))
				l.user.Email = strings.TrimSpace(r.FormValue("username or email"))
				l.user.Password = r.FormValue("password")

				userData, isConnected := verifyLog(db, l.user.Username, l.user.Email, l.user.Password)
				if isConnected == nil {
					sessionToken, _ := uuid.NewV4()

					userSessions[sessionToken.String()] = structs.User{
						Username: userData.Username,
						Email:    userData.Email,
					}

					CreateCookie(w, "session_token", sessionToken.String())

					http.Redirect(w, r, "/profile", http.StatusFound)
					return
				}
			}
		} else {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		t, _ = template.ParseFiles("src/html/login.html")
	}
	t.Execute(w, l)
}
