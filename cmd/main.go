package main

import (
	"fmt"
	"github.com/EgorAist/TP_DB_project/cmd/handlers"
	"github.com/EgorAist/TP_DB_project/internal/services"
	"github.com/EgorAist/TP_DB_project/internal/storages/databaseService"
	"github.com/EgorAist/TP_DB_project/internal/storages/forumStorage"
	"github.com/EgorAist/TP_DB_project/internal/storages/postStorage"
	"github.com/EgorAist/TP_DB_project/internal/storages/threadStorage"
	"github.com/EgorAist/TP_DB_project/internal/storages/userStorage"
	"github.com/EgorAist/TP_DB_project/internal/storages/voteStorage"
	"github.com/buaazp/fasthttprouter"
	"github.com/jackc/pgx"
	_ "github.com/swaggo/echo-swagger/example/docs"
	"github.com/valyala/fasthttp"
	"log"
)

func main() {
	connectionString := "postgres://forum_user:1221@localhost/tp_forum?sslmode=disable"
	config, err := pgx.ParseURI(connectionString)
	if err != nil {
		return
	}

	db, err := pgx.NewConnPool(
		pgx.ConnPoolConfig{
			ConnConfig:     config,
			MaxConnections: 2000,
		})

	if err != nil {
		fmt.Println(err)
		return
	}


	forums := forumStorage.NewStorage(db)
	threads := threadStorage.NewStorage(db)
	users := userStorage.NewStorage(db)
	votes := voteStorage.NewStorage(db)
	posts := postStorage.NewStorage(db)
	dbService := databaseService.NewStorage(db)

	service := services.NewService(forums, threads, users, posts, votes, dbService)

	handler := handlers.NewHandler(service, forums, users, threads, posts)
	rout := router(handler)

	fmt.Println("start server")
	err = fasthttp.ListenAndServe(":5000", redirect(rout, handler))
	if err != nil {
		log.Fatal(err)
	}
}

func redirect(router *fasthttprouter.Router, handler handlers.Handler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		path := string(ctx.Path())
		if path == "/api/forum/create" {
			handler.ForumCreate(ctx)
			return
		}
		router.Handler(ctx)
	}
}

func router(handler handlers.Handler) *fasthttprouter.Router {
	r := fasthttprouter.New()
	r.POST("/api/user/:nickname/create", handler.UserCreate)
	r.POST("/api/forum/:slug/create", handler.ThreadCreate)
	r.GET("/api/forum/:slug/details", handler.ForumGet)
	r.GET("/api/user/:nickname/profile", handler.UserGet)
	r.POST("/api/user/:nickname/profile", handler.UserUpdate)
	r.POST("/api/thread/:slug_or_id/vote", handler.ThreadVote)
	r.GET("/api/thread/:slug_or_id/details", handler.ThreadGet)
	r.POST("/api/thread/:slug_or_id/details", handler.ThreadUpdate)
	r.GET("/api/forum/:slug/threads", handler.ForumGetThreads)
	r.POST("/api/thread/:slug_or_id/create", handler.PostsCreate)
	r.POST("/api/service/clear", handler.Clear)
	r.GET("/api/service/status", handler.Status)
	r.POST("/api/post/:id/details", handler.PostUpdate)
	r.GET("/api/post/:id/details", handler.PostGet)
	r.GET("/api/thread/:slug_or_id/posts", handler.ThreadGetPosts)
	r.GET("/api/forum/:slug/users", handler.ForumGetUsers)
	return r
}
