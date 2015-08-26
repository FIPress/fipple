package fipple

import (
	"log"
	"net/http"
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
	log.Println("get user")
	log.Println("userid:", ctx.GetParam("id"))
}

func getUserEx(ctx *Context) {
	log.Println("get userEx")
	log.Println("userid:", ctx.GetParam("id"))
	log.Println("user name:", ctx.GetParam("name"))
}

func addUser(ctx *Context) {

}

func TestParse(t *testing.T) {
	routes := []*Route{GetRoute("/", home),
		GetRoute("/user/:id/", getUser),
		GetRoute("/user/:id/:name", getUserEx),
		PostRoute("/user/add", addUser),
	}
	parsed := parseAll(routes)
	//parsed.print()
	fakeReq := &http.Request{Method: "GET"}
	ctx := NewContext(FakeResponseWriter{}, fakeReq)
	parsed.dispatch("/", ctx)
	parsed.dispatch("/user/123", ctx)
	parsed.dispatch("/user/23/tom", ctx)
}

func TestParseSplat(t *testing.T) {
	routes := []*Route{
		GetRoute("/:id/static/*path", getStatic),
	}
	parsed := parseAll(routes)
	//parsed.print()
	fakeReq := &http.Request{Method: "GET"}
	ctx := NewContext(FakeResponseWriter{}, fakeReq)
	parsed.dispatch("/a/static/b/c/d", ctx)
}

func getStatic(ctx *Context) {
	log.Println("id:", ctx.GetStringParam("id"))
	log.Println("path:", ctx.GetStringParam("path"))
}

/*func TestParseCustom(t *testing.T) {
	routes := []*Route{GetRoute("/",home),
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
