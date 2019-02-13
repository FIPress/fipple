package fipple

type Encoding int

const (
	None Encoding = iota
	Gzip
	Deflate
)

type HttpMethod string

const (
	NotSupported HttpMethod = "" //not http
	GET          HttpMethod = "GET"
	PUT          HttpMethod = "PUT"
	POST         HttpMethod = "POST"
	DELETE       HttpMethod = "DELETE"
	HEAD                    = "HEAD"
)

type action func(*Context)

type route struct {
	path string
	//	encoding Encoding
	httpMethod HttpMethod
	action     action
	paras      []string //named parameters
}

func (r *route) addPara(para string) {
	if r.paras == nil {
		r.paras = make([]string, 1)
		r.paras[0] = para
	} else {
		r.paras = append(r.paras, para)
	}
}

/*
func Rpc(name string, execFunc func()) *route {
	return &route{name,Gob,None,execFunc}
}*/

func GetRoute(path string, action action) *route {
	return &route{path, GET, action, nil}
}

func PostRoute(path string, action action) *route {
	return &route{path, POST, action, nil}
}

func PutRoute(path string, action action) *route {
	return &route{path, PUT, action, nil}
}

func DeleteRoute(path string, action action) *route {
	return &route{path, DELETE, action, nil}
}

/*
type RpcRoute struct {
	name string
}*/
