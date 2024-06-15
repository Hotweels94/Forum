package forum

import (
	"database/sql"
	"forum/core/structs"
)

type ReportedPosts struct {
	Posts []structs.Post
}

// update to set a post reported
func reportPostByID(db *sql.DB, id string) error {
	_, err := db.Exec("UPDATE post SET reported = 1 WHERE id = ?", id)
	return err
}

// get all reported posts
func getReportedPosts(db *sql.DB) ([]structs.Post, error) {
	rows, err := db.Query("SELECT id, user, text, title, imageURL, category_id FROM post WHERE reported = 1")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []structs.Post
	for rows.Next() {
		var post structs.Post
		err := rows.Scan(&post.ID, &post.User, &post.Text, &post.Title, &post.ImageURL, &post.SelectedCategory)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	return posts, nil
}
