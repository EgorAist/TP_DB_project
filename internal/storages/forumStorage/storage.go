package forumStorage

import (
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx"
	"github.com/EgorAist/TP_DB_project/internal/models"
)

type Storage interface {
	CreateForum(forumSlug models.ForumCreate) (forum models.Forum, err error)
	GetDetails(forumSlug models.ForumInput) (forum models.Forum, err error)
	UpdateThreadsCount(input models.ForumInput) (err error)
	UpdatePostsCount(input models.ForumInput, posts int) (err error)
//	AddUserToForum(userID int, forumID int) (err error)
	GetForumSlug(slug string) (string, error)
	AddUserToForum(user string, forum string) (err error)
	CheckIfForumExists(input models.ForumInput) (err error)
	GetForumID(input models.ForumInput) (ID int, err error)
	GetForumForPost(forumSlug string, forum *models.Forum) (err error)
}

type storage struct {
	db *pgx.ConnPool
}

func NewStorage(db *pgx.ConnPool) Storage {
	return &storage{
		db: db,
	}
}

func (s *storage) CreateForum(forumSlug models.ForumCreate) (forum models.Forum, err error) {
	err = s.db.QueryRow("INSERT INTO forums (slug, title, user_nick) VALUES ($1, $2,(SELECT u.nickname FROM users u WHERE u.nickname = $3)) RETURNING slug, title, user_nick",
						forumSlug.Slug, forumSlug.Title, forumSlug.User).Scan(&forum.Slug, &forum.Title, &forum.User)

	if pqErr, ok := err.(pgx.PgError); ok {
		switch pqErr.Code {
		case pgerrcode.UniqueViolation:
			return forum, models.Error{Code: "409"}
		case pgerrcode.NotNullViolation, pgerrcode.ForeignKeyViolation:
			return forum, models.Error{Code: "404"}
		default:
			return forum, models.Error{Code: "500"}
		}
	}

	return forum, nil
}

func (s *storage) GetDetails(forumSlug models.ForumInput) (forum models.Forum, err error) {
	err = s.db.QueryRow("SELECT slug, title, threads, posts, user_nick FROM forums WHERE slug = $1", forumSlug.Slug).
				Scan(&forum.Slug, &forum.Title, &forum.Threads, &forum.Posts, &forum.User)

	if err != nil {
		if err == pgx.ErrNoRows {
			return forum, models.Error{Code: "404"}

		}
		return forum, models.Error{Code: "500"}
	}

	return forum, nil
}

func (s *storage) UpdateThreadsCount(input models.ForumInput) (err error) {
	_, err = s.db.Exec("UPDATE forums SET threads = threads + 1 WHERE slug = $1", input.Slug)
	if err != nil {
		return models.Error{Code: "500"}
	}
	return
}

func (s *storage) UpdatePostsCount(input models.ForumInput, posts int) (err error) {
	_, err = s.db.Exec("UPDATE forums SET posts = posts + $2 WHERE slug = $1", input.Slug, posts)
	if err != nil {
		return models.Error{Code: "500"}
	}
	return
}
/*
func (s *storage) AddUserToForum(userID int, forumID int) (err error) {
	_, err = s.db.Exec("INSERT INTO forum_users (forumID, userID) VALUES ($1, $2)", forumID, userID)
	if err != nil {
		if pqErr, ok := err.(pgx.PgError); ok {
			switch pqErr.Code {
			case pgerrcode.UniqueViolation:
				return models.Error{Code: "409"}
			}
		}
		return models.Error{Code: "500"}
	}

	return
}
*/

func (s *storage) AddUserToForum(user string, forum string) (err error) {
	_, err = s.db.Exec("INSERT INTO forum_users (forum, nickname) VALUES ($1, $2)", forum, user)
	if err != nil {
		if pqErr, ok := err.(pgx.PgError); ok {
			switch pqErr.Code {
			case pgerrcode.UniqueViolation:
				return models.Error{Code: "409"}
			}
		}
		return models.Error{Code: "500"}
	}

	return
}


func (s *storage) CheckIfForumExists(input models.ForumInput) (err error) {
	var ID int
	err = s.db.QueryRow("SELECT ID from forums WHERE slug = $1", input.Slug).Scan(&ID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return models.Error{Code: "404"}
		}
		return models.Error{Code: "500"}
	}

	return
}

func (s storage) GetForumID(input models.ForumInput) (ID int, err error) {
	err = s.db.QueryRow("SELECT ID from forums WHERE slug = $1", input.Slug).Scan(&ID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return ID, models.Error{Code: "404"}
		}
		return ID, models.Error{Code: "500"}
	}

	return
}
func (s storage) GetForumSlug(slug string) (string, error) {
	var rightSlug string
	err := s.db.QueryRow("SELECT slug from forums WHERE slug = $1", slug).Scan(&rightSlug)
	if err != nil {
		if err == pgx.ErrNoRows {
			return rightSlug, models.Error{Code: "404"}
		}
		return rightSlug, models.Error{Code: "500"}
	}

	return rightSlug, nil
}

func (s *storage) GetForumForPost(forumSlug string, forum *models.Forum) (err error) {
	forum.Slug = forumSlug
	err = s.db.QueryRow("SELECT title, threads, posts, user_nick FROM forums WHERE slug = $1", forumSlug).
		Scan(&forum.Title, &forum.Threads, &forum.Posts, &forum.User)

	if err != nil {
		return models.Error{Code: "500"}
	}

	return
}