package handlers

import (
	"encoding/json"
	"github.com/valyala/fasthttp"
)

func (h handler) Clear(c *fasthttp.RequestCtx) {
	h.Service.Clear()

	c.SetContentType("application/json")
	c.SetStatusCode(fasthttp.StatusOK)
	return
}

func (h handler) Status(c *fasthttp.RequestCtx) {
	status := h.Service.Status()

	response, _ := json.Marshal(status)

	h.WriteResponse(c, fasthttp.StatusOK, response)
	return
}

