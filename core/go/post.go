package forum

import (
	"database/sql"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gofrs/uuid"
)

type Post struct {
	User             string
	Text             string
	Title            string
	ImageURL         string
	SelectedCategory int
	Categories       []Category
}

type PostPage struct {
	Post       Post
	Categories []Category
}

func initDBPost() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./databases/forum.db")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS post (
            id TEXT PRIMARY KEY,
            user TEXT NOT NULL,
            text TEXT NOT NULL,
            title TEXT NOT NULL,
            imageURL TEXT,
            category_id INTEGER NOT NULL,
            UNIQUE(id)
        )
    `)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func insertPost(db *sql.DB, user string, text string, title string, imageURL string, categoryID int) error {
	id, err := uuid.NewV4()
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO post (id, user, text, title, imageURL, category_id) VALUES(?, ?, ?, ?, ?, ?)", id.String(), user, text, title, imageURL, categoryID)
	if err != nil {
		return err
	}

	return nil
}

func GetPostByID(db *sql.DB, id string) (Post, error) {
	var post Post
	err := db.QueryRow("SELECT user, text, title, imageURL FROM post WHERE id = ?", id).Scan(&post.User, &post.Text, &post.Title, &post.ImageURL)
	if err != nil {
		return Post{}, err
	}
	return post, nil
}

const uploadPath = "/databases/upload_image"

func (p Post) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	db, err := initDBPost()
	if err != nil {
		http.Error(w, "Erreur lors de la connexion à la base de données", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var t *template.Template
	switch r.URL.Path {
	case "/post":
		id := r.URL.Query().Get("id")
		if id != "" {
			post, err := GetPostByID(db, id)
			if err != nil {
				http.Error(w, "Erreur lors de la récupération du post", http.StatusInternalServerError)
				return
			}
			t, _ = template.ParseFiles("src/html/post.html")
			t.Execute(w, post)
			return
		}

		if r.Method == "POST" {
			err := r.ParseMultipartForm(20 << 20)
			if err != nil {
				http.Error(w, "Erreur lors de l'analyse du formulaire", http.StatusInternalServerError)
				return
			}
			p.Title = r.FormValue("title")
			p.Text = r.FormValue("content")
			categoryIDStr := r.FormValue("category")
			categoryID, err := strconv.Atoi(categoryIDStr)
			if err != nil {
				http.Error(w, "ID de catégorie invalide", http.StatusBadRequest)
				return
			}

			file, fileHeader, err := r.FormFile("image")
			if err != nil && err != http.ErrMissingFile {
				http.Error(w, "Erreur lors de la récupération du fichier", http.StatusInternalServerError)
				return
			}
			if err == nil {
				defer file.Close()

				ext := filepath.Ext(fileHeader.Filename)
				allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true}
				if !allowedExts[ext] {
					http.Error(w, "Extension de fichier non autorisée", http.StatusBadRequest)
					return
				}

				fileSize := fileHeader.Size
				var maxFileSize int64 = 20 * 1024 * 1024
				if fileSize > maxFileSize {
					http.Error(w, "Image trop grande (max 20 Mo)", http.StatusBadRequest)
					return
				}

				fileID, err := generateUniqueFilename(uploadPath, ext)
				if err != nil {
					http.Error(w, "Erreur lors de la génération de l'ID de fichier unique", http.StatusInternalServerError)
					return
				}

				filePath := filepath.Join("databases/upload_image", fileID+ext)
				outFile, err := os.Create(filePath)
				if err != nil {
					http.Error(w, "Erreur lors de la création du fichier ", http.StatusInternalServerError)
					return
				}
				defer outFile.Close()

				_, err = io.Copy(outFile, file)
				if err != nil {
					http.Error(w, "Erreur lors de la copie des données du fichier", http.StatusInternalServerError)
					return
				}

				p.ImageURL = uploadPath + "/" + fileID + ext
			}
			err = insertPost(db, p.User, p.Text, p.Title, p.ImageURL, categoryID)
			if err != nil {
				http.Error(w, "Erreur lors de l'insertion du post dans la base de données", http.StatusInternalServerError)
				return
			}

			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		categories, err := getCategories(db)
		if err != nil {
			http.Error(w, "Erreur lors de la récupération des catégories", http.StatusInternalServerError)
			return
		}

		t, _ = template.ParseFiles("src/html/create_post.html")
		pp := PostPage{
			Post:       p,
			Categories: categories,
		}
		t.Execute(w, pp)
	default:
		http.NotFound(w, r)
		return
	}
}

func generateUniqueFilename(uploadPath string, ext string) (string, error) {
	for {
		fileID := uuid.Must(uuid.NewV4()).String()
		filePath := filepath.Join(uploadPath, fileID+ext)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return fileID, nil
		}
	}
}

func getCategories(db *sql.DB) ([]Category, error) {
	rows, err := db.Query("SELECT id, name, description FROM category")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var category Category
		if err := rows.Scan(&category.ID, &category.Name, &category.Description); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}
