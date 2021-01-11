package voteStorage

import (
	"database/sql"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx"
	"github.com/EgorAist/TP_DB_project/internal/models"
)

type Storage interface {
	CreateVote(vote models.Vote, update bool) (thread models.Thread, err error)
	CheckDoubleVote(vote models.Vote) (thread models.Thread, err error)
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
	insertVote            = "INSERT INTO votes (user_nick, voice, thread) VALUES ($1, $2, $3) ON CONFLICT ON CONSTRAINT uniq_votes DO UPDATE SET voice = EXCLUDED.voice;"
	createThreadVotesUp   = "UPDATE threads SET votes = votes + 1 WHERE ID = $1 RETURNING ID, author, created, forum, message, slug, title, votes"
	createThreadVotesDown = "UPDATE threads SET votes = votes - 1 WHERE ID = $1 RETURNING ID, author, created, forum, message, slug, title, votes"

	updateThreadVotesUp   = "UPDATE threads SET votes = votes + 2 WHERE ID = $1 RETURNING ID, author, created, forum, message, slug, title, votes"
	updateThreadVotesDown = "UPDATE threads SET votes = votes - 2 WHERE ID = $1 RETURNING ID, author, created, forum, message, slug, title, votes"
)

func (s *storage) CreateVote(vote models.Vote, update bool) (thread models.Thread, err error) {
	boolVoice := getBoolVoice(vote)
	tx, err := s.db.Begin()
	if err != nil {
		return thread, models.Error{Code: "500"}
	}

	_, err = tx.Exec("SET LOCAL synchronous_commit TO OFF")
	if err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			return thread, models.Error{Code: "500"}
		}
		return thread, models.Error{Code: "500"}
	}

	_, err = tx.Exec(insertVote, vote.User, boolVoice, vote.Thread.ThreadID)
	if err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			return thread, models.Error{Code: "500"}
		}
		if pqErr, ok := err.(pgx.PgError); ok {
			switch pqErr.Code {
			case pgerrcode.ForeignKeyViolation:
				return thread, models.Error{Code: "404"}
			default:
				return thread, models.Error{Code: "500"}
			}
		}
		return thread, models.Error{Code: "500"}
	}

	slug := sql.NullString{}

	if update {
		if boolVoice {
			err = tx.QueryRow(updateThreadVotesUp, vote.Thread.ThreadID).
				Scan(&thread.ID, &thread.Author, &thread.Created, &thread.Forum, &thread.Message, &slug, &thread.Title, &thread.Votes)
		} else {
			err = tx.QueryRow(updateThreadVotesDown, vote.Thread.ThreadID).
				Scan(&thread.ID, &thread.Author, &thread.Created, &thread.Forum, &thread.Message, &slug, &thread.Title, &thread.Votes)
		}
	} else {
		if boolVoice {
			err = tx.QueryRow(createThreadVotesUp, vote.Thread.ThreadID).
				Scan(&thread.ID, &thread.Author, &thread.Created, &thread.Forum, &thread.Message, &slug, &thread.Title, &thread.Votes)
		} else {
			err = tx.QueryRow(createThreadVotesDown, vote.Thread.ThreadID).
				Scan(&thread.ID, &thread.Author, &thread.Created, &thread.Forum, &thread.Message, &slug, &thread.Title, &thread.Votes)
		}
	}

	if err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			return thread, models.Error{Code: "500"}
		}

		return thread, models.Error{Code: "500"}
	}

	if slug.Valid {
		thread.Slug = slug.String
	}

	if commitErr := tx.Commit(); commitErr != nil {
		return thread, models.Error{Code: "500"}
	}

	return
}

func getBoolVoice(vote models.Vote) bool {
	if vote.Voice == 1 {
		return true
	}
	return false
}

func (s *storage) CheckDoubleVote(vote models.Vote) (thread models.Thread, err error) {
	boolVoice := getBoolVoice(vote)
	var oldVoice bool
	err = s.db.QueryRow("SELECT voice FROM votes WHERE user_nick = $1 AND thread = $2", vote.User, vote.Thread.ThreadID).
				Scan(&oldVoice)
	if err != nil {
		if err == pgx.ErrNoRows {
			return thread, nil
		}
		return thread, models.Error{Code: "500"}
	}

	if oldVoice != boolVoice {
		return thread, models.Error{Code: "101"}
	}

	slug := sql.NullString{}
	err = s.db.QueryRow("SELECT ID, author, created, forum, message, slug, title, votes FROM threads WHERE ID = $1", vote.Thread.ThreadID).
				Scan(&thread.ID, &thread.Author, &thread.Created, &thread.Forum, &thread.Message, &slug, &thread.Title, &thread.Votes)

	if slug.Valid {
		thread.Slug = slug.String
	}

	if err != nil {
		return thread, models.Error{Code: "500"}
	}

	return thread, models.Error{Code: "409"}
}

