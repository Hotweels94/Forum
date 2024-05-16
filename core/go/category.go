package forum

import (
	"database/sql"
	"html/template"
	"net/http"
)

type Category struct {
	ID          int
	Name        string
	Description string
}

func initDBCategory() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./databases/forum.db")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS category (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT NOT NULL,
			UNIQUE(name)
		)
	`)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func insertCategory(db *sql.DB, name string, description string) error {
	_, err := db.Exec("INSERT INTO category (name, description) VALUES(?, ?)", name, description)
	if err != nil {
		return err
	}

	return nil
}

func (ch Category) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	db, err := initDBCategory()
	if err != nil {
		http.Error(w, "Erreur de connexion à la base de données", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	switch r.URL.Path {
	case "/category":
		if r.Method == "POST" {
			err := r.ParseForm()
			if err != nil {
				http.Error(w, "Erreur lors de l'analyse du formulaire", http.StatusInternalServerError)
				return
			}
			name := r.FormValue("name")
			description := r.FormValue("description")

			err = insertCategory(db, name, description)
			if err != nil {
				http.Error(w, "Erreur lors de l'insertion de la catégorie dans la base de données", http.StatusInternalServerError)
				return
			}

			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		t, _ := template.ParseFiles("src/html/category.html")
		t.Execute(w, nil)
	case "/list_category":
		categories, err := getCategories(db)
		if err != nil {
			http.Error(w, "Erreur lors de la récupération des catégories", http.StatusInternalServerError)
			return
		}
		t, _ := template.ParseFiles("src/html/list_category.html")
		t.Execute(w, categories)
	default:
		http.NotFound(w, r)
		return
	}
}
