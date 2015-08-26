package fipple

import (
	"log"
	"regexp"
)

var (
	paramEx = regexp.MustCompile(`[^/]+`) //match a single param part, start with ':', e.g. /user/:id
	splatEx = regexp.MustCompile(`.+`)    //a part match the rest of the url, start with '*', e.g. /user/*name
	//dynamic part(s) match the provided regular express, start with '$', e.g. /user/$id<[0-9]+>
	customEx = regexp.MustCompile(`\$(.+)<(.+)>`) //custom regexp extractor,
)

//Url pattern indicate the parse type of a part of url
type routePartPattern uint

const (
	static routePartPattern = iota //a single dynamic part, start with ':', it's pattern is [^/]+, e.g.  /user/:id
	param
	splat   //a dynamic part match the rest of the url, start with '*', it's patter is .+, e.g. /user/*name
	custom  //dynamic part(s) match the provided regular express, start with '$', e.g. /user/$id<[0-9]+>
	discard //discard this part, represented by '*', e.g. /user/*/:id, discard the second par
	//todo: optional
)

type routeAction struct {
	//encode Encode
	httpMethod HttpMethod
	action     Action
}

func newAction(route *Route) *routeAction {
	return &routeAction{route.httpMethod, route.action}
}

type routePart interface {
	isNil() bool
	setAction(*routeAction)
	getAction() *routeAction
	getOrNewSub() *parsedRoutes
	getSub() *parsedRoutes
}

type dynamicPart struct {
	*staticPart
	para string
	//typ  routePartPattern
	regex *regexp.Regexp
}

type staticPart struct {
	action *routeAction
	sub    *parsedRoutes
}

func (sp *staticPart) isNil() bool {
	return sp == nil
}

func (sp *staticPart) setAction(a *routeAction) {
	if sp != nil {
		sp.action = a
	}
}

func (sp *staticPart) getOrNewSub() *parsedRoutes {
	if sp == nil {
		return nil
	}

	if sp.sub == nil {
		sp.sub = newParsedRoutes()
	}
	return sp.sub
}

func (sp *staticPart) getAction() *routeAction {
	if sp == nil {
		return nil
	}
	return sp.action
}

func (sp *staticPart) getSub() *parsedRoutes {
	if sp == nil {
		return nil
	}

	return sp.sub
}

/*
func newPart(key string) *routePart {

	return &routePart{pattern,nil,nil}
}*/

/*func (rp *routePart) matchCustomSub() bool {
	for _,srp := range rp.sub.dict {
		if srp.pattern == custom {
			return true
		}
	}
	return false
}*/

type parsedRoutes struct {
	staticDict  map[string]*staticPart
	dynamicDict map[routePartPattern]*dynamicPart
}

func newParsedRoutes() *parsedRoutes {
	return &parsedRoutes{make(map[string]*staticPart), make(map[routePartPattern]*dynamicPart)}
}

func (pr *parsedRoutes) getPart(key string) routePart {
	pattern, para, regex := getParaAndPattern(key)
	if pattern == static {
		part := pr.staticDict[key]
		if part == nil {
			part = &staticPart{nil, nil}
			pr.staticDict[key] = part
		}
		return part
	} else {
		part := pr.dynamicDict[pattern]
		if part == nil {
			s := &staticPart{nil, nil}
			part = &dynamicPart{s, para, regex}
			pr.dynamicDict[pattern] = part
		}
		return part
	}
}

func (pr *parsedRoutes) add(route *Route) {
	if pr == nil {
		return
	}
	parse(route.path, newAction(route), pr)
}

func (pr *parsedRoutes) addAll(routeList []*Route) {
	if pr == nil {
		return
	}
	for _, route := range routeList {
		parse(route.path, newAction(route), pr)
	}
}

func parseAll(routeList []*Route) *parsedRoutes {
	parsed := newParsedRoutes()
	for _, route := range routeList {
		action := newAction(route)
		parse(route.path, action, parsed)
	}
	return parsed
}

func parse(path string, action *routeAction, parsed *parsedRoutes) {
	key, rest := splitPart(path)
	part := parsed.getPart(key)
	if len(rest) == 0 {
		//log.Println("before set action,parsed.static:",parsed.staticDict,"parsed.dynamic",parsed.dynamicDict)
		part.setAction(action)
		//parsed.addAction(key,action)
		//log.Println("after set action,parsed.static:",parsed.staticDict,"parsed.dynamic",parsed.dynamicDict)
	} else {
		sub := part.getOrNewSub()
		parse(rest, action, sub)
		//log.Println("len of sub static:",len(sub.staticDict),"len of sub dynamic",len(sub.dynamicDict))
	}
}

func (pr *parsedRoutes) dispatch(url string, ctx *Context) {
	//log.Println("dispatch...")
	var part routePart
	key, rest := splitPart(url)
	part = pr.staticDict[key]
	//log.Println("dispatch, key:",key,",rest:",rest,",part:",part)
	if part.isNil() {
		for typ, d := range pr.dynamicDict {
			switch typ {
			case param:
				//log.Println("param part:",d.para,",val:",key)
				ctx.params[d.para] = key
				part = d
			case splat:
				//log.Println("splat part:",d.para,"val:",url)
				ctx.params[d.para] = url
				part = d
				rest = ""
			case discard:
				//log.Println("discard part")
			case custom:
				//log.Println("todo: custom,should match custom regex")
			}
		}
	}

	if part.isNil() {
		//log.Println("404 - path not found")
		ctx.PageNotFound()
		return
	}

	if len(rest) == 0 {
		//log.Println("part:",part)
		ra := part.getAction()
		//log.Println("get action:",ra)
		if ra != nil {
			if getMethod(ctx.req.Method) != ra.httpMethod {
				log.Println("req.Method:", ctx.req.Method, ",ra.httpMethod:", ra.httpMethod, ",ra:", ra)
				ctx.MethodNotAllowed()
			} else {
				//log.Println("execute action:")
				ra.action(ctx)
			}
		} else {
			log.Println("404 - action not found")
			ctx.PageNotFound()
		}
	} else {
		part.getSub().dispatch(rest, ctx)
	}
	return
}

/*
func (pr *parsedRoutes) addAction(key string,action *action) {
	part,ok := pr.dict[key]
	//log.Println("add action,before,key:",key,",dict:",pr.dict,",part,ok:",part,ok)
	if ok {
		part.action = action
	} else {
		part = newPart(key)
		part.action = action
		pr.dict[key] = part
	}
	//log.Println("added action, after,action:",part.action,",dict:",pr.dict)
}


func (pr *parsedRoutes) getSub(key string) (sub *parsedRoutes) {
	part,ok := pr.dict[key]
	//log.Println("part",part,",ok:",ok)
	if !ok || part == nil {
		part = newPart(key)
		//log.Println("new part,key:",key,",part:",part)
		pr.dict[key] = part
	}

	sub = part.sub

	if sub == nil {
		//log.Println("new sub")
		sub = newParsedRoutes()
		part.sub = sub
	}

	//log.Println("Added part,key:",key,",key.part:",pr.dict[key],"key.part.sub:",pr.dict[key].sub)
	return
}*/
/*
func getUrlPattern(key string) urlPattern {
	switch key[0] {
	case ':':
		return dynamic
	case '*':
		return splat
	case '$':
		return custom
	default:
		return static
	}
}*/

func getParaAndPattern(key string) (pattern routePartPattern, para string, regex *regexp.Regexp) {
	switch key[0] {
	case ':':
		para = string(key[1:])
		pattern = param
	case '*':
		if len(key) == 1 {
			pattern = discard
		} else {
			para = string(key[1:])
			pattern = splat
		}
	case '$':
		pattern = custom
		r := customEx.FindStringSubmatch(key)
		para = r[1]
		regex = regexp.MustCompile(r[2])
	default:
		pattern = static
	}
	return
}

func splitPart(path string) (part, rest string) {
	if len(path) != 0 {
		start := 0
		if path[0] == '/' {
			start = 1
		}
		i := start

		for i < len(path) && path[i] != '/' {
			i++
		}
		part = path[start:i]
		if i < len(path)-1 {
			rest = path[i:]
		}
	}

	if len(part) == 0 {
		part = "/"
	}
	return
}

func getMethod(ms string) (method HttpMethod) {
	switch ms {
	case "GET":
		method = GET
	case "POST":
		method = POST
	default:
		method = NotSupported
	}
	return
}

//for test
func (pr *parsedRoutes) print() {
	log.Println("static len:", len(pr.staticDict))
	for key := range pr.staticDict {
		log.Println("static route:", key)
		rp := pr.staticDict[key]
		printActionAndSub(rp)
		log.Println("end of static route:", key)
	}

	log.Println("dynamic len:", len(pr.dynamicDict))
	for typ, p := range pr.dynamicDict {
		log.Println("dynamic route, type:", typ, ",para:", p.para, ",regex:", p.regex)
		printActionAndSub(p)
		log.Println("end of dynamic route:", p.para)
	}
}

func printActionAndSub(part routePart) {
	if part.getAction() != nil {
		log.Println("Action:", part.getAction())
	}
	if part.getSub() != nil {
		part.getSub().print()
	}
}
