package fipple


type Encoding int

const (
	None Encoding = iota
	Gzip
	Deflate
)

type HttpMethod int

const (
	NotSupported HttpMethod = iota //not http
	GET
	PUT
	POST
	DELETE
	HEAD
)

type Action func(*Context)

type Route struct {
	path string
//	encoding Encoding
	httpMethod HttpMethod
	action Action
}
/*
func Rpc(name string, execFunc func()) *Route {
	return &Route{name,Gob,None,execFunc}
}*/

func GetRoute(path string,action Action)  *Route {
	return &Route{path, GET, action}
}

func PostRoute(path string,action Action)  *Route {
	return &Route{path, POST, action}
}


/*
type RpcRoute struct {
	name string
}*/

