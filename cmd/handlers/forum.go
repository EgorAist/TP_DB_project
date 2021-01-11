package handlers

import (
	"encoding/json"
	"github.com/EgorAist/TP_DB_project/internal/models"
	"github.com/valyala/fasthttp"
	"log"
)


func (h handler) ForumCreate(c *fasthttp.RequestCtx) {
	forumInput := &models.ForumCreate{}
	err := forumInput.UnmarshalJSON(c.PostBody())
	if err != nil {
		log.Println(err)
		return
	}

	forum, err := h.Service.CreateForum(*forumInput)
	if err != nil {
		status, respErr, _ := h.ConvertError(err)
		if status == fasthttp.StatusConflict {
			response, _ := forum.MarshalJSON()
			h.WriteResponse(c, status, response)
			return
		}
		h.WriteResponse(c, status, respErr)
		return
	}

	response, _ := forum.MarshalJSON()

	h.WriteResponse(c, fasthttp.StatusCreated, response)
}

func (h handler) ForumGet(c *fasthttp.RequestCtx) {
	forumInput := models.ForumInput{}
	forumInput.Slug = c.UserValue("slug").(string)
	forum, err := h.Service.GetForum(forumInput)
	if err != nil {
		status, respErr, _ := h.ConvertError(err)
		h.WriteResponse(c, status, respErr)
		return
	}

	response, _ := forum.MarshalJSON()

	h.WriteResponse(c, fasthttp.StatusOK, response)
}

func (h handler) ForumGetThreads(c *fasthttp.RequestCtx) {
	input := models.ForumGetThreads{
		Slug:  c.UserValue("slug").(string),
		Limit: c.QueryArgs().GetUintOrZero("limit"),
		Since: string(c.QueryArgs().Peek("since")),
		Desc:  getBool("desc", c.QueryArgs()),
	}

	threads, err := h.Service.GetForumThreads(input)
	if err != nil {
		status, respErr, _ := h.ConvertError(err)
		h.WriteResponse(c, status, respErr)
		return
	}

	response, _ := json.Marshal(threads)

	h.WriteResponse(c, fasthttp.StatusOK, response)
	return
}

func (h handler) ForumGetUsers(c *fasthttp.RequestCtx) {
	input := models.ForumGetUsers{
		Slug:  c.UserValue("slug").(string),
		Limit: c.QueryArgs().GetUintOrZero("limit"),
		Since: string(c.QueryArgs().Peek("since")),
		Desc:  getBool("desc", c.QueryArgs()),
	}

	users, err := h.Service.GetForumUsers(input)
	if err != nil {
		status, respErr, _ := h.ConvertError(err)
		h.WriteResponse(c, status, respErr)
		return
	}

	response, _ := json.Marshal(users)

	h.WriteResponse(c, fasthttp.StatusOK, response)
	return
}
