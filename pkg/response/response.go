package response

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Response struct {
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func JSONResponse(c *gin.Context, statusCode int, status bool, message string, data interface{}) {
	c.JSON(statusCode, Response{
		Status:  status,
		Message: message,
		Data:    data,
	})
}

func ErrorResponse(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, Response{
		Status:  false,
		Message: message,
		Error:   message,
	})
}

func SuccessResponse(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Status:  true,
		Message: message,
		Data:    data,
	})
}
