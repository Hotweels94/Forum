package forum

import (
	"forum/core/structs"
	"html/template"
	"net/http"
	"strings"

	"github.com/gofrs/uuid"
)

type Login struct {
	user structs.User
}

// We create the map for the user Sessions
var userSessions = make(map[string]structs.User)

// ServeHTTP handles the HTTP requests for the Login struct
func (l *Login) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// initialization of the database
	db, err := initDB()
	if err != nil {
		return
	}
	defer db.Close()

	var t *template.Template

	// We verify if th user is connected and has cookie
	if !verifyCookie(r) {
		if r.Method == "POST" {
			// We get the data from user when he login
			l.user.Username = strings.TrimSpace(r.FormValue("username or email"))
			l.user.Email = strings.TrimSpace(r.FormValue("username or email"))
			l.user.Password = r.FormValue("password")

			// We verify if the data (username OR email) and the password are in the database
			userData, isConnected := verifyLog(db, l.user.Username, l.user.Email, l.user.Password)
			// If he isn't connected
			if isConnected == nil {
				// We create the unique session Token and fill the User struct
				sessionToken, _ := uuid.NewV4()

				userSessions[sessionToken.String()] = structs.User{
					Username: userData.Username,
					Email:    userData.Email,
					Role:     userData.Role,
				}

				// We create the cookie and redirect the user to his profile
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
	t.Execute(w, l)
}
