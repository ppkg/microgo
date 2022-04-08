// http testing
// mayb 2019-07-02
package microgo

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/ppkg/microgo/utils"

	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

var route *gin.Engine
var auth string

func InitGinTest(token string, fn func(e *gin.Engine)) {
	auth = token
	route = gin.New()
	route.Use(logger())
	route.Use(recoverGin)
	fn(route)
}

func mapToValues(mp map[string]interface{}) url.Values {
	v := url.Values{}
	for key, val := range mp {
		switch t := val.(type) {
		case int:
			v.Add(key, strconv.Itoa(t))
		case float64:
			v.Add(key, strconv.FormatFloat(t, 'E', -1, 64))
		case float32:
			v.Add(key, strconv.FormatFloat(float64(t), 'E', -1, 32))
		default:
			v.Add(key, val.(string))
		}
	}
	return v
}

func action(uri string, httpMethod string, contentType string, param map[string]interface{}) (string, int) {
	var req *http.Request
	switch httpMethod {
	case http.MethodGet:
		if param != nil {
			uri += "?" + mapToValues(param).Encode()
		}

		req = httptest.NewRequest(httpMethod, uri, nil)
	case http.MethodPost:
		httpMethod = http.MethodPost
		var reader io.Reader

		if contentType == "application/x-www-form-urlencoded" {
			reader = strings.NewReader(mapToValues(param).Encode())
		} else if contentType == "application/json;charset=UTF-8" {
			byteData, _ := json.Marshal(param)
			reader = bytes.NewReader(byteData)
		}

		req = httptest.NewRequest(httpMethod, uri, reader)
		req.Header.Set("Content-Type", contentType)
	default:
		panic("error " + httpMethod)
	}

	req.Header.Add("Authorization", "bearer "+auth)

	var result *http.Response
	if strings.HasPrefix(uri, "http") {
		req.RequestURI = ""
		if r, err := utils.HttpDefaultClient.Do(req); err != nil {
			panic(err)
		} else {
			result = r
		}
	} else {
		res := httptest.NewRecorder()
		route.ServeHTTP(res, req)
		result = res.Result()
	}

	defer result.Body.Close()
	body, _ := ioutil.ReadAll(result.Body)
	return string(body), result.StatusCode
}

func TestGet(uri string, param map[string]interface{}) (string, int) {
	return action(uri, http.MethodGet, "", param)
}

func TestPostForm(uri string, param map[string]interface{}) (string, int) {
	return action(uri, http.MethodPost, "application/x-www-form-urlencoded", param)
}

func TestPostJson(uri string, param map[string]interface{}) (string, int) {
	return action(uri, http.MethodPost, "application/json;charset=UTF-8", param)
}
