package httputil

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"

	"github.com/hkensame/goken/pkg/errors"
)

// JsonResult 统一响应结构体
type JsonResult struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func newJsonResult(code int, msg string, data interface{}) JsonResult {
	return JsonResult{
		Code:    code,
		Message: msg,
		Data:    data,
	}
}

func (r JsonResult) WriteResponse(c *gin.Context, abort bool) {
	c.JSON(r.Code, gin.H{
		"code": r.Code,
		"msg":  r.Message,
		"data": r.Data,
	})
	if abort {
		c.Abort()
	}
}

// WriteResponse 写标准响应
func WriteResponse(c *gin.Context, code int, msg string, data interface{}, abort bool) {
	newJsonResult(code, msg, data).WriteResponse(c, abort)
}

// WriteError 写普通 error 错误响应
func WriteError(c *gin.Context, code int, err error, abort bool) {
	writeErrorJson(c, code, err.Error(), abort)
}

// WriteRpcError 写 RPC 错误响应，自动提取 gRPC 错误码与消息
func WriteRpcError(c *gin.Context, err error, abort bool) {
	e := errors.ExtractCodeErrorFromGRPC(err)
	if cerr := errors.ExtractCoderFromError(e); cerr != nil {
		writeErrorJson(c, cerr.HTTPCode(), cerr.Message(), abort)
	} else {
		writeErrorJson(c, http.StatusInternalServerError, "服务器内部错误，请稍后再试", abort)
	}
}

// WriteValidateError 写入表单校验错误响应
func WriteValidateError(c *gin.Context, trans ut.Translator, err error, abort bool) {
	verr, ok := err.(validator.ValidationErrors)
	if !ok {
		writeErrorJson(c, http.StatusInternalServerError, err.Error(), abort)
		return
	}
	errors := translateErr(verr, trans)
	newJsonResult(http.StatusBadRequest, "传入参数错误", gin.H{"err": errors}).WriteResponse(c, abort)
}

// translateErr 将校验错误翻译成 map[string]string，移除结构体前缀
func translateErr(verr validator.ValidationErrors, trans ut.Translator) map[string]string {
	result := make(map[string]string)
	for field, msg := range verr.Translate(trans) {
		// 去掉结构体名前缀（如 CreateUserReq.Username -> Username）
		if idx := strings.Index(field, "."); idx != -1 {
			field = field[idx+1:]
		}
		result[field] = msg
	}
	return result
}

// MustIsMethod 判断当前请求方法是否匹配指定方法
func MustIsMethod(c *gin.Context, methods ...string) bool {
	reqMethod := c.Request.Method
	for _, m := range methods {
		if reqMethod == m {
			return true
		}
	}
	return false
}

// writeErrorJson 内部错误响应封装
func writeErrorJson(c *gin.Context, code int, msg string, abort bool) {
	newJsonResult(code, "", gin.H{"err": msg}).WriteResponse(c, abort)
}
