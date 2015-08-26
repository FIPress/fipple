package fipple

import (
	"net/http"
	"log"
)

const (
	allowOrigin = "Access-Control-Allow-Origin"
	allowMethods = "Access-Control-Allow-Methods"
	allowHeaders = "Access-Control-Allow-Headers"
	oringinAllowed = "*"
	headersAllowed = "Content-Type, X-Token"
	methodsAllowed = "POST, GET, OPTIONS"

)

type Service struct {
	routes *parsedRoutes
	//req *RpcRequest
}

func NewService() *Service {
	return &Service{newParsedRoutes()}
}

func (svc *Service) Start(port string) error {
	log.Println("http listen:",port)
	return http.ListenAndServe(port,svc)
}

func (svc *Service) AddRoutes(route ...*Route) {
	svc.routes.addAll(route)
}

func (svc *Service) AddRoute(route *Route) {
	svc.routes.add(route)
}

func (svc *Service) Get(path string, action Action) {
	svc.routes.add(GetRoute(path, action))
}

func (svc *Service) Post(path string, action Action) {
	svc.routes.add(PostRoute(path, action))
}

func (svc *Service) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.RequestURI == "*" {
		if req.ProtoAtLeast(1, 1) {
			rw.Header().Set("Connection", "close")
		}
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	originControl(rw)
	if req.Method == "OPTIONS" {
		return
	}

	ctx := NewContext(rw,req)
	svc.routes.dispatch(req.URL.Path,ctx)
}

func originControl(rw http.ResponseWriter) {
	//todo: config acceptable origin
	rw.Header().Set(allowOrigin,oringinAllowed)
	rw.Header().Set(allowMethods,methodsAllowed)
	rw.Header().Set(allowHeaders,headersAllowed)
}



