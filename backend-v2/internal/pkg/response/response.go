package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

func Result(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Data:    data,
		Message: message,
	})
}

func Success(c *gin.Context, data interface{}) {
	Result(c, 0, "success", data)
}

func Error(c *gin.Context, code int, message string) {
	Result(c, code, message, nil)
}
