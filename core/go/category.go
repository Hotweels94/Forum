package forum

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
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
		if r.URL.Query().Get("id") != "" {
			id := r.URL.Query().Get("id")
			idint, _ := strconv.Atoi(id)
			posts, err := GetListPostByCategoryID(db, idint)
			if err != nil {
				http.Error(w, "Erreur lors de la récupération des posts de la catégorie", http.StatusInternalServerError)
				return
			}
			t, _ := template.ParseFiles("src/html/list_post.html")
			t.Execute(w, posts)
		}
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

			t, _ := template.ParseFiles("src/html/category.html")
			t.Execute(w, nil)
		}

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

type list_Post struct {
	Posts        []Post
	NameCategory string
}

func (p list_Post) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var t *template.Template
	t, _ = template.ParseFiles("src/html/list_post.html")

	t.Execute(w, nil)
}

func GetListPostByCategoryID(db *sql.DB, categoryID int) (list_Post, error) {
	var listPost list_Post

	categoryQuery := `SELECT name FROM category WHERE id = ?`
	err := db.QueryRow(categoryQuery, categoryID).Scan(&listPost.NameCategory)
	if err != nil {
		if err == sql.ErrNoRows {
			return listPost, fmt.Errorf("no category found with id %d", categoryID)
		}
		return listPost, err
	}

	// Query to get the posts by category_id
	postsQuery := `SELECT id, user, text, title, imageURL, category_id FROM post WHERE category_id = ?`
	rows, err := db.Query(postsQuery, categoryID)
	if err != nil {
		return listPost, err
	}
	defer rows.Close()

	for rows.Next() {
		var post Post
		err := rows.Scan(&post.id, &post.User, &post.Text, &post.Title, &post.ImageURL, &post.SelectedCategory)
		if err != nil {
			return listPost, err
		}
		listPost.Posts = append(listPost.Posts, post)
	}

	if err = rows.Err(); err != nil {
		return listPost, err
	}

	return listPost, nil
}
