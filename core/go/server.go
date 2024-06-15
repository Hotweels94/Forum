package forum

import (
	"fmt"
	"forum/core/structs"
	"html/template"
	"net/http"
	"strconv"
)

// http://localhost:8080/ -> Forum Jeux video

type mainInfo struct {
	PageName      string
	NumberOfVisit int
	User          structs.User
	IsConnected   bool
}

// ServeHTTP handles the HTTP requests for the mainInfo struct
func (m *mainInfo) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// initialization of the database
	db, err := initDB()
	if err != nil {
		return
	}
	defer db.Close()

	// We verify if th user is connected and has cookie
	if verifyCookie(r) {
		m.IsConnected = true
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
		m.User = userSession
	} else {
		m.IsConnected = false
	}

	m.NumberOfVisit++
	m.PageName = "Il y a " + strconv.Itoa(m.NumberOfVisit) + " visites !!"
	t, _ := template.ParseFiles("src/html/index.html")
	t.Execute(w, m)

}

// HandleForum handles the routing and server configuration for the forum
func HandleForum() {

	http.Handle("/src/", http.StripPrefix("/src/", http.FileServer(http.Dir("src"))))
	http.Handle("/databases/upload_image/", http.StripPrefix("/databases/upload_image/", http.FileServer(http.Dir("databases/upload_image"))))

	http.Handle("/", new(mainInfo))
	http.Handle("/register", new(Register))
	http.Handle("/login", new(Login))
	http.Handle("/profile", new(Profil))
	http.Handle("/post", new(Posts))
	http.Handle("/category", new(Categories))
	http.Handle("/list_category", new(Categories))
	http.Handle("/list_post", new(list_Post))
	http.Handle("/panel_admin", new(Admin))
	http.Handle("/report", new(Posts))
	http.Handle("/user_posts", new(list_Post))

	// Start the server and listen on port 8080
	http.ListenAndServe(":8080", nil)
}
