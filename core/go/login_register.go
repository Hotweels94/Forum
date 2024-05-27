package forum

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/gofrs/uuid"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

type user struct {
	Email       string
	Username    string
	Password    string
	IsConnected bool
}

var userSessions = make(map[string]user)
var userInfo user

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

func verifyLog(db *sql.DB, username string, email string, password string) (user, error) {
	var hashedPassword string
	var userData user
	err := db.QueryRow("SELECT password, username, email FROM users WHERE username = ? OR email = ?", username, email).Scan(&hashedPassword, &userData.Username, &userData.Email)
	if err != nil {
		return userData, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return userData, err
}

func getUsername(db *sql.DB, username string) string {
	var userData user
	err := db.QueryRow("SELECT username FROM users WHERE username = ?", username).Scan(&userData.Username)
	if err != nil {
		return ""
	}
	return userData.Username
}

func getEmail(db *sql.DB, email string) string {
	var userData user
	err := db.QueryRow("SELECT email FROM users WHERE email = ?", email).Scan(&userData.Email)
	if err != nil {
		return ""
	}
	return userData.Email
}

func modifyUsername(db *sql.DB, newUsername string, oldUsername string) error {
	_, err := db.Exec("UPDATE users SET username = ? WHERE username = ?", newUsername, oldUsername)
	if err != nil {
		fmt.Println("Error updating user:", err)
	}
	return err
}

func modifyEmail(db *sql.DB, newEmail string, oldEmail string) error {
	_, err := db.Exec("UPDATE users SET email = ? WHERE email = ?", newEmail, oldEmail)
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

			userData, isConnected := verifyLog(db, u.Username, u.Email, u.Password)
			if isConnected == nil {
				sessionToken, _ := uuid.NewV4()

				userSessions[sessionToken.String()] = user{
					Username: userData.Username,
					Email:    userData.Email,
				}

				CreateCookie(w, "session_token", sessionToken.String())

				http.Redirect(w, r, "/profile", http.StatusFound)
				return
			}
		}
		t, _ = template.ParseFiles("src/html/login.html")
	}

	if r.URL.Path == "/profile" {
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

		userSession, exists := userSessions[sessionToken]
		if !exists {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		userInfo = userSession

		if r.Method == "POST" {
			action := r.FormValue("action")

			if action == "Modifier votre pseudo" {
				oldUsername := getUsername(db, userInfo.Username)
				fmt.Println(oldUsername)
				userInfo.Username = strings.TrimSpace(r.FormValue("username"))
				err := modifyUsername(db, userInfo.Username, oldUsername)
				if err != nil {
					fmt.Println(err)
				}
			}
			if action == "Modifier votre email" {
				oldEmail := getEmail(db, userInfo.Email)
				userInfo.Email = strings.TrimSpace(r.FormValue("email"))
				err := modifyEmail(db, userInfo.Email, oldEmail)
				if err != nil {
					fmt.Println(err)
				}
			}

			if action == "logout" {
				DeleteCookie(w, "session_token")
				http.Redirect(w, r, "/login", http.StatusFound)
			}
		}
		userInfo = verifyCookie(r, cookie, userInfo)
		t, _ = template.ParseFiles("src/html/profile.html")
	}
	t.Execute(w, userInfo)
}

func CreateCookie(w http.ResponseWriter, name string, value string) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   86400,
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

func verifyCookie(r *http.Request, cookies string, userInfo user) user {
	cookie, _ := getCookie(r, "session_token")
	if len(cookie) == 0 {
		userInfo.IsConnected = false
		return user{}
	} else {
		userInfo.IsConnected = true
		fmt.Println(userInfo.IsConnected)
		return userInfo
	}
}
