package utils

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
}

type PaginatedData struct {
	Data     interface{} `json:"data"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:      200,
		Message:   "success",
		Data:      data,
		Timestamp: time.Now().Unix(),
	})
}

func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:      200,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().Unix(),
	})
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Code:      201,
		Message:   "created",
		Data:      data,
		Timestamp: time.Now().Unix(),
	})
}

func Paginated(c *gin.Context, data interface{}, total int64, page, pageSize int) {
	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "success",
		Data: PaginatedData{
			Data:     data,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		},
		Timestamp: time.Now().Unix(),
	})
}

func Error(c *gin.Context, httpStatus int, code int, message string) {
	c.JSON(httpStatus, Response{
		Code:      code,
		Message:   message,
		Data:      nil,
		Timestamp: time.Now().Unix(),
	})
}

func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, 400, message)
}

func Unauthorized(c *gin.Context, message string) {
	if message == "" {
		message = "未授权访问"
	}
	Error(c, http.StatusUnauthorized, 401, message)
}

func Forbidden(c *gin.Context, message string) {
	if message == "" {
		message = "无权访问"
	}
	Error(c, http.StatusForbidden, 403, message)
}

func NotFound(c *gin.Context, message string) {
	if message == "" {
		message = "资源不存在"
	}
	Error(c, http.StatusNotFound, 404, message)
}

func TooManyRequests(c *gin.Context, message string) {
	if message == "" {
		message = "请求过于频繁，请稍后再试"
	}
	Error(c, http.StatusTooManyRequests, 429, message)
}

func InternalError(c *gin.Context, message string) {
	if message == "" {
		message = "服务器内部错误"
	}
	Error(c, http.StatusInternalServerError, 500, message)
}
