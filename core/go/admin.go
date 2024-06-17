package forum

import (
	"fmt"
	"forum/core/structs"
	"html/template"
	"net/http"
)

type Admin struct {
	User        structs.User
	ListUser    []structs.User
	IsConnected bool
}

// ServeHTTP handles the HTTP requests for the Admin struct
func (a *Admin) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// initialization of the database
	db, err := initDB()
	if err != nil {
		return
	}
	defer db.Close()

	var t *template.Template

	// We verify if th user is connected and has cookie
	if verifyCookie(r) {
		// We get the list of all created users in the database
		users, err := getAllUsers(db)

		if err != nil {
			http.Error(w, "Erreur lors de la récupération des utilisateurs ", http.StatusInternalServerError)
			fmt.Println(err)
			return
		}

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
		a.User = userSession

		a.ListUser = users
		a.IsConnected = true

		// Processes POST requests to change a user's role
		if r.Method == "POST" {
			action := r.FormValue("action")
			// If the action if modify_role, it change the role of the user in the database
			if action == "modify_role" {
				username := r.FormValue("username")
				err := modifyRole(db, username)
				if err != nil {
					http.Error(w, "Erreur lors de la modification du rôle", http.StatusInternalServerError)
					fmt.Println(err)
					return
				}
				http.Redirect(w, r, "/panel_admin", http.StatusFound)
				return
			}
		}
		// If the user is not connected, he is redirected to the main page
	} else {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	t, _ = template.ParseFiles("src/html/panel_admin.html")
	t.Execute(w, a)
}
