package web

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"routing/template"
	"strconv"
)

type Context struct {
	Req              *http.Request
	Resp             http.ResponseWriter
	PathParams       map[string]string
	cacheQueryValues url.Values
	MatchedRoute     string
	RespData         []byte
	RespStatusCode   int
	tplEngine        template.TemplateEngine
}

type stringValue struct {
	val string
	err error
}

func (c *Context) BindJson(val any) error {
	if c.Req.Body == nil {
		return errors.New("web: body 为 nil")
	}
	decoder := json.NewDecoder(c.Req.Body)
	return decoder.Decode(val)
}

func (c *Context) FormValue(key string) stringValue {
	err := c.Req.ParseForm()
	if err != nil {
		return stringValue{
			err: err,
		}
	}
	val, ok := c.Req.Form[key]
	if !ok {
		return stringValue{
			err: errors.New("form param not found"),
		}
	}
	return stringValue{
		val: val[0],
	}
}

func (s stringValue) AsInt64() (int64, error) {
	if s.err != nil {
		return 0, s.err
	}
	return strconv.ParseInt(s.val, 10, 64)
}

func (s stringValue) AsString() (string, error) {
	if s.err != nil {
		return "", s.err
	}
	return s.val, nil
}

func (c *Context) QueryValue(key string) stringValue {
	if c.cacheQueryValues == nil {
		c.cacheQueryValues = c.Req.URL.Query()
	}
	val, ok := c.cacheQueryValues[key]
	if !ok {
		return stringValue{
			err: errors.New("query param not found"),
		}
	}
	return stringValue{
		val: val[0],
	}
}

func (c *Context) PathValue(key string) stringValue {
	val, ok := c.PathParams[key]
	if !ok {
		return stringValue{
			err: errors.New("query param not found"),
		}
	}
	return stringValue{
		val: val,
	}
}

func (c *Context) RespJson(status int, val any) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}
	c.RespStatusCode = status
	c.RespData = data
	return nil
}

func (c *Context) SetCookie(ck *http.Cookie) {
	http.SetCookie(c.Resp, ck)
}

func (c *Context) Render(tplName string, data any) error {
	var err error
	c.RespData, err = c.tplEngine.Render(c.Req.Context(), tplName, data)
	if err != nil {
		c.RespStatusCode = http.StatusInternalServerError
		return err
	}
	c.RespStatusCode = http.StatusOK
	return nil
}
