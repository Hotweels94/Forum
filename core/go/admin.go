package forum

import (
	"fmt"
	"forum/core/structs"
	"html/template"
	"net/http"
)

type Admin struct {
	User     structs.User
	ListUser []structs.User
}

func (a *Admin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	db, err := initDB()
	if err != nil {
		return
	}
	defer db.Close()

	var t *template.Template

	if verifyCookie(r) {
		a.User = userSession
		users, err := getAllUsers(db)

		if err != nil {
			http.Error(w, "Erreur lors de la récupération des utilisateurs ", http.StatusInternalServerError)
			fmt.Println(err)
			return
		}

		a.ListUser = users

		if r.Method == "POST" {
			action := r.FormValue("action")
			if action == "modify_role" {
				username := r.FormValue("username")
				err := modifyRoleAsModerator(db, username)
				if err != nil {
					http.Error(w, "Erreur lors de la modification du rôle", http.StatusInternalServerError)
					fmt.Println(err)
					return
				}
				http.Redirect(w, r, "/panel_admin", http.StatusFound)
				return
			}
		}

	} else {
		http.Redirect(w, r, "/profile", http.StatusFound)
		return
	}
	fmt.Println(a.ListUser)

	t, _ = template.ParseFiles("src/html/panel_admin.html")
	t.Execute(w, a)
}
