package postStorage

import (
	"fmt"
	"github.com/EgorAist/TP_DB_project/internal/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx"
	"strings"
)

type Storage interface {
	CreatePosts(thread models.ThreadInput, forum string, created string, posts []models.PostCreate) (post []models.Post, err error)
	CreatePost(input models.Post) (post models.Post, err error)
	GetPostDetails(input models.PostInput, post *models.Post) (err error)
	UpdatePost(input models.PostUpdate) (post models.Post, err error)
	GetPostsByThread(input models.ThreadGetPosts) (posts []models.Post, err error)
	CheckParentPostThread(post int) (thread int, err error)
}

type storage struct {
	db *pgx.ConnPool
}

func NewStorage(db *pgx.ConnPool) Storage {
	return &storage{
		db: db,
	}
}

func (s storage) CreatePosts(thread models.ThreadInput, forum string, created string, posts []models.PostCreate) (post []models.Post, err error) {
	query := `INSERT INTO posts(
                 author,
                 created,
                 message,
                 parent,
				 thread,
				 forum) VALUES `
	data := make([]models.Post, 0, 0)
	if len(posts) == 0 {
		return data, nil
	}


	var valNam []string
	var values []interface{}
	i := 1
	for _, element := range posts {
		valNam = append(valNam, fmt.Sprintf(
			"($%d, $%d, $%d, $%d, $%d, $%d)",
			i, i+1, i+2, i+3, i+4, i+5))
		i += 6
		values = append(values, element.Author, created, element.Message, element.Parent, thread.ThreadID, forum)
	}

	query += strings.Join(valNam[:], ",")
	query += " RETURNING  id, parent, thread, forum, author, created, message, edited"
	row, err := s.db.Query(query, values...)
	if err != nil {
		fmt.Println(err)
		return data, err
	}

	defer func() {
		if row != nil {
			row.Close()
		}
	}()

	for row.Next() {
		scanPost := models.Post{}
		// id, parent, thread, forum, author, created, message, edited
		err = row.Scan(&scanPost.ID, &scanPost.Parent, &scanPost.ThreadID, &scanPost.Forum,  &scanPost.Author, &scanPost.Created,&scanPost.Message, &scanPost.IsEdited)

		if err != nil {
			fmt.Println(err)
			return data, err
		}
		data = append(data, scanPost)
	}

	if len(data) == 0 {
		return data, models.Error{Code: "409"}
	}
	return data, err
}

func (s *storage) CreatePost(input models.Post) (post models.Post, err error) {
	if input.Parent == 0 {
		err = s.db.QueryRow("INSERT INTO posts (author, created, forum, message, parent, thread, path) VALUES ($1,$2,$3,$4,$5,$6, array[(select currval('post_id_seq')::integer)]) RETURNING ID",
			input.Author, input.Created, input.Forum, input.Message, input.Parent, input.ThreadInput.ThreadID).Scan(&post.ID)
	} else {
		err = s.db.QueryRow("INSERT INTO posts (author, created, forum, message, parent, thread, path) VALUES ($1,$2,$3,$4,$5,$6, (SELECT path FROM posts WHERE id = $5) || (select currval('post_id_seq')::integer)) RETURNING ID",
			input.Author, input.Created, input.Forum, input.Message, input.Parent, input.ThreadInput.ThreadID).Scan(&post.ID)
	}

	if pqErr, ok := err.(pgx.PgError); ok {
		switch pqErr.Code {
		case pgerrcode.UniqueViolation:
			return post, models.Error{Code: "409", Message: "conflict post"}
		case pgerrcode.NotNullViolation, pgerrcode.ForeignKeyViolation:
			return post, models.Error{Code: "404", Message: "conflict post"}
		default:
			return post, models.Error{Code: "500", Message: "conflict post"}
		}
	}

	post.Author = input.Author
	post.Created = input.Created
	post.Forum = input.Forum
	post.Message = input.Message
	post.Parent = input.Parent
	post.ThreadInput.ThreadID = input.ThreadInput.ThreadID
	post.IsEdited = false

	return
}

func (s *storage) GetPostDetails(input models.PostInput, post *models.Post) (err error) {
	err = s.db.QueryRow("SELECT author, created, forum, message, ID , edited, parent, thread FROM posts WHERE ID = $1", input.ID).
				Scan(&post.Author, &post.Created, &post.Forum, &post.Message, &post.ID, &post.IsEdited, &post.Parent, &post.ThreadInput.ThreadID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return models.Error{Code: "404"}

		}
		return models.Error{Code: "500"}
	}
	return
}

func (s *storage) UpdatePost(input models.PostUpdate) (post models.Post, err error) {
	var oldMessage string
	err = s.db.QueryRow("SELECT message FROM posts WHERE ID = $1", input.ID).
		Scan(&oldMessage)
	if err != nil {
		if err == pgx.ErrNoRows {
			return post, models.Error{Code: "404"}
		}
		return post, models.Error{Code: "500"}
	}

	if input.Message != "" && input.Message != oldMessage {
		err = s.db.QueryRow("UPDATE posts SET message = $1, edited = $2 WHERE ID = $3 RETURNING author, created, forum, message, ID , edited, parent, thread", input.Message, true, input.ID).
			Scan(&post.Author, &post.Created, &post.Forum, &post.Message, &post.ID, &post.IsEdited, &post.Parent, &post.ThreadInput.ThreadID)
	} else {
		err = s.db.QueryRow("SELECT author, created, forum, message, ID , edited, parent, thread FROM posts WHERE ID = $1", input.ID).
			Scan(&post.Author, &post.Created, &post.Forum, &post.Message, &post.ID, &post.IsEdited, &post.Parent, &post.ThreadInput.ThreadID)
		}

	if err != nil {
		return post, models.Error{Code: "500"}
	}
	return
}

const selectPostsFlatLimitByID = `
	SELECT p.id, p.author, p.created, p.edited, p.message, p.parent, p.thread, p.forum
	FROM posts p
	WHERE p.thread = $1
	ORDER BY p.created, p.id
	LIMIT $2
`

const selectPostsFlatLimitDescByID = `
	SELECT p.id, p.author, p.created, p.edited, p.message, p.parent, p.thread, p.forum
	FROM posts p
	WHERE p.thread = $1
	ORDER BY p.created DESC, p.id DESC
	LIMIT $2
`
const selectPostsFlatLimitSinceByID = `
	SELECT p.id, p.author, p.created, p.edited, p.message, p.parent, p.thread, p.forum
	FROM posts p
	WHERE p.thread = $1 and p.id > $2
	ORDER BY p.created, p.id
	LIMIT $3
`
const selectPostsFlatLimitSinceDescByID = `
	SELECT p.id, p.author, p.created, p.edited, p.message, p.parent, p.thread, p.forum
	FROM posts p
	WHERE p.thread = $1 and p.id < $2
	ORDER BY p.created DESC, p.id DESC
	LIMIT $3
`
const selectPostsTreeLimitByID = `
	SELECT p.id, p.author, p.created, p.edited, p.message, p.parent, p.thread, p.forum
	FROM posts p
	WHERE p.thread = $1
	ORDER BY p.path
	LIMIT $2
`
const selectPostsTreeLimitDescByID = `
	SELECT p.id, p.author, p.created, p.edited, p.message, p.parent, p.thread, p.forum
	FROM posts p
	WHERE p.thread = $1
	ORDER BY path DESC
	LIMIT $2
`
const selectPostsTreeLimitSinceByID = `
	SELECT p.id, p.author, p.created, p.edited, p.message, p.parent, p.thread, p.forum
	FROM posts p
	WHERE p.thread = $1 and (p.path > (SELECT p2.path from posts p2 where p2.id = $2))
	ORDER BY p.path
	LIMIT $3
`
const selectPostsTreeLimitSinceDescByID = `
	SELECT p.id, p.author, p.created, p.edited, p.message, p.parent, p.thread, p.forum
	FROM posts p
	WHERE p.thread = $1 and (p.path < (SELECT p2.path from posts p2 where p2.id = $2))
	ORDER BY p.path DESC
	LIMIT $3
`
const selectPostsParentTreeLimitByID = `
	SELECT p.id, p.author, p.created, p.edited, p.message, p.parent, p.thread, p.forum
	FROM posts p
	WHERE p.thread = $1 and p.path[1] IN (
		SELECT p2.path[1]
		FROM posts p2
		WHERE p2.thread = $2 AND p2.parent = 0
		ORDER BY p2.path
		LIMIT $3
	)
	ORDER BY path
`
const selectPostsParentTreeLimitDescByID = `
	SELECT p.id, p.author, p.created, p.edited, p.message, p.parent, p.thread, p.forum
	FROM posts p
	WHERE p.thread = $1 and p.path[1] IN (
		SELECT p2.path[1]
		FROM posts p2
		WHERE p2.parent = 0 and p2.thread = $2
		ORDER BY p2.path DESC
		LIMIT $3
	)
	ORDER BY p.path[1] DESC, p.path[2:]
`

const selectPostsParentTreeLimitSinceByID = `
	SELECT p.id, p.author, p.created, p.edited, p.message, p.parent, p.thread, p.forum
	FROM posts p
	WHERE p.thread = $1 and p.path[1] IN (
		SELECT p2.path[1]
		FROM posts p2
		WHERE p2.thread = $2 AND p2.parent = 0 and p2.path[1] > (SELECT p3.path[1] from posts p3 where p3.id = $3)
		ORDER BY p2.path
		LIMIT $4
	)
	ORDER BY p.path
`
const selectPostsParentTreeLimitSinceDescByID = `
	SELECT p.id, p.author, p.created, p.edited, p.message, p.parent, p.thread, p.forum
	FROM posts p
	WHERE p.thread = $1 and p.path[1] IN (
		SELECT p2.path[1]
		FROM posts p2
		WHERE p2.thread = $2 AND p2.parent = 0 and p2.path[1] < (SELECT p3.path[1] from posts p3 where p3.id = $3)
		ORDER BY p2.path DESC
		LIMIT $4
	)
	ORDER BY p.path[1] DESC, p.path[2:]
`

func (s *storage) GetPostsByThread(input models.ThreadGetPosts) (posts []models.Post, err error){
	var rows *pgx.Rows
	posts  = make([]models.Post, 0)
	switch input.Sort {
	case "flat":
		if input.Since > 0 {
			if input.Desc {
				rows, err = s.db.Query(selectPostsFlatLimitSinceDescByID, input.ThreadInput.ThreadID,
					input.Since, input.Limit)
			} else {
				rows, err = s.db.Query(selectPostsFlatLimitSinceByID, input.ThreadInput.ThreadID,
					input.Since, input.Limit)
			}
		} else {
			if input.Desc == true {
				rows, err = s.db.Query(selectPostsFlatLimitDescByID, input.ThreadInput.ThreadID, input.Limit)
			} else {
				rows, err = s.db.Query(selectPostsFlatLimitByID, input.ThreadInput.ThreadID, input.Limit)
			}
		}
	case "tree":
		if input.Since > 0 {
			if input.Desc {
				rows, err = s.db.Query(selectPostsTreeLimitSinceDescByID, input.ThreadInput.ThreadID,
					input.Since, input.Limit)
			} else {
				rows, err = s.db.Query(selectPostsTreeLimitSinceByID, input.ThreadInput.ThreadID,
					input.Since, input.Limit)
			}
		} else {
			if input.Desc {
				rows, err = s.db.Query(selectPostsTreeLimitDescByID, input.ThreadInput.ThreadID, input.Limit)
			} else {
				rows, err = s.db.Query(selectPostsTreeLimitByID, input.ThreadInput.ThreadID, input.Limit)
			}
		}
	case "parent_tree":
		if input.Since > 0 {
			if input.Desc {
				rows, err = s.db.Query(selectPostsParentTreeLimitSinceDescByID, input.ThreadInput.ThreadID, input.ThreadInput.ThreadID,
					input.Since, input.Limit)
			} else {
				rows, err = s.db.Query(selectPostsParentTreeLimitSinceByID, input.ThreadInput.ThreadID, input.ThreadInput.ThreadID,
					input.Since, input.Limit)
			}
		} else {
			if input.Desc {
				rows, err = s.db.Query(selectPostsParentTreeLimitDescByID, input.ThreadInput.ThreadID, input.ThreadInput.ThreadID,
					input.Limit)
			} else {
				rows, err = s.db.Query(selectPostsParentTreeLimitByID, input.ThreadInput.ThreadID, input.ThreadInput.ThreadID,
					input.Limit)
			}
		}
	default:
		if input.Since > 0 {
			if input.Desc {
				rows, err = s.db.Query(selectPostsFlatLimitSinceDescByID, input.ThreadInput.ThreadID,
					input.Since, input.Limit)
			} else {
				rows, err = s.db.Query(selectPostsFlatLimitSinceByID, input.ThreadInput.ThreadID,
					input.Since, input.Limit)
			}
		} else {
			if input.Desc == true {
				rows, err = s.db.Query(selectPostsFlatLimitDescByID, input.ThreadInput.ThreadID, input.Limit)
			} else {
				rows, err = s.db.Query(selectPostsFlatLimitByID, input.ThreadInput.ThreadID, input.Limit)
			}
		}
	}

	if err != nil {
		return posts, models.Error{Code: "500"}
	}
	defer rows.Close()

	if rows == nil {
		return posts, models.Error{Code: "500"}
	}

	for rows.Next() {
		post := models.Post{}

		err = rows.Scan(&post.ID, &post.Author, &post.Created, &post.IsEdited, &post.Message, &post.Parent, &post.ThreadInput.ThreadID, &post.Forum)
		if err != nil {
			return posts, models.Error{Code: "500"}
		}

		posts = append(posts, post)
	}

	return 
}

func (s storage) CheckParentPostThread(post int) (thread int, err error) {
	err = s.db.QueryRow("SELECT thread FROM posts WHERE ID = $1", post).Scan(&thread)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, models.Error{Code: "409"}
		}
		return 0, models.Error{Code: "500"}
	}

	return
}
