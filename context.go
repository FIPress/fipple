package fipple

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/fipress/fiplog"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"text/template"
)

const (
	//errInternalError = "500 internal server error"
	//errNotFount = "404 page not found"
	//errMethodNotAllowed = "405 method not allowed"
	contentType = "Content-Type"
	typeJson    = "application/json; charset=utf-8"
	typeHtml    = "text/html; charset=utf-8"
	typeXML     = "text/xml; charset=utf-8"
)

type RpcRequest struct {
	path string
	id   int
	//next *Request
}

type RpcResponse struct {
	path string
	id   int
}

type Encoder interface {
	Encode(v interface{}) error
}

type Decoder interface {
	Decode(v interface{}) error
}

type Context struct {
	req        *http.Request
	rw         http.ResponseWriter
	params     map[string]interface{}
	query      url.Values
	paramArray []string //intermediary
}

type Fields struct {
	m map[string]interface{}
}

func (ctx *Context) Request() *http.Request {
	return ctx.req
}

func (ctx *Context) ResponseWriter() http.ResponseWriter {
	return ctx.rw
}

func (f Fields) GetStringField(name string) string {
	val, ok := f.m[name]
	if !ok {
		return ""
	}
	switch v := val.(type) {
	case string:
		return v
	default:
		return ""
	}
}

func (f Fields) GetIntField(name string) int {
	val, ok := f.m[name]
	if !ok {
		return 0
	}
	switch v := val.(type) {
	case int:
		return v
	default:
		return 0
	}
}

func NewContext(rw http.ResponseWriter, req *http.Request) *Context {
	getLogger().Debug("receive request, remote addr:", req.RemoteAddr, ",request uri:", req.RequestURI)
	//query :=
	return &Context{req, rw, make(map[string]interface{}), req.URL.Query(), nil}
}

func (ctx *Context) GetQuery(key string) string {
	return ctx.query.Get(key)
}

func (ctx *Context) GetIntQuery(key string, dflt int) int {
	raw := ctx.query.Get(key)
	if raw == "" {
		return dflt
	}
	i, e := strconv.Atoi(raw)
	if e != nil {
		return dflt
	} else {
		return i
	}
}

func (ctx *Context) GetParam(key string) interface{} {
	v, ok := ctx.params[key]
	if ok {
		return v
	} else {
		return ""
	}
}

func (ctx *Context) GetStringParam(key string) (s string) {
	v, ok := ctx.params[key]
	if ok {
		switch val := v.(type) {
		case string:
			return val
		default:
			return
		}
	} else {
		return
	}
}

func (ctx *Context) GetIntParam(key string) (i int) {
	v, ok := ctx.params[key]

	if ok {
		switch val := v.(type) {
		case int:
			return val
		case string:
			i, _ = strconv.Atoi(val)
			return
		default:
			return
		}
	} else {
		return
	}
}

func (ctx *Context) GetIntParamWithDefault(key string, defaultVal int) int {
	v, ok := ctx.params[key]
	if ok {
		switch val := v.(type) {
		case int:
			return val
		default:
			return defaultVal
		}
	} else {
		return defaultVal
	}
}

func (ctx *Context) GetEntity(v interface{}) (err error) {
	body, err := ioutil.ReadAll(ctx.req.Body)
	if err != nil {
		return
	}

	return json.Unmarshal(body, v)
}

func (ctx *Context) GetFields() (fields *Fields, err error) {
	body, err := ioutil.ReadAll(ctx.req.Body)
	if err != nil {
		return
	}

	fiplog.GetLogger().Debug("body:", body)
	fiplog.GetLogger().Debug("body string:", string(body))
	var m map[string]interface{}
	json.Unmarshal(body, &m)
	fields = &Fields{m}
	return
}

/*func (ctx *Context) GetStringField(name string) (s string) {
	val,err := ctx.GetField(name)
	fiplog.GetLogger().Debug("name",name,",val:",val)
	if err != nil {
		return
	}

	switch v := val.(type) {
	case string:
		return v
	default:
		return
	}
	return
}

func (ctx *Context) GetIntField(name string) (i int) {
	val,err := ctx.GetField(name)
	if err != nil {
		return
	}

	switch v := val.(type) {
	case float64:
		return int(v)
	default:
		return
	}
	return
}*/

func (ctx *Context) GetPlainReq() (body []byte) {
	body, err := ioutil.ReadAll(ctx.req.Body)
	if err != nil {
		return
	}
	return
}

func (ctx *Context) GetReqHeader(key string) (val string) {
	return ctx.req.Header.Get(key)
}

func (ctx *Context) SetHeader(key, val string) {
	ctx.rw.Header().Set(key, val)
}

func (ctx *Context) ServeBody(content []byte) {
	//todo:gzip
	ctx.rw.Write(content)
}

func (ctx *Context) ServeString(o ...interface{}) {
	//fmt.Fprint(ctx.rw, o)
	s := fmt.Sprint(o...)
	ctx.rw.Write([]byte(s))
}

func (ctx *Context) ServeJson(o interface{}) {
	bs, err := json.Marshal(o)
	if err != nil {
		getLogger().Error("marshal json error:", err)
		ctx.InternalError()
	}
	ctx.SetHeader(contentType, typeJson)
	ctx.rw.Write(bs)
}

func (ctx *Context) ServeHtml(content []byte) {
	ctx.SetHeader(contentType, typeHtml)
	ctx.rw.Write(content)
}

func (ctx *Context) ServeXML(content []byte) {
	ctx.SetHeader(contentType, typeXML)
	ctx.rw.Write(content)
}

func (ctx *Context) ServeStatic(name string, content []byte) {
	ctype := mime.TypeByExtension(filepath.Ext(name))
	if ctype == "" {
		ctype = typeHtml
	}
	ctx.SetHeader(contentType, ctype)
	ctx.rw.Write(content)
}

func (ctx *Context) ServeHtmlFile(path string) {
	ctx.writeFile(path, typeHtml)
}

func (ctx *Context) ServeFile(path string) {
	ctype := mime.TypeByExtension(filepath.Ext(path))
	if ctype == "" {
		ctype = typeHtml
	}
	ctx.writeFile(path, ctype)
}

func (ctx *Context) writeFile(path, fileType string) {
	ctx.SetHeader(contentType, fileType)
	f, err := os.Open(path)
	if err != nil {
		ctx.ServeStatus(http.StatusNotFound)
		return
	}
	io.Copy(ctx.rw, f)
}

func (ctx *Context) ServeByTemplate(templ *template.Template, data interface{}) {
	ctx.SetHeader(contentType, typeHtml)

	var err error
	if compress {
		ctx.SetHeader("Content-Encoding", "gzip")
		w := gzip.NewWriter(ctx.rw)
		err = templ.Execute(w, data) // templates.ExecuteTemplate(w, templ, data)
		defer w.Close()
	} else {
		//err = templates.ExecuteTemplate(ctx.rw, templ, data)
		err = templ.Execute(ctx.rw, data)
	}
	//log.Println("templ:",templ,"data:",data)

	if err != nil {
		//http.Error(ctx.rw, err.Error(), http.StatusInternalServerError)
		ctx.InternalError()
		getLogger().Error(ctx.req, "Error sending response", err)
	}
}

func (ctx *Context) ServeStatus(code int) {
	ctx.rw.WriteHeader(code)
}

func (ctx *Context) Ok() {
	ctx.ServeStatus(http.StatusOK)
}

func (ctx *Context) InternalError() {
	ctx.ServeStatus(http.StatusInternalServerError)
}

func (ctx *Context) OkOrError(ok bool) {
	if ok {
		ctx.Ok()
	} else {
		ctx.InternalError()
	}
}

func (ctx *Context) BadRequest() {
	ctx.ServeStatus(http.StatusBadRequest)
}

func (ctx *Context) Unauthorized() {
	ctx.ServeStatus(http.StatusUnauthorized)
}

func (ctx *Context) PageNotFound() {
	//if no custom handler
	ctx.ServeStatus(http.StatusNotFound)
}

func (ctx *Context) Conflict() {
	ctx.ServeStatus(http.StatusConflict)
}

func (ctx *Context) MethodNotAllowed() {
	ctx.ServeStatus(http.StatusMethodNotAllowed)
}

func (ctx *Context) addParam(p string) {
	if ctx.paramArray == nil {
		ctx.paramArray = make([]string, 1)
		ctx.paramArray[0] = p
	} else {
		ctx.paramArray = append(ctx.paramArray, p)
	}
}

func (ctx *Context) buildNamedParams(names []string) {
	l := len(ctx.paramArray)
	if l != len(names) {
		ctx.PageNotFound()
		return
	}
	for i := 0; i < l; i++ {
		ctx.params[names[i]] = ctx.paramArray[i]
	}
}
