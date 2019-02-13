package fipple

import "text/template"

import (
	"github.com/fipress/fiplog"
	"net/http"
)

var (
	defaultService *Service
	templates      *template.Template
	compress       = true
	logger         fiplog.Logger
	//todo: other config
)

type RouterType int

const (
	Regex RouterType = iota
	Trie
)

type Router interface {
	addRoute(route *route)
	dispatch(url string, ctx *Context)
}

func DefaultService() *Service {
	if defaultService == nil {
		defaultService = NewService()
	}
	return defaultService
}

func Start(port string) error {
	return DefaultService().Start(port)
}

func AddRoutes(route ...*route) {
	DefaultService().AddRoutes(route...)
}

func AddRoute(route *route) {
	DefaultService().AddRoute(route)
}

func Get(path string, action action) {
	DefaultService().Get(path, action)
}

func Post(path string, action action) {
	DefaultService().Post(path, action)
}

func Delete(path string, action action) {
	getLogger().Debug("delete, path:", path)
	DefaultService().Delete(path, action)
}

func ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	DefaultService().ServeHTTP(rw, req)
}

func AddTemplGlob(pattern string) {
	var err error
	templates, err = template.ParseGlob(pattern)
	if err != nil {
		getLogger().Error("parse templates failed,", err)
	}
	//templates.ParseFiles("/temps/cover.html")
}

func SetLogger(log fiplog.Logger) {
	logger = log
}

func getLogger() fiplog.Logger {
	if logger == nil {
		logger = fiplog.GetLogger()
	}
	return logger
}
