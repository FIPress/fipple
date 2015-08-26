package fipple

import (
	"net/http"
	"fmt"
	"encoding/json"
	"log"
	"compress/gzip"
	"path/filepath"
	"mime"
	"io/ioutil"
	"fiplog"
	"net/url"
	"strconv"
)

const (
	//errInternalError = "500 internal server error"
	//errNotFount = "404 page not found"
	//errMethodNotAllowed = "405 method not allowed"
	contentType = "Content-Type"
	typeJson = "application/json; charset=utf-8"
	typeHtml = "text/html; charset=utf-8"
)

type RpcRequest struct {
	path string
	id int
	//next *Request
}

type RpcResponse struct {
	path string
	id int
}

type Encoder interface {
	Encode(v interface{}) error
}

type Decoder interface {
	Decode(v interface{}) error
}

type Context struct {
	req *http.Request
	rw http.ResponseWriter
	params map[string]interface {}
	query url.Values
}

type Fields struct {
	m map[string]interface {}
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
	fiplog.GetLogger().Debug("receive request, remote addr:",req.RemoteAddr,",request uri:",req.RequestURI)
	//query :=
	return &Context{req,rw,make(map[string]interface {}),req.URL.Query()}
}

func (ctx *Context) GetQuery(key string) string {
	return ctx.query.Get(key)
}

func (ctx *Context) GetIntQuery(key string, dflt int) int {
	raw := ctx.query.Get(key)
	if raw == "" {
		return dflt
	}
	i,e := strconv.Atoi(raw)
	if e != nil {
		return dflt
	} else {
		return i
	}
}

func (ctx *Context) GetParam(key string) interface {} {
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
		default:
			return
		}
	} else {
		return
	}
}

func (ctx *Context) GetEntity(v interface {}) (err error) {
	body,err := ioutil.ReadAll(ctx.req.Body)
	if err != nil {
		return
	}

	return json.Unmarshal(body,v)
}

func (ctx *Context) GetFields() (fields *Fields, err error) {
	body,err := ioutil.ReadAll(ctx.req.Body)
	if err != nil {
		return
	}

	fiplog.GetLogger().Debug("body:",body)
	var m map[string]interface {}
	json.Unmarshal(body,&m)
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

func (ctx *Context) GetReqHeader(key string) ( val string) {
	return ctx.req.Header.Get(key)
}

func (ctx *Context) SetHeader(key,val string) {
	ctx.rw.Header().Set(key,val)
}

func (ctx *Context) ServeBody(content []byte) {
	//todo:gzip
	ctx.rw.Write(content)
}

func (ctx *Context) ServeString(o ...interface {}) {
	fmt.Fprint(ctx.rw,o)
}

func (ctx *Context) ServeJson(o interface {}) {
	bs,err := json.Marshal(o)
	if err != nil {
		log.Println("marshal json error:",err)
		ctx.InternalError()
	}
	ctx.SetHeader(contentType,typeJson)
	ctx.rw.Write(bs)
}

func (ctx *Context) ServeHtml(content []byte) {
	ctx.SetHeader(contentType, typeHtml )
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

func (ctx *Context) ServeByTemplate(templ string,data interface {}) {
	ctx.SetHeader("Content-Type","text/html; charset=utf-8")

	var err error
	if compress {
		ctx.SetHeader("Content-Encoding","gzip")
		w := gzip.NewWriter(ctx.rw)
		err = templates.ExecuteTemplate(w,templ,data)
		defer w.Close()
	} else {
		err = templates.ExecuteTemplate(ctx.rw,templ,data)
	}
	//log.Println("templ:",templ,"data:",data)

	if err != nil {
		//http.Error(ctx.rw, err.Error(), http.StatusInternalServerError)
		ctx.InternalError()
		log.Println(ctx.req, "Error sending response", err)
	}
}

func (ctx *Context) Ok() {
	ctx.rw.WriteHeader(http.StatusOK)
}

func (ctx *Context) InternalError() {
	ctx.Error(http.StatusInternalServerError)
}

func (ctx *Context) BadRequest() {
	ctx.Error(http.StatusBadRequest)
}

func (ctx *Context) Unauthorized() {
	ctx.Error(http.StatusUnauthorized)
}

func (ctx *Context) PageNotFound() {
	//if no custom handler
	ctx.Error(http.StatusNotFound)
}

func (ctx *Context) MethodNotAllowed() {
	ctx.Error(http.StatusMethodNotAllowed)
}

func (ctx *Context) Error(status int) {
	//ctx.rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
	ctx.rw.WriteHeader(status)
	//fmt.Fprintln(ctx.rw, error)
}


