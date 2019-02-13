package fipple

import (
	"net/http"
)

const (
	allowOrigin    = "Access-Control-Allow-Origin"
	allowMethods   = "Access-Control-Allow-Methods"
	allowHeaders   = "Access-Control-Allow-Headers"
	oringinAllowed = "*"
	headersAllowed = "Content-Type, X-Token"
	methodsAllowed = "POST, GET, OPTIONS"
)

type Service struct {
	//routes *parsedRoutes
	router Router
	//req *RpcRequest
}

func NewService() *Service {
	return NewServiceWithRouterType(Regex)
}

func NewServiceWithRouterType(rType RouterType) *Service {
	r := newRegexRouter()
	//if rType == Regex {}
	return &Service{r}
}

func (svc *Service) Start(port string) error {
	getLogger().Info("http listen:", port)
	return http.ListenAndServe(port, svc)
}

func (svc *Service) AddRoutes(route ...*route) {
	//svc.routes.addAll(route)
}

func (svc *Service) AddRoute(route *route) {
	//svc.routes.add(route)
	svc.router.addRoute(route)
}

func (svc *Service) Get(path string, action action) {
	svc.router.addRoute(GetRoute(path, action))
}

func (svc *Service) Post(path string, action action) {
	svc.router.addRoute(PostRoute(path, action))
}

func (svc *Service) Delete(path string, action action) {
	getLogger().Debug("add path:", path)
	svc.router.addRoute(DeleteRoute(path, action))
}

func (svc *Service) FileServer() {
	//todo: add file server
	//http.Handle("/static/",http.StripPrefix("/static/",http.FileServer(http.Dir("./static/"))))
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

	ctx := NewContext(rw, req)
	svc.router.dispatch(req.URL.Path, ctx)
}

func originControl(rw http.ResponseWriter) {
	//todo: config acceptable origin
	rw.Header().Set(allowOrigin, oringinAllowed)
	rw.Header().Set(allowMethods, methodsAllowed)
	rw.Header().Set(allowHeaders, headersAllowed)
}
