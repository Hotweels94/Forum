# Forum of Nicolas Decayeux and Ryan Amsellem--Bousignac 

## How to Launch the Forum

To launch the forum, you can clone or directly download this Repository. Then go to your terminal, in the folder where you cloned / downloaded the application, then execute the command `go run main.go`. Then go to your browser and type the URL: http://localhost:8080/

## Objectives

- Create a themed web forum.
- Implement categories, posts, comments, likes, dislikes, and post filtering.
- Use SQLite for database management.
- Enable user authentication and forum moderation.

## SQLite

Store data using SQLite.

## Pages

Include the following pages:

- **Main Pages:**
  - Landing page
  - Login/Registration page
  - Categories view
  - Posts view within a category
  - Post and comments view

- **Creation Pages:**
  - Create category
  - Create post

- **User Pages:**
  - Profile and edit page
  - Account activity page
  - Other users' profiles

Data must be fetched from the database, not hardcoded.

## Authentication

Users must be able to register and log in with email/username and password. Use encrypted passwords and session cookies with expiration dates.

## Categories, Posts, and Comments

- Only logged-in users can create categories, posts, and comments.
- Posts and comments are publicly visible.
- Registered users can like/dislike posts and comments.
- Support images (JPEG, PNG, GIF) up to 20MB.

## Filter

Implement filtering by:
- Categories
- Created posts (for logged-in users)
- Liked posts (for logged-in users)

## Moderation

- Moderators approve posts before public visibility.
- User roles: Guests, Users, Moderators, Administrators.
- Moderators can delete or report posts.
- Admins manage user roles and delete content.

## Security

- Encrypt passwords and optionally the database.
- Implement secure sessions and cookies.
