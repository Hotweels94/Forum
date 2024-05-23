package forum

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

type user struct {
	Email    string
	Username string
	Password string
}

func initDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./databases/forum.db")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT NOT NULL UNIQUE,
			username TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL
		)
	`)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func insertUser(db *sql.DB, email string, username string, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO users (email, username, password) VALUES(?, ?, ?)", email, username, hashedPassword)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			fmt.Println("Username ou Email déjà connu.")
		} else {
			fmt.Println(err)
		}
	}
	return err
}

func verifyLog(db *sql.DB, username string, email string, password string) bool {
	var hashedPassword string
	err := db.QueryRow("SELECT password FROM users WHERE username = ? OR email = ?", username, email).Scan(&hashedPassword)
	if err != nil {
		return false
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func modifyLog(db *sql.DB, newUsername string, oldUsername string) error {
	_, err := db.Exec("UPDATE users SET username = ? WHERE username = ?", newUsername, oldUsername)
	if err != nil {
		fmt.Println("Error updating user:", err)
	}
	return err
}

func (u user) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	db, err := initDB()
	if err != nil {
		return
	}
	defer db.Close()

	var t *template.Template

	if r.URL.Path == "/register" {
		if r.Method == "POST" {
			u.Username = strings.TrimSpace(r.FormValue("username"))
			u.Email = strings.TrimSpace(r.FormValue("email"))
			u.Password = r.FormValue("password")

			insertUser(db, u.Email, u.Username, u.Password)
			fmt.Println(u.Email, u.Username, u.Password)
		}
		t, _ = template.ParseFiles("src/html/register.html")
	}

	if r.URL.Path == "/login" {
		if r.Method == "POST" {
			u.Username = strings.TrimSpace(r.FormValue("username or email"))
			u.Email = strings.TrimSpace(r.FormValue("username or email"))
			u.Password = r.FormValue("password")

			if verifyLog(db, u.Username, u.Email, u.Password) {

				CreateCookie(w, "username", u.Username)
				CreateCookie(w, "email", u.Email)

				http.Redirect(w, r, "/profile", http.StatusFound)
				return
			}
		}
		t, _ = template.ParseFiles("src/html/login.html")
	}

	if r.URL.Path == "/profile" {
		u.Username, _ = getCookie(r, "username")
		u.Email, _ = getCookie(r, "email")

		fmt.Println(u.Username + "/" + u.Email)

		if r.Method == "POST" {
			action := r.FormValue("action")

			if action == "Modifier" {
				oldUsername, _ := getCookie(r, "username")
				u.Username = strings.TrimSpace(r.FormValue("username"))
				u.Email = strings.TrimSpace(r.FormValue("email"))
				err := modifyLog(db, u.Username, oldUsername)
				if err == nil {
					CreateCookie(w, "username", u.Username)
				}

			}

			if action == "logout" {
				DeleteCookie(w, "username")
				DeleteCookie(w, "email")
				http.Redirect(w, r, "/login", http.StatusFound)
			}
		}

		t, _ = template.ParseFiles("src/html/profile.html")
	}

	t.Execute(w, u)

}

func CreateCookie(w http.ResponseWriter, name string, value string) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
		Secure:   true,
	}
	http.SetCookie(w, cookie)
}

func DeleteCookie(w http.ResponseWriter, name string) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
	}
	http.SetCookie(w, cookie)
}

func getCookie(r *http.Request, name string) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", err
	}
	return cookie.Value, err
}
