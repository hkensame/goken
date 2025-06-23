package httputil

import (
	"net/http"
	"strings"

	"github.com/hkensame/goken/pkg/errors"

	"github.com/gin-gonic/gin"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

type JsonResult struct {
	Code int     `json:"code"`
	Data Mapping `json:"data"`
}

type Mapping map[string]interface{}

func newJsonResult(code int, data Mapping) JsonResult {
	if data == nil {
		data = make(Mapping)
	}
	return JsonResult{
		Code: code,
		Data: data,
	}
}

func (r JsonResult) WriteResponse(c *gin.Context) {
	c.JSON(r.Code, gin.H{
		"code": r.Code,
		"data": r.Data,
	})
}

func WriteResponse(c *gin.Context, code int, data Mapping) {
	newJsonResult(code, data).WriteResponse(c)
}

func WriteError(c *gin.Context, code int, err error, data Mapping) {
	if data == nil {
		data = make(Mapping)
	}
	data["error"] = err.Error()
	newJsonResult(code, data).WriteResponse(c)
}

func WriteRpcError(c *gin.Context, err error) {
	e := errors.ExtractKerrorFromGRPC(err)
	if ke, ok := e.(*errors.Kerror); ok {
		data := Mapping{"error": ke.Error()}
		newJsonResult(ke.HTTPCode(), data).WriteResponse(c)
	} else {
		data := Mapping{"error": "服务器内部错误,请稍后再试"}
		newJsonResult(http.StatusInternalServerError, data).WriteResponse(c)
	}
}

func WriteValidateError(c *gin.Context, trans ut.Translator, err error) {
	verr, ok := err.(validator.ValidationErrors)
	if !ok {
		WriteError(c, http.StatusInternalServerError, err, nil)
		return
	}
	translated := translateErr(verr, trans)
	WriteResponse(c, http.StatusBadRequest, Mapping{"errors": translated})
}

func translateErr(err validator.ValidationErrors, trans ut.Translator) map[string]string {
	simplify := func(msg map[string]string) map[string]string {
		res := make(map[string]string, len(msg))
		for k, v := range msg {
			if idx := strings.Index(k, "."); idx != -1 {
				k = k[idx+1:]
			}
			res[k] = v
		}
		return res
	}
	return simplify(err.Translate(trans))
}
