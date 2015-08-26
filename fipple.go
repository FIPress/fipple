package fipple

import "log"

import "text/template"

import "net/http"


import (
)

var (
	defaultService *Service
	templates *template.Template
	compress = true
	//todo: other config
)

func DefaultService() *Service {
	if defaultService == nil {
		defaultService = NewService()
	}
	return defaultService
}

func Start(port string) error {
	return DefaultService().Start(port)
}

func AddRoutes(route ...*Route) {
	DefaultService().AddRoutes(route...)
}

func AddRoute(route *Route) {
	DefaultService().AddRoute(route)
}

func Get(path string, action Action) {
	DefaultService().Get(path,action)
}

func Post(path string, action Action) {
	DefaultService().Post(path,action)
}

func ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	DefaultService().ServeHTTP(rw,req)
}

func AddTemplGlob(pattern string) {
	var err error
	templates,err = template.ParseGlob(pattern)
	if err != nil {
		log.Println("parse templates failed,",err)
	}
	//templates.ParseFiles("/temps/cover.html")
}
