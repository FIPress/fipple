package fipple

import (
	"net/http"
	"net/url"
	"regexp"
	"testing"
)

func TestRegexRouter(t *testing.T) {
	router := newRegexRouter()
	router.addRoute(GetRoute("/", func(context *Context) {
		t.Log("home")
	}))

	router.addRoute(GetRoute("/a/:id", func(context *Context) {
		id := context.GetStringParam("id")
		t.Log("get id:", id)
	}))

	router.addRoute(GetRoute("/:type/:id", func(context *Context) {
		id := context.GetStringParam("id")
		t.Log("get id:", id)
	}))

	router.addRoute(GetRoute("/book/:id/:version", func(context *Context) {
		id := context.GetStringParam("id")
		t.Log("get id:", id)
	}))

	fakeReq := &http.Request{Method: "GET", URL: &url.URL{}}
	ctx := NewContext(FakeResponseWriter{}, fakeReq)
	//router.dispatch("/",ctx)
	//router.dispatch("/a/b",ctx)
	//router.dispatch("/a/b/c",ctx)
	router.dispatch("/book/test-0828", ctx)
}

func TestRegex(t *testing.T) {
	regex := regexp.MustCompile(`^/([^/]+)/([^/]+)$`)
	matches := regex.FindAllStringSubmatch("/a/b", -1)
	if matches != nil && len(matches) == 1 {
		t.Log(matches)
	}
}
