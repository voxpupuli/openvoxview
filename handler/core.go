package handler

import (
	"time"

	"github.com/gin-gonic/gin"
)

func baseResponse() map[string]any {
	return gin.H{
		"Timestamp": time.Now().Unix(),
	}
}

func NewErrorResponse(err error) map[string]any {
	resp := baseResponse()
	resp["Error"] = err.Error()

	return resp
}

func NewSuccessResponse(data interface{}) map[string]any {
	resp := baseResponse()
	resp["Data"] = data

	return resp
}
