package forum

import (
	"fmt"
	"forum/core/structs"
	"html/template"
	"net/http"
	"strings"
)

type Profil struct {
	User structs.User
}

var userSession structs.User
var exists bool

func (p *Profil) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	db, err := initDB()
	if err != nil {
		return
	}
	defer db.Close()

	var t *template.Template

	if r.URL.Path == "/profile" {
		if verifyCookie(r) {
			cookie, err := getCookie(r, "session_token")
			if err != nil {
				if err == http.ErrNoCookie {
					// If the cookie is not set, return an unauthorized status
					w.WriteHeader(http.StatusUnauthorized)
					fmt.Println(err)
					return
				}
				// For any other type of error, return a bad request status
				w.WriteHeader(http.StatusBadRequest)
				fmt.Println(err)
				return
			}
			sessionToken := cookie

			userSession, exists = userSessions[sessionToken]
			if !exists {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			p.User = userSession
			fmt.Println(p.User.Role)

			if r.Method == "POST" {
				action := r.FormValue("action")

				if action == "Modifier votre pseudo" {
					oldUsername := getUsername(db, p.User.Username)
					p.User.Username = strings.TrimSpace(r.FormValue("username"))
					err := modifyUsername(db, p.User.Username, oldUsername)
					if err != nil {
						fmt.Println(err)
					}
				}
				if action == "Modifier votre email" {
					oldEmail := getEmail(db, p.User.Email)
					p.User.Email = strings.TrimSpace(r.FormValue("email"))
					err := modifyEmail(db, p.User.Email, oldEmail)
					if err != nil {
						fmt.Println(err)
					}
				}

				if action == "logout" {
					DeleteCookie(w, "session_token")
					http.Redirect(w, r, "/login", http.StatusFound)
				}
			}
		} else {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		t, _ = template.ParseFiles("src/html/profile.html")
	}
	t.Execute(w, p)
}
