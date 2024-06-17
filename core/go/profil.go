package forum

import (
	"fmt"
	"forum/core/structs"
	"html/template"
	"net/http"
	"strings"
)

type Profil struct {
	User        structs.User
	IsConnected bool
}

var userSession structs.User
var exists bool

// ServeHTTP handles the HTTP requests for the Profil struct
func (p *Profil) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// initialization of the database
	db, err := initDB()
	if err != nil {
		return
	}
	defer db.Close()

	var t *template.Template

	// We verify if th user is connected and has cookie
	if verifyCookie(r) {
		// We get the cookie
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
		// Get the session Token from the cookie
		sessionToken := cookie

		// Verify if the session user exist in the map of user Sessions
		userSession, exists = userSessions[sessionToken]
		if !exists {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		p.IsConnected = true
		p.User = userSession

		if r.Method == "POST" {
			action := r.FormValue("action")

			// Action to modify the username
			if action == "Modifier votre pseudo" {
				oldUsername := getUsername(db, p.User.Username)
				p.User.Username = strings.TrimSpace(r.FormValue("username"))
				err := modifyUsername(db, p.User.Username, oldUsername)
				if err != nil {
					fmt.Println(err)
				}
			}
			// Action to modify the email
			if action == "Modifier votre email" {
				oldEmail := getEmail(db, p.User.Email)
				p.User.Email = strings.TrimSpace(r.FormValue("email"))
				err := modifyEmail(db, p.User.Email, oldEmail)
				if err != nil {
					fmt.Println(err)
				}
			}

			// Action to modify to disconect from the Forum (logout)
			if action == "logout" {
				if err != nil {
					fmt.Println(err)
				}
				p.User.Role = ""
				DeleteCookie(w, "session_token")
				http.Redirect(w, r, "/login", http.StatusFound)
			}
		}
	} else {
		p.IsConnected = false
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	t, _ = template.ParseFiles("src/html/profile.html")
	t.Execute(w, p)
}
