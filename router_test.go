package fipple

import (
	"log"
	"net/http"
	"net/url"
	"testing"
)

func TestSplitPart(t *testing.T) {
	path := "/user/all"
	part, rest := splitPart(path)
	if part != "user" || rest != "/all" {
		t.Error("part:", part, ",rest:", rest)
	}

	path = "/user"
	part, rest = splitPart(path)
	if part != "user" || len(rest) != 0 {
		t.Error("part:", part, ",rest:", rest)
	}

	path = "/user/"
	part, rest = splitPart(path)
	if part != "user" || len(rest) != 0 {
		t.Error("part:", part, ",rest:", rest)
	}
}

func home(ctx *Context) {
	log.Println("home page")
}

func getAll(ctx *Context) {
	log.Println("get all", "Abby", "Tony")
}

func getUser(ctx *Context) {
	log.Println("get user,userid:", ctx.GetParam("id"))
}

func putUser(ctx *Context) {
	log.Println("put user, user id:", ctx.GetParam("uid"))
}

func getUserEx(ctx *Context) {
	log.Println("get userEx")
	log.Println("userid:", ctx.GetParam("id"))
	log.Println("user name:", ctx.GetParam("name"))
	log.Println("user email:", ctx.GetParam("email"))
}

func addUser(ctx *Context) {

}

func TestParse(t *testing.T) {
	routes := []*route{GetRoute("/", home),
		GetRoute("/user/:id", getUser),
		GetRoute("/user/:id/:name(/:email)", getUserEx), //todo: optional route
		PostRoute("/user/add", addUser),
		PutRoute("/user/:uid", putUser),
	}
	parsed := parseAll(routes)
	parsed.print()
	fakeReq := &http.Request{Method: "GET", URL: &url.URL{}}
	ctx := NewContext(FakeResponseWriter{}, fakeReq)
	parsed.dispatch("/", ctx)

	ctx = NewContext(FakeResponseWriter{}, fakeReq)
	parsed.dispatch("/user/123", ctx)

	ctx = NewContext(FakeResponseWriter{}, fakeReq)
	parsed.dispatch("/user/23/tom", ctx)

	ctx = NewContext(FakeResponseWriter{}, fakeReq)
	parsed.dispatch("/user/234/sam/abcd", ctx)

	fakeReq = &http.Request{Method: "PUT", URL: &url.URL{}}
	ctx = NewContext(FakeResponseWriter{}, fakeReq)
	parsed.dispatch("/user/123", ctx)
}

func TestParseSplat(t *testing.T) {
	routes := []*route{
		GetRoute("/:id/static/*path", getStatic),
	}
	parsed := parseAll(routes)
	//parsed.print()
	fakeReq := &http.Request{Method: "GET", URL: &url.URL{}}
	ctx := NewContext(FakeResponseWriter{}, fakeReq)
	parsed.dispatch("/a/static/b/c/d", ctx)
}

func getStatic(ctx *Context) {
	log.Println("id:", ctx.GetStringParam("id"))
	log.Println("path:", ctx.GetStringParam("path"))
}

/*func TestParseCustom(t *testing.T) {
	routes := []*route{GetRoute("/",home),
		GetRoute("/page/$lang<>",getUser),
	}
	parsed := parseAll(routes)
	//parsed.print()
	fakeReq := &http.Request{Method:"GET"}
	ctx := NewContext(FakeResponseWriter{},fakeReq)
	parsed.dispatch("/",ctx)
	parsed.dispatch("/user/123",ctx)
	parsed.dispatch("/user/23/tom",ctx)
}*/

type FakeResponseWriter struct {
}

// Header returns the header map that will be sent by WriteHeader.
func (fw FakeResponseWriter) Header() http.Header {
	return nil
}

func (fw FakeResponseWriter) Write([]byte) (int, error) {
	return 0, nil
}

func (fw FakeResponseWriter) WriteHeader(int) {
}
