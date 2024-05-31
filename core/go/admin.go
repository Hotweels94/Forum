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
		for i := 0; i < 10; i++ {
			userAtI := getAllUserDataById(db, i)
			a.ListUser = append(a.ListUser, userAtI)
		}
	} else {
		http.Redirect(w, r, "/profile", http.StatusFound)
		return
	}
	fmt.Println(a.ListUser)

	t, _ = template.ParseFiles("src/html/panel_admin.html")
	t.Execute(w, a)
}
