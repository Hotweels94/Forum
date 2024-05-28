package forum

import (
	"database/sql"
	"errors"
	"fmt"
	"forum/core/structs"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gofrs/uuid"
)

type Posts struct {
	post structs.Post
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

	err = initDBComment(db)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func initDBComment(db *sql.DB) error {
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS comment (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            post_id TEXT NOT NULL,
            user TEXT NOT NULL,
            text TEXT NOT NULL,
            FOREIGN KEY (post_id) REFERENCES post(id)
        )
    `)
	return err
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

func GetPostByID(db *sql.DB, id string) (structs.Post, error) {
	var post structs.Post
	err := db.QueryRow("SELECT id, user, text, title, imageURL FROM post WHERE id = ?", id).Scan(&post.ID, &post.User, &post.Text, &post.Title, &post.ImageURL)
	if err != nil {
		return structs.Post{}, err
	}
	return post, nil
}

func insertComment(db *sql.DB, postID, user, text string) error {
	_, err := db.Exec("INSERT INTO comment (post_id, user, text) VALUES (?, ?, ?)", postID, user, text)
	if err != nil {
		return err
	}
	return nil
}

func getCommentsByPostID(db *sql.DB, postID string) ([]structs.Comment, error) {
	rows, err := db.Query("SELECT id, post_id, user, text FROM comment WHERE post_id = ?", postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []structs.Comment
	for rows.Next() {
		var comment structs.Comment
		if err := rows.Scan(&comment.ID, &comment.PostID, &comment.User, &comment.Text); err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}

const uploadPath = "/databases/upload_image"

func (p *Posts) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	db, err := initDBPost()
	err2 := initDBComment(db)
	if err != nil || err2 != nil {
		http.Error(w, "Erreur lors de la connexion à la base de données", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	switch r.URL.Path {
	case "/post":
		id := r.URL.Query().Get("id")
		if id != "" {
			post, err := GetPostByID(db, id)
			if err != nil {
				http.Error(w, "Erreur lors de la récupération du post", http.StatusInternalServerError)
				return
			}

			if r.Method == "POST" {
				text := r.FormValue("comment")
				user := "testUser"
				err := insertComment(db, id, user, text)
				if err != nil {
					http.Error(w, "Erreur lors de l'insertion du commentaire ", http.StatusInternalServerError)
					fmt.Println(err)
					return
				}

			}
			comments, err := getCommentsByPostID(db, id)
			if err != nil {
				http.Error(w, "Erreur lors de la récupération des commentaires", http.StatusInternalServerError)
				return
			}
			t, _ := template.ParseFiles("src/html/post.html")
			t.Execute(w, structs.PostWithComments{Post: post, Comments: comments})
			return
		}

		if r.Method == "POST" {
			err := r.ParseMultipartForm(20 << 20)
			if err != nil {
				http.Error(w, "Erreur lors de l'analyse du formulaire", http.StatusInternalServerError)
				return
			}
			p.post.Title = r.FormValue("title")
			p.post.Text = r.FormValue("content")
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

				p.post.ImageURL = uploadPath + "/" + fileID + ext
			}
			err = insertPost(db, p.post.User, p.post.Text, p.post.Title, p.post.ImageURL, categoryID)
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

		var convertedCategories []structs.Category
		for _, c := range categories {
			convertedCategories = append(convertedCategories, structs.Category(c))
		}

		t, _ := template.ParseFiles("src/html/create_post.html")
		pp := structs.PostPage{
			Post:       p.post,
			Categories: convertedCategories,
		}
		t.Execute(w, pp)
	default:
		http.NotFound(w, r)
	}
}

func generateUniqueFilename(uploadPath string, ext string) (string, error) {
	for i := 0; i < 100; i++ {
		id, err := uuid.NewV4()
		if err != nil {
			return "", err
		}

		filePath := filepath.Join(uploadPath, id.String()+ext)
		_, err = os.Stat(filePath)
		if os.IsNotExist(err) {
			return id.String(), nil
		}
	}
	return "", errors.New("failed to generate unique filename")
}

func getCategories(db *sql.DB) ([]structs.Category, error) {
	rows, err := db.Query("SELECT id, name, description FROM category")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []structs.Category
	for rows.Next() {
		var category structs.Category
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
