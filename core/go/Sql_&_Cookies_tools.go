package forum

import (
	"database/sql"
	"fmt"
	"forum/core/structs"
	"net/http"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

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

func verifyLog(db *sql.DB, username string, email string, password string) (structs.User, error) {
	var hashedPassword string
	var userData structs.User
	err := db.QueryRow("SELECT password, username, email FROM users WHERE username = ? OR email = ?", username, email).Scan(&hashedPassword, &userData.Username, &userData.Email)
	if err != nil {
		return userData, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return userData, err
}

func getUsername(db *sql.DB, username string) string {
	var userData structs.User
	err := db.QueryRow("SELECT username FROM users WHERE username = ?", username).Scan(&userData.Username)
	if err != nil {
		return ""
	}
	return userData.Username
}

func getEmail(db *sql.DB, email string) string {
	var userData structs.User
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

func verifyCookie(r *http.Request) bool {
	cookie, err := getCookie(r, "session_token")
	if len(cookie) == 0 || err != nil {
		return false
	} else {
		return true
	}
}
