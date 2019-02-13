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

/*

type routeAction struct {
	//encode Encode
	httpMethod HttpMethod
	action     action
}

func newAction(route *route) *routeAction {
	return &routeAction{route.httpMethod, route.action}
}
*/

type routePart interface {
	isNil() bool
	getAction() action
	getParamNames() []string
	getOrNewSub() *parsedRoutes
	getSub() *parsedRoutes
}

type dynamicPart struct {
	*staticPart
	//para string
	//typ  routePartPattern
	regex *regexp.Regexp
}

type staticPart struct {
	action     action
	paramNames []string
	//action action
	//httpMethod HttpMethod
	sub *parsedRoutes
}

func (sp *staticPart) isNil() bool {
	return sp == nil
}

/*func (sp *staticPart) setAction(a action) {
	if sp != nil {
		sp.action = a
	}
}*/

func (sp *staticPart) getOrNewSub() *parsedRoutes {
	if sp == nil {
		return nil
	}

	if sp.sub == nil {
		sp.sub = newParsedRoutes()
	}
	return sp.sub
}

func (sp *staticPart) getAction() action {
	if sp == nil {
		return nil
	}
	return sp.action
}

func (sp *staticPart) getParamNames() []string {
	if sp == nil {
		return nil
	}
	return sp.paramNames
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

func (pr *parsedRoutes) getPart(key string) (routePart, string) {
	pattern, para, regex := getParaAndPattern(key)
	if pattern == static {
		part := pr.staticDict[key]
		if part == nil {
			part = &staticPart{}
			pr.staticDict[key] = part
		}
		return part, para
	} else {
		part := pr.dynamicDict[pattern]
		if part == nil {
			s := new(staticPart)
			part = &dynamicPart{s, regex}
			pr.dynamicDict[pattern] = part
		}
		return part, para
	}
}

func (pr *parsedRoutes) add(route *route) {
	if pr == nil {
		return
	}
	parse(route, pr)
}

func (pr *parsedRoutes) addAll(routeList []*route) {
	if pr == nil {
		return
	}
	for _, route := range routeList {
		parse(route, pr)
	}
}

func parseAll(routeList []*route) *parsedRoutes {
	parsed := newParsedRoutes()
	for _, route := range routeList {
		parse(route, parsed)
	}
	return parsed
}

func parse(route *route, parsed *parsedRoutes) {
	key, rest := splitPart(route.path)
	getLogger().Info("key:", key, "rest:", rest)
	part, para := parsed.getPart(key)
	if para != "" {
		route.addPara(para)
	}
	sub := part.getOrNewSub()
	if len(rest) == 0 {
		sub.staticDict[string(route.httpMethod)] = &staticPart{route.action, route.paras, nil}
	} else {
		route.path = rest
		parse(route, sub)
	}
}

func (pr *parsedRoutes) dispatch(url string, ctx *Context) {
	getLogger().Debug("dispatch, url", url)
	var part routePart
	key, rest := splitPart(url)
	part = pr.staticDict[key]
	//log.Println("dispatch, key:",key,",rest:",rest,",part:",part)
	if part.isNil() {
		for typ, d := range pr.dynamicDict {
			switch typ {
			case param:
				//getLogger().Debug("param part:",d.para,",val:",key)
				//ctx.params[d.para] = key
				ctx.addParam(key)
				part = d
			case splat:
				//log.Println("splat part:",d.para,"val:",url)
				//ctx.params[d.para] = url
				ctx.addParam(key)
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
		getLogger().Debug("rest:", rest)

		sub := part.getSub()
		if sub == nil || len(sub.staticDict) == 0 {
			getLogger().Error("404 - action not found")
			ctx.PageNotFound()
			return
		}

		part = sub.staticDict[ctx.req.Method]
		action := part.getAction()

		if part == nil || action == nil {
			getLogger().Debug("req.Method:", ctx.req.Method, "not allowed")
			ctx.MethodNotAllowed()
			return
		}
		ctx.buildNamedParams(part.getParamNames())
		action(ctx)
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

/*func getMethod(ms string) (method HttpMethod) {
	switch ms {
	case "GET":
		method = GET
	case "POST":
		method = POST
	case "PUT":
		method = PUT
	case "DELETE":
		method = DELETE
	default:
		method = NotSupported
	}
	return
}*/

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
		log.Println("dynamic route, type:", typ, ",regex:", p.regex)
		printActionAndSub(p)
		log.Println("end of dynamic route:", typ)
	}

}

func printActionAndSub(part routePart) {
	if part.getAction() != nil {
		log.Println("action:", part.getAction())
	}
	if part.getSub() != nil {
		part.getSub().print()
	}
	if part.getParamNames() != nil {
		log.Println(part.getParamNames())
	}
}
