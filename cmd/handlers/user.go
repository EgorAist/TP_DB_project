package handlers

import (
	"encoding/json"
	"github.com/EgorAist/TP_DB_project/internal/models"
	"github.com/valyala/fasthttp"
	"log"
)

func (h handler) UserCreate(c *fasthttp.RequestCtx) {
	userInput := &models.User{}
	userInput.Nickname = c.UserValue("nickname").(string)
	err := userInput.UnmarshalJSON(c.PostBody())
	if err != nil {
		log.Println(err)
		return
	}

	user, err := h.Service.CreateUser(*userInput)


	if err != nil {
		status, respErr, _ := h.ConvertError(err)
		if status == fasthttp.StatusConflict {
			response, _ := json.Marshal(user)
			h.WriteResponse(c, status, response)
			return
		}
		h.WriteResponse(c, status, respErr)
		return
	}

	response, _ := json.Marshal(user[0])

	h.WriteResponse(c, fasthttp.StatusCreated, response)
	return
}

func (h handler) UserGet(c *fasthttp.RequestCtx) {
	nickname := c.UserValue("nickname").(string)

	user, err := h.Service.GetUser(nickname)
	if err != nil {
		status, respErr, _ := h.ConvertError(err)
		h.WriteResponse(c, status, respErr)
		return
	}

	response, _ := json.Marshal(user)

	h.WriteResponse(c, fasthttp.StatusOK, response)
	return
}

func (h handler) UserUpdate(c *fasthttp.RequestCtx) {
	userInput := &models.User{}
	userInput.Nickname = c.UserValue("nickname").(string)
	err := userInput.UnmarshalJSON(c.PostBody())
	if err != nil {
		log.Println(err)
		return
	}

	user, err := h.Service.UpdateUser(*userInput)
	if err != nil {
		status, respErr, _ := h.ConvertError(err)
		h.WriteResponse(c, status, respErr)
		return
	}

	response, _ := json.Marshal(user)

	h.WriteResponse(c, fasthttp.StatusOK, response)
	return
}

