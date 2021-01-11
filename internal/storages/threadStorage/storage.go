package threadStorage

import (
	"database/sql"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx"
	"github.com/EgorAist/TP_DB_project/internal/models"
)

type Storage interface {
	CreateThread(input models.Thread) (thread models.Thread, err error)
	GetDetails(input models.ThreadInput) (thread models.Thread, err error)
	UpdateThread(input models.ThreadUpdate) (thread models.Thread, err error)
	GetThreadsByForum(input models.ForumGetThreads) (threads []models.Thread, err error)
	CheckThreadIfExists(input models.ThreadInput) (thread models.ThreadInput, err error)
	GetThreadForPost(input models.ThreadInput, post *models.Thread) (err error)
	GetForumByThread(input *models.ThreadInput) (forum string, err error)
}

type storage struct {
	db *pgx.ConnPool

}

func NewStorage(db *pgx.ConnPool) Storage {
	return &storage{
		db: db,
	}
}

var (
	insertWithSlug = "INSERT INTO threads (author, created, forum, message, slug, title, votes) VALUES ((SELECT u.nickname FROM users u WHERE u.nickname = $1), $2, (SELECT f.slug FROM forums f WHERE f.slug = $3), $4, $5, $6, $7) RETURNING ID, author, created, forum, message, slug, title, votes"
	insertWithoutSlug = "INSERT INTO threads (author, created, forum, message, title, votes) VALUES ((SELECT u.nickname FROM users u WHERE u.nickname = $1), $2, (SELECT f.slug FROM forums f WHERE f.slug = $3), $4, $5, $6) RETURNING ID, author, created, forum, message, title, votes"

	selectBySlug = "SELECT author, created, forum, ID, message, slug, title, votes FROM threads WHERE slug = $1"
	selectByID = "SELECT author, created, forum, ID, message, slug, title, votes FROM threads WHERE ID = $1"

	selectThreads = "SELECT id, slug, author, created, forum, title, message, votes FROM threads WHERE forum = $1 ORDER BY created LIMIT $2"
	selectThreadsSince = "SELECT id, slug, author, created, forum, title, message, votes FROM threads WHERE forum = $1 AND created >= $2 ORDER BY created LIMIT $3"
	selectThreadsDesc = "SELECT id, slug, author, created, forum, title, message, votes FROM threads WHERE forum = $1 ORDER BY created DESC LIMIT $2"
	selectThreadsSinceDesc =  "SELECT id, slug, author, created, forum, title, message, votes FROM threads WHERE forum = $1 AND created <= $2 ORDER BY created DESC LIMIT $3"
)

func (s *storage) CreateThread(input models.Thread) (thread models.Thread, err error) {
	if input.Slug == "" {
		err = s.db.QueryRow(insertWithoutSlug, input.Author, input.Created, input.Forum, input.Message, input.Title, input.Votes).
					Scan(&thread.ID, &thread.Author, &thread.Created, &thread.Forum, &thread.Message, &thread.Title, &thread.Votes)
	} else {
		err = s.db.QueryRow(insertWithSlug, input.Author, input.Created, input.Forum, input.Message, input.Slug, input.Title, input.Votes).
					Scan(&thread.ID, &thread.Author, &thread.Created, &thread.Forum, &thread.Message, &thread.Slug, &thread.Title, &thread.Votes)
	}

	if pqErr, ok := err.(pgx.PgError); ok {
		switch pqErr.Code {
		case pgerrcode.UniqueViolation:
			return thread, models.Error{Code: "409"}
		case pgerrcode.NotNullViolation, pgerrcode.ForeignKeyViolation:
			return thread, models.Error{Code: "404"}
		default:
			return thread, models.Error{Code: "500"}
		}
	}

	return
}

func (s *storage) GetDetails(input models.ThreadInput) (thread models.Thread, err error) {
	slug := sql.NullString{}
	if input.Slug == "" {
		err = s.db.QueryRow(selectByID, input.ThreadID).
					Scan(&thread.Author, &thread.Created, &thread.Forum, &thread.ID, &thread.Message, &slug, &thread.Title, &thread.Votes)
	} else {
		err = s.db.QueryRow(selectBySlug, input.Slug).
			Scan(&thread.Author, &thread.Created, &thread.Forum, &thread.ID, &thread.Message, &slug, &thread.Title, &thread.Votes)
	}

	if err != nil {
		if err == pgx.ErrNoRows {
			return thread, models.Error{Code: "404"}

		}
		return thread, models.Error{Code: "500"}
	}

	if slug.Valid {
		thread.Slug = slug.String
	}

	return
}

func (s *storage) UpdateThread(input models.ThreadUpdate) (thread models.Thread, err error) {
	if input.Title != "" && input.Message != "" {
		err = s.db.QueryRow("UPDATE threads SET message = $1, title = $2 WHERE ID = $3 OR slug = $4 " +
								"RETURNING author, created, forum, ID, message, slug, title, votes",
							input.Message, input.Title, input.ThreadID, input.Slug).
					Scan(&thread.Author, &thread.Created, &thread.Forum, &thread.ID, &thread.Message, &thread.Slug, &thread.Title, &thread.Votes)

	} else if input.Title != "" && input.Message == "" {
		err = s.db.QueryRow("UPDATE threads SET title = $1 WHERE ID = $2 OR slug = $3 " +
								"RETURNING author, created, forum, ID, message, slug, title, votes",
								input.Title, input.ThreadID, input.Slug).
					Scan(&thread.Author, &thread.Created, &thread.Forum, &thread.ID, &thread.Message, &thread.Slug, &thread.Title, &thread.Votes)

	} else if input.Title == "" && input.Message != "" {
		err = s.db.QueryRow("UPDATE threads SET message = $1 WHERE ID = $2 OR slug = $3 " +
			"RETURNING author, created, forum, ID, message, slug, title, votes",
			input.Message, input.ThreadID, input.Slug).
			Scan(&thread.Author, &thread.Created, &thread.Forum, &thread.ID, &thread.Message, &thread.Slug, &thread.Title, &thread.Votes)


	} else if input.Title == "" && input.Message == "" {
		err = s.db.QueryRow("SELECT author, created, forum, ID, message, slug, title, votes FROM threads WHERE ID = $1 OR slug = $2", input.ThreadID, input.Slug).
					Scan(&thread.Author, &thread.Created, &thread.Forum, &thread.ID, &thread.Message, &thread.Slug, &thread.Title, &thread.Votes)
	}

	if err != nil {
		if err == pgx.ErrNoRows {
			return thread, models.Error{Code: "404"}

		}
		return thread, models.Error{Code: "500"}
	}

	return
}

func (s *storage) GetThreadsByForum(input models.ForumGetThreads) (threads []models.Thread, err error) {
	var rows *pgx.Rows
	if input.Since == "" && !input.Desc {
		rows, err = s.db.Query(selectThreads, input.Slug, input.Limit)
	} else if input.Since == "" && input.Desc {
		rows, err = s.db.Query(selectThreadsDesc,  input.Slug, input.Limit)
	}  else if input.Since != "" && !input.Desc {
		rows, err = s.db.Query(selectThreadsSince,  input.Slug, input.Since, input.Limit)
	} else if input.Since != "" && input.Desc {
		rows, err = s.db.Query(selectThreadsSinceDesc, input.Slug, input.Since, input.Limit)
	}

	if err != nil {
		return threads, models.Error{Code: "500"}
	}
	defer rows.Close()

	threads = make([]models.Thread, 0)
	for rows.Next() {
		thread := models.Thread{}
		slug := sql.NullString{}

		err = rows.Scan(&thread.ID, &slug, &thread.Author, &thread.Created, &thread.Forum, &thread.Title, &thread.Message, &thread.Votes)
		if err != nil {
			return threads, models.Error{Code: "500"}
		}

		if slug.Valid {
			thread.Slug = slug.String
		}

		threads = append(threads, thread)
	}

	return
}

func (s storage) CheckThreadIfExists(input models.ThreadInput) (thread models.ThreadInput, err error) {
	if input.Slug == "" {
		err = s.db.QueryRow("SELECT ID from threads WHERE ID = $1", input.ThreadID).Scan(&thread.ThreadID)
	} else {
		err = s.db.QueryRow("SELECT ID from threads WHERE slug = $1", input.Slug).Scan(&thread.ThreadID)
	}

	if err != nil {
		if err == pgx.ErrNoRows {
			return thread, models.Error{Code: "404"}
		}
		return thread, models.Error{Code: "500"}
	}

	return
}

func (s *storage) GetThreadForPost(input models.ThreadInput, thread *models.Thread) (err error) {
	slug := sql.NullString{}
	err = s.db.QueryRow(selectByID, input.ThreadID).
				Scan(&thread.Author, &thread.Created, &thread.Forum, &thread.ID, &thread.Message, &slug, &thread.Title, &thread.Votes)

	if err != nil {
		return models.Error{Code: "500"}
	}

	if slug.Valid {
		thread.Slug = slug.String
	}

	return
}

func (s *storage) GetForumByThread(input *models.ThreadInput) (forum string, err error) {
	if input.Slug == "" {
		err = s.db.QueryRow("SELECT forum FROM threads WHERE ID = $1", input.ThreadID).Scan(&forum)
	} else {
		err = s.db.QueryRow("SELECT forum, ID FROM threads WHERE slug = $1", input.Slug).Scan(&forum, &input.ThreadID)
	}
	if err != nil {
		if err == pgx.ErrNoRows {
			return forum, models.Error{Code: "404"}
		}

		return forum, models.Error{Code: "500"}
	}

	return
}
