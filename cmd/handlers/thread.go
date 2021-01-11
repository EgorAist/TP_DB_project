package handlers

import (
	"encoding/json"
	"github.com/EgorAist/TP_DB_project/internal/models"
	"github.com/valyala/fasthttp"
	"log"
)

func (h handler) ThreadCreate(c *fasthttp.RequestCtx) {
	threadInput := &models.Thread{}
	err :=threadInput.UnmarshalJSON(c.PostBody())
	if err != nil {
		log.Println(err)
		return
	}

	threadInput.Forum = c.UserValue("slug").(string)

	thread, err := h.Service.CreateThread(*threadInput)
	if err != nil {
		status, respErr, _ := h.ConvertError(err)
		if status == fasthttp.StatusConflict {
			response, _ := thread.MarshalJSON()
			h.WriteResponse(c, status, response)
			return
		}
		h.WriteResponse(c, status, respErr)
		return
	}

	response, _ := json.Marshal(thread)

	h.WriteResponse(c, fasthttp.StatusCreated, response)
	return
}

func (h handler) ThreadVote(c *fasthttp.RequestCtx) {
	voteInput := &models.Vote{}

	err := voteInput.UnmarshalJSON(c.PostBody())
	if err != nil {
		log.Println(err)
		return
	}

	voteInput.Thread = SlagOrID(c)

	thread, err := h.Service.ThreadVote(*voteInput)
	if err != nil {
		status, respErr, _ := h.ConvertError(err)
		h.WriteResponse(c, status, respErr)
		return
	}

	response, _ := json.Marshal(thread)

	h.WriteResponse(c, fasthttp.StatusOK, response)
	return
}

func (h handler) ThreadGet(c *fasthttp.RequestCtx) {
	threadInput := SlagOrID(c)

	thread, err := h.Service.GetThread(threadInput)
	if err != nil {
		status, respErr, _ := h.ConvertError(err)
		h.WriteResponse(c, status, respErr)
		return
	}

	response, _ := json.Marshal(thread)

	h.WriteResponse(c, fasthttp.StatusOK, response)
	return
}

func (h handler) ThreadUpdate(c *fasthttp.RequestCtx) {
	threadInput := &models.ThreadUpdate{}
	err := threadInput.UnmarshalJSON(c.PostBody())
	if err != nil {
		log.Println(err)
		return
	}

	slagOrID := SlagOrID(c)

	threadInput.ThreadID = slagOrID.ThreadID
	threadInput.Slug = slagOrID.Slug

	thread, err := h.Service.UpdateThread(*threadInput)
	if err != nil {
		status, respErr, _ := h.ConvertError(err)
		h.WriteResponse(c, status, respErr)
		return
	}

	response, _ := json.Marshal(thread)

	h.WriteResponse(c, fasthttp.StatusOK, response)
	return
}

func (h handler) ThreadGetPosts(c *fasthttp.RequestCtx) {
	threadInput := models.ThreadGetPosts{
		Limit:    c.QueryArgs().GetUintOrZero("limit"),
		Since:    c.QueryArgs().GetUintOrZero("since"),
		Sort:     string(c.QueryArgs().Peek("sort")),
		Desc:     getBool("desc", c.QueryArgs()),
	}

	slugOrID := SlagOrID(c)
	threadInput.ThreadID = slugOrID.ThreadID
	threadInput.Slug = slugOrID.Slug

	posts, err := h.Service.GetThreadPosts(threadInput)
	if err != nil {
		status, respErr, _ := h.ConvertError(err)
		h.WriteResponse(c, status, respErr)
		return
	}

	response, _ := json.Marshal(posts)

	h.WriteResponse(c, fasthttp.StatusOK, response)
	return
}

