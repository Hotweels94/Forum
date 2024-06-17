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
	"time"

	"github.com/gofrs/uuid"
)

type Posts struct {
	post        structs.Post
	User        structs.User
	IsConnected bool
}

// inirt the db post comment and like
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
            reported INTEGER DEFAULT 0,
			date TEXT NOT NULL,
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
	err = initDBlike(db)
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
			date TEXT NOT NULL,
            FOREIGN KEY (post_id) REFERENCES post(id)
        )
    `)
	return err
}

// insert in the db a post
func insertPost(db *sql.DB, user string, text string, title string, imageURL string, categoryID int, date string) error {
	id, err := uuid.NewV4()
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO post (id, user, text, title, imageURL, category_id, date) VALUES(?, ?, ?, ?, ?, ?,?)", id.String(), user, text, title, imageURL, categoryID, date)
	if err != nil {
		return err
	}

	return nil
}

// get all informations of a post by id
func GetPostByID(db *sql.DB, id string) (structs.Post, error) {
	var post structs.Post
	err := db.QueryRow("SELECT id, user, text, title, imageURL FROM post WHERE id = ?", id).Scan(&post.ID, &post.User, &post.Text, &post.Title, &post.ImageURL)
	if err != nil {
		return structs.Post{}, err
	}
	return post, nil
}

// add a comment to a post
func insertComment(db *sql.DB, postID, user, text string, date string) error {
	_, err := db.Exec("INSERT INTO comment (post_id, user, text, date) VALUES (?, ?, ?, ?)", postID, user, text, date)
	if err != nil {
		return err
	}
	return nil
}

// get all comments of a post
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

// ServerHTTP for the post
func (p *Posts) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	db, err := initDBPost()
	err2 := initDBComment(db)
	if err != nil || err2 != nil {
		http.Error(w, "Erreur lors de la connexion à la base de données", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var pageData structs.PostWithComments

	if verifyCookie(r) {
		pageData.IsConnected = true
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
		pageData.User = userSession
	} else {
		pageData.IsConnected = false
	}

	switch r.URL.Path {
	case "/post":
		if r.Method == "POST" {
			// display the post and the comments
			id := r.FormValue("id")
			action := r.FormValue("action")
			fmt.Println(action)
			switch action {
			case "comment":
				// button for add a comment
				text := r.FormValue("comment")
				if verifyCookie(r) {
					p.IsConnected = true
					p.User = userSession
				} else {
					p.IsConnected = false
				}
				currentTime := time.Now()
				datecomment := currentTime.Format("2006-01-02 15:04:05")
				err := insertComment(db, id, p.User.Username, text, datecomment)
				if err != nil {
					http.Error(w, "Erreur lors de l'insertion du commentaire ", http.StatusInternalServerError)
					fmt.Println(err)
					return
				}
			case "delete":
				// button for delete a post
				if verifyCookie(r) {
					err := deletePostByID(db, id)
					if err != nil {
						http.Error(w, "Erreur lors de la suppression du post", http.StatusInternalServerError)
						fmt.Println(err)
						return
					}
					http.Redirect(w, r, "/", http.StatusSeeOther)
					return
				}
			case "deletecomment":
				// button for delete a comment
				if verifyCookie(r) {
					idcomment := r.FormValue("id")
					idint, _ := strconv.Atoi(idcomment)
					err := deleteCommentByID(db, idint)
					if err != nil {
						http.Error(w, "Erreur lors de la suppression du commentaire", http.StatusInternalServerError)
						fmt.Println(err)
						return
					}
					http.Redirect(w, r, "/"+id, http.StatusSeeOther)
					return
				}
			case "report":
				// button for report a post
				if verifyCookie(r) {
					err := reportPostByID(db, id)
					if err != nil {
						http.Error(w, "Erreur lors du signalement du post", http.StatusInternalServerError)
						fmt.Println(err)
						return
					}
					http.Redirect(w, r, "/report", http.StatusSeeOther)
					return
				}
			case "dislike":
				// button for dislike a post
				if verifyCookie(r) {
					p.IsConnected = true
					p.User = userSession
				} else {
					p.IsConnected = false
				}
				err := insertLike(db, id, p.User.Username, false)
				if err != nil {
					http.Error(w, "Erreur lors du dislike du post", http.StatusInternalServerError)
					fmt.Println(err)
					return
				}
				http.Redirect(w, r, "/post?id="+id, http.StatusSeeOther)
				return
			case "like":
				// button for like a post
				if verifyCookie(r) {
					p.IsConnected = true
					p.User = userSession
				} else {
					p.IsConnected = false
				}
				err := insertLike(db, id, p.User.Username, true)
				if err != nil {
					http.Error(w, "Erreur lors du like du post", http.StatusInternalServerError)
					fmt.Println(err)
					return
				}
				http.Redirect(w, r, "/post?id="+id, http.StatusSeeOther)
				return
			}
		}
		id := r.URL.Query().Get("id")
		if id != "" {
			// get all informations of a post
			post, err := GetPostByID(db, id)
			if err != nil {
				http.Error(w, "Erreur lors de la récupération du post", http.StatusInternalServerError)
				return
			}

			comments, err := getCommentsByPostID(db, id)
			if err != nil {
				http.Error(w, "Erreur lors de la récupération des commentaires", http.StatusInternalServerError)
				return
			}
			likes, err := countLike(db, id)
			if err != nil {
				http.Error(w, "Erreur lors de la récupération des likes", http.StatusInternalServerError)
				return
			}
			deslikes, err := countDislike(db, id)
			if err != nil {
				http.Error(w, "Erreur lors de la récupération des deslikes", http.StatusInternalServerError)
				return
			}

			pageData.Post = post
			pageData.Comments = comments
			pageData.Likes = likes
			pageData.Dislikes = deslikes

			t, _ := template.ParseFiles("src/html/post.html")
			t.Execute(w, pageData)
			return
		}

		if r.Method == "POST" {
			// insert a post in the db
			if pageData.User.Role == "admin" || pageData.User.Role == "moderator" || pageData.User.Role == "user" { //double check useless
				if verifyCookie(r) {
					p.IsConnected = true
					p.User = userSession
				} else {
					p.IsConnected = false
				}
				err := r.ParseForm()
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
				// Check if the category exists
				file, fileHeader, err := r.FormFile("image")
				if err != nil && err != http.ErrMissingFile {
					http.Error(w, "Erreur lors de la récupération du fichier", http.StatusInternalServerError)
					return
				}
				if err == nil {
					defer file.Close()

					ext := filepath.Ext(fileHeader.Filename)
					// Check if the file extension is allowed
					allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true}
					if !allowedExts[ext] {
						http.Error(w, "Extension de fichier non autorisée", http.StatusBadRequest)
						return
					}
					// Check if the file size is allowed
					fileSize := fileHeader.Size
					var maxFileSize int64 = 20 * 1024 * 1024
					if fileSize > maxFileSize {
						http.Error(w, "Image trop grande (max 20 Mo)", http.StatusBadRequest)
						return
					}
					// Generate a unique filename by appending an UUID to the upload path
					fileID, err := generateUniqueFilename(uploadPath, ext)
					if err != nil {
						http.Error(w, "Erreur lors de la génération de l'ID de fichier unique", http.StatusInternalServerError)
						return
					}
					// Create the file on the server
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
				// get the current date
				currentTime := time.Now()
				date := currentTime.Format("2006-01-02 15:04:05")
				err = insertPost(db, p.User.Username, p.post.Text, p.post.Title, p.post.ImageURL, categoryID, date)
				p.post.ImageURL = ""
				if err != nil {
					http.Error(w, "Erreur lors de l'insertion du post dans la base de données", http.StatusInternalServerError)
					return
				}
			}
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		if pageData.User.Role == "admin" || pageData.User.Role == "moderator" || pageData.User.Role == "user" {
			// get all categories to create a post in good category
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
				Post:        p.post,
				Categories:  convertedCategories,
				User:        p.User,
				IsConnected: p.IsConnected,
			}
			if verifyCookie(r) {
				pp.IsConnected = true
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

				userSession, exists = userSessions[sessionToken]
				if !exists {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				pp.User = userSession
			} else {
				pp.IsConnected = false
			}
			t.Execute(w, pp)
		} else {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
	case "/report":
		// see all reported post
		reportedPosts, err := getReportedPosts(db)
		if err != nil {
			http.Error(w, "Erreur lors de la récupération des posts signalés", http.StatusInternalServerError)
			return
		}
		t, _ := template.ParseFiles("src/html/report.html")
		t.Execute(w, ReportedPosts{Posts: reportedPosts})
		return
	default:
		http.NotFound(w, r)
	}
}

// generate a unique filename by using an UUID
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

// get all categories from the db
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

// delete a post by id and delete all comments and likes of this post
func deletePostByID(db *sql.DB, id string) error {
	_, err := db.Exec("DELETE FROM post WHERE id = ?", id)
	deleteCommentByPostID(db, id)
	deleteLike(db, id)
	return err
}

// delete a comment by his id
func deleteCommentByID(db *sql.DB, id int) error {
	_, err := db.Exec("DELETE FROM comment WHERE id = ?", id)
	return err
}

// delete all comment by post id
func deleteCommentByPostID(db *sql.DB, id string) error {
	_, err := db.Exec("DELETE FROM comment WHERE post_id = ?", id)
	return err
}
