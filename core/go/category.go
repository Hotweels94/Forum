package forum

import (
	"database/sql"
	"fmt"
	"forum/core/structs"
	"html/template"
	"net/http"
	"strconv"
)

type Categories struct {
	Categories  []structs.Category
	User        structs.User
	IsConnected bool
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

func (ch *Categories) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	db, err := initDBCategory()
	if err != nil {
		http.Error(w, "Erreur de connexion à la base de données", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	if verifyCookie(r) {
		ch.IsConnected = true
		ch.User = userSession
	} else {
		ch.IsConnected = false
	}

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
		} else if ch.User.Role == "admin" || ch.User.Role == "moderator" || ch.User.Role == "user" {
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
		} else {
			http.Redirect(w, r, "/", http.StatusFound)
		}

	case "/list_category":
		if r.Method == "POST" {
			action := r.FormValue("action")
			switch action {
			case "delete":
				fmt.Println("delete category")
				if verifyCookie(r) {
					idcategory := r.FormValue("id")
					idcategoryint, _ := strconv.Atoi(idcategory)
					err := deleteCategory(db, idcategoryint)
					if err != nil {
						fmt.Println(err)
						http.Error(w, "Erreur lors de la suppression de la category", http.StatusInternalServerError)
						fmt.Println(err)
						return
					}
					http.Redirect(w, r, "/", http.StatusSeeOther)
					return
				}

			default:
				err := r.ParseForm()
				if err != nil {
					http.Error(w, "Erreur lors de l'analyse du formulaire", http.StatusInternalServerError)
					return
				}
			}
			http.Redirect(w, r, "/list_category", http.StatusSeeOther)
		}
		categories, err := getCategories(db)
		if err != nil {
			http.Error(w, "Erreur lors de la récupération des catégories", http.StatusInternalServerError)
			return
		}
		ch.Categories = categories
		t, _ := template.ParseFiles("src/html/list_category.html")
		t.Execute(w, ch)

	default:
		http.NotFound(w, r)
		return
	}
}

type list_Post struct {
	Posts        []structs.Post
	NameCategory string
	User         structs.User
}

func (p list_Post) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	db, err := initDBPost()
	if err != nil {
		http.Error(w, "Erreur de connexion à la base de données", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var t *template.Template

	switch r.URL.Path {
	case "/list_post":
		t, _ = template.ParseFiles("src/html/list_post.html")
		t.Execute(w, nil)
		return
	case "/user_posts":
		var posts list_Post

		if r.URL.Query().Get("username") != "" {
			username := r.URL.Query().Get("username")
			p.User.Username = username
			posts, err = GetListPostByUsername(db, username)
			p.Posts = posts.Posts
			if err != nil {
				http.Error(w, "Erreur lors de la récupération des posts de la liste de vos posts", http.StatusInternalServerError)
				return
			}
		}
		t, _ = template.ParseFiles("src/html/user_posts.html")
		t.Execute(w, p)
	default:
		http.NotFound(w, r)
	}
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
		var post structs.Post
		err := rows.Scan(&post.ID, &post.User, &post.Text, &post.Title, &post.ImageURL, &post.SelectedCategory)
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

func deleteCategory(db *sql.DB, id int) error {

	posts, err := GetListPostByCategoryID(db, id)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	for _, post := range posts.Posts {
		err = deletePostByID(db, post.ID)
		if err != nil {
			fmt.Println(err)
			return nil
		}
	}
	_, err = db.Exec("DELETE FROM category WHERE id = ?", id)
	if err != nil {
		return err
	}
	return nil
}
