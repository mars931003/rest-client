package rest_client

import (
	"github.com/emicklei/go-restful"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type HttpMethod int

const (
	Get HttpMethod = iota
	Post
	Put
	Delete
)

type Registrar struct{}

var WebContainer *restful.Container
var cache = make(map[string]*restful.WebService)

func init() {
	container := restful.NewContainer()
	// 跨域过滤器
	cors := restful.CrossOriginResourceSharing{
		ExposeHeaders:  []string{"X-My-Header"},
		AllowedHeaders: []string{"Content-Type", "Accept"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
		CookiesAllowed: false,
		Container:      container}
	container.Filter(cors.Filter)
	// Add container filter to respond to OPTIONS
	container.Filter(container.OPTIONSFilter)
	WebContainer = container
}

func (r *Registrar) RegisterRoute(path string, method HttpMethod, function func(request *restful.Request, response *restful.Response), queryParam ...string) {
	rootPath := getRootPath(path)
	subPath := getSubPath(path)
	ws, ok := cache[rootPath]
	if !ok {
		ws = new(restful.WebService)
		ws.Path(rootPath).Consumes("application/json").Produces("application/json")
		cache[rootPath] = ws
	} else {
		ws = cache[rootPath]
	}
	var routeBuilder *restful.RouteBuilder
	switch method {
	case Get:
		routeBuilder = ws.GET(subPath).To(function).Doc("")
		for _, v := range queryParam {
			routeBuilder.Param(ws.QueryParameter(v, "").DataType("string"))
		}
		break
	case Post:
		routeBuilder = ws.POST(subPath).To(function).Doc("")
		break
	case Put:
		routeBuilder = ws.PUT(subPath).To(function).Doc("")
		break
	case Delete:
		routeBuilder = ws.DELETE(subPath).To(function).Doc("")
		break
	}
	routeBuilder.Do(requestSuccessful, requestFailed)
	ws.Route(routeBuilder)
}

func getRootPath(path string) string {
	index := strings.Index(path[1:], "/")
	if index < 0 {
		return path
	}
	return path[:index+1]
}

func getSubPath(path string) string {
	index := strings.Index(path[1:], "/")
	return path[index+1:]
}

func ApplicationRun(port int) {
	for _, v := range cache {
		WebContainer.Add(v)
	}
	log.Printf("********* start listening on localhost:%d *********", port)
	server := &http.Server{Addr: ":" + strconv.Itoa(port), Handler: WebContainer}
	defer func(server *http.Server) {
		err := server.Close()
		if err != nil {
			log.Panicf("server close exception : %v ", err)
		}
	}(server)
	log.Fatal(server.ListenAndServe())
}

func requestSuccessful(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", "success")
}

func requestFailed(b *restful.RouteBuilder) {
	b.Returns(http.StatusInternalServerError, "Bummer, something went wrong", nil)
}
