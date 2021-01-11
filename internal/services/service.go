package services

import (
	"fmt"
	"github.com/EgorAist/TP_DB_project/internal/models"
	"github.com/EgorAist/TP_DB_project/internal/storages/databaseService"
	"github.com/EgorAist/TP_DB_project/internal/storages/forumStorage"
	"github.com/EgorAist/TP_DB_project/internal/storages/postStorage"
	"github.com/EgorAist/TP_DB_project/internal/storages/threadStorage"
	"github.com/EgorAist/TP_DB_project/internal/storages/userStorage"
	"github.com/EgorAist/TP_DB_project/internal/storages/voteStorage"
	"math"
	"strings"
)

type Service interface {
	CreateForum(input models.ForumCreate) (models.Forum, error)
	GetForum(input models.ForumInput) (models.Forum, error)
	GetForumThreads(input models.ForumGetThreads) ([]models.Thread, error)
	GetForumUsers(input models.ForumGetUsers) ([]models.User, error)

	CreateUser(input models.User) ([]models.User, error)
	GetUser(nickname string) (models.User, error)
	UpdateUser(input models.User) (models.User, error)

	CreateThread(input models.Thread) (models.Thread, error)
	ThreadVote(input models.Vote) (models.Thread, error)
	GetThread(input models.ThreadInput) (models.Thread, error)
	UpdateThread(input models.ThreadUpdate) (models.Thread, error)
	GetThreadPosts(input models.ThreadGetPosts) ([]models.Post, error)

	GetPost(id int, related string) (models.PostFull, error)
	UpdatePost(input models.PostUpdate) (models.Post, error)

	Clear()
	Status() models.Status
}

type service struct {
	forumStorage forumStorage.Storage
	threadStorage threadStorage.Storage
	userStorage userStorage.Storage
	postStorage postStorage.Storage
	voteStorage voteStorage.Storage
	databaseService databaseService.Service
}

func NewService(forumStorage forumStorage.Storage, threadStorage threadStorage.Storage, userStorage userStorage.Storage, postStorage postStorage.Storage, voteStorage voteStorage.Storage, databaseService databaseService.Service) Service {
	return &service{
		forumStorage:  forumStorage,
		threadStorage: threadStorage,
		userStorage:   userStorage,
		postStorage:   postStorage,
		voteStorage:   voteStorage,
		databaseService: databaseService,
	}
}

func (s service) CreateForum(input models.ForumCreate) (models.Forum, error) {
	forum, err := s.forumStorage.CreateForum(input)
	if err != nil && err.Error() == "409" {
		oldForum, err := s.forumStorage.GetDetails(models.ForumInput{Slug: input.Slug})
		if err != nil {
			return models.Forum{}, err
		}

		return oldForum, models.Error{Code: "409", Message: "conflict slug"}
	}

	if err != nil {
		return models.Forum{}, err
	}

	return forum, nil
}

func (s service) GetForum(input models.ForumInput) (models.Forum, error) {
	return s.forumStorage.GetDetails(input)
}

func (s service) GetForumThreads(input models.ForumGetThreads) ([]models.Thread, error) {
	err := s.forumStorage.CheckIfForumExists(models.ForumInput{Slug: input.Slug})
	if err != nil {
		return []models.Thread{}, err
	}
	if input.Limit == 0 {
		input.Limit = math.MaxInt32
	}
	return s.threadStorage.GetThreadsByForum(input)
}

func (s service) GetForumUsers(input models.ForumGetUsers) ([]models.User, error) {
	forumID, err := s.forumStorage.GetForumID(models.ForumInput{Slug: input.Slug})
	if err != nil {
		return []models.User{}, err
	}

	if input.Limit == 0 {
		input.Limit = math.MaxInt32
	}
	return s.userStorage.GetUsers(input, forumID)
}

func (s service) CreateUser(input models.User) ([]models.User, error) {
	user, err := s.userStorage.CreateUser(input)

	if err == nil {
		return []models.User{user}, err
	}

	users := make([]models.User, 0)
	if err.Error() == "409" {
		userNick, err := s.userStorage.GetProfile(input.Nickname)
		if err != nil && err.Error() != "404"{
			return []models.User{}, err
		}
		if err == nil {
			users = append(users, userNick)
		}

		if strings.ToLower(userNick.Email) == strings.ToLower(input.Email){
			return users, models.Error{Code: "409", Message: "conflict"}
		}

		userEmail, err := s.userStorage.GetEmailConflictUser(input.Email)
		if err != nil && err.Error() != "404"{
			return []models.User{}, err
		}
		if err == nil {
			users = append(users, userEmail)
		}

		return users, models.Error{Code: "409", Message: "conflict"}
	}

	return []models.User{}, err
}

func (s service) GetUser(nickname string) (models.User, error) {
	return s.userStorage.GetProfile(nickname)
}

func (s service) UpdateUser(input models.User) (models.User, error) {
	if input.Email == "" && input.Fullname == "" && input.About == "" {
		return s.userStorage.GetProfile(input.Nickname)
	}
	return s.userStorage.UpdateProfile(input)
}

func (s service) CreateThread(input models.Thread) (models.Thread, error) {
	thread, err := s.threadStorage.CreateThread(input)
	if err == nil {
		err = s.forumStorage.UpdateThreadsCount(models.ForumInput{Slug: input.Forum})
		if err != nil {
			return models.Thread{}, err
		}
		userID, err := s.userStorage.GetUserIDByNickname(input.Author)
		if err != nil {
			return models.Thread{}, err
		}

		forumID, err := s.forumStorage.GetForumID(models.ForumInput{Slug: input.Forum})
		if err != nil {
			return models.Thread{}, err
		}

		err = s.forumStorage.AddUserToForum(userID, forumID)
		if err != nil && err.Error() != "409" {
			return models.Thread{}, err
		}

		return thread, nil
	}

	if err.Error() == "409"  {
		oldThread, err := s.threadStorage.GetDetails(models.ThreadInput{Slug: input.Slug})
		if err == nil {
			return oldThread, models.Error{Code: "409"}
		}
		return thread, err
	}

	return thread, err
}

func (s service) ThreadVote(input models.Vote) (models.Thread, error) {
	thread, err := s.threadStorage.CheckThreadIfExists(input.Thread)
	if err != nil {
		return models.Thread{}, err
	}
	input.Thread = thread

	var updateFlag bool

	checkThread, err := s.voteStorage.CheckDoubleVote(input)
	if err != nil {
		if err.Error() == "409" {
			return checkThread, nil
		}
		if err.Error() == "500" {
			return models.Thread{}, err
		}
		if err.Error() == "101" {
			updateFlag = true
		}
	}

	output, err := s.voteStorage.CreateVote(input, updateFlag)
	if err != nil {
		return models.Thread{}, err
	}

	return output, nil
}

func (s service) GetThread(input models.ThreadInput) (models.Thread, error) {
	return s.threadStorage.GetDetails(input)
}

func (s service) UpdateThread(input models.ThreadUpdate) (models.Thread, error) {
	return s.threadStorage.UpdateThread(input)
}

func (s service) GetThreadPosts(input models.ThreadGetPosts) ([]models.Post, error) {
	thread, err := s.threadStorage.CheckThreadIfExists(input.ThreadInput)
	if err != nil {
		return []models.Post{}, err
	}

	input.ThreadInput = thread

	if input.Limit == 0 {
		input.Limit = math.MaxInt32
	}
	return s.postStorage.GetPostsByThread(input)
}

func (s service) GetPost(id int, related string) (models.PostFull, error) {
	postFull := models.PostFull{
		Author: nil,
		Forum:  nil,
		Post:   nil,
		Thread: nil,
	}
	post := new(models.Post)
	err := s.postStorage.GetPostDetails(models.PostInput{ID: id}, post)
	postFull.Post = post
	if err != nil {
		return models.PostFull{}, err
	}

	author := new(models.User)
	if strings.Contains(related, "user") {
		err = s.userStorage.GetUserForPost(postFull.Post.Author, author)
		postFull.Author = author
		if err != nil {
			return models.PostFull{}, err
		}
	}

	forum := new(models.Forum)
	if strings.Contains(related, "forum") {
		err = s.forumStorage.GetForumForPost(postFull.Post.Forum, forum)
		postFull.Forum = forum
		if err != nil {
			return models.PostFull{}, err
		}
	}

	thread := new(models.Thread)
	if strings.Contains(related, "thread") {
		err = s.threadStorage.GetThreadForPost(postFull.Post.ThreadInput, thread)
		postFull.Thread = thread
		if err != nil {
			return models.PostFull{}, err
		}
	}

	return postFull, nil
}

func (s service) UpdatePost(input models.PostUpdate) (models.Post, error) {
	return s.postStorage.UpdatePost(input)
}

func (s service) Clear() {
	err := s.databaseService.Clear()
	if err != nil {
		fmt.Println(err)
	}
}

func (s service) Status() models.Status {
	status, err := s.databaseService.Status()
	if err != nil {
		fmt.Println(err)
	}
	return status
}
