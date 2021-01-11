package databaseService

import (
	"github.com/EgorAist/TP_DB_project/internal/models"
	"github.com/jackc/pgx"
)

type Service interface {
	Clear() (err error)
	Status() (status models.Status, err error)
}

type service struct {
	db *pgx.ConnPool

}

func NewStorage(db *pgx.ConnPool) Service {
	return &service{
		db: db,
	}
}

func (s *service) Clear() (err error) {
	_, err = s.db.Exec("TRUNCATE users, forums, threads, posts, forum_users, votes CASCADE")
	if err != nil {
		return models.Error{Code: "500"}
	}
	return
}

func (s *service) Status() (status models.Status, err error) {
	err = s.db.QueryRow("SELECT (SELECT COUNT(*) FROM forums), (SELECT COUNT(*) FROM threads), (SELECT COUNT(*) FROM posts), (SELECT COUNT(*) FROM users)").
				Scan(&status.Forum, &status.Thread, &status.Post, &status.User)
	if err != nil && err != pgx.ErrNoRows {
		return status, models.Error{Code: "500"}
	}

	return
}