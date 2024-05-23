// Copyright 2014 Julien Schmidt. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"runtime"

	"github.com/aerogo/aero"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"github.com/gin-gonic/gin"
	"github.com/gocraft/web"
	"github.com/julienschmidt/httprouter"
)

type route struct {
	method string
	path   string
}

type mockResponseWriter struct{}

func (m *mockResponseWriter) Header() (h http.Header) {
	return http.Header{}
}

func (m *mockResponseWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (m *mockResponseWriter) WriteString(s string) (n int, err error) {
	return len(s), nil
}

func (m *mockResponseWriter) WriteHeader(int) {}

var nullLogger *log.Logger

// flag indicating if the normal or the test handler should be loaded
var loadTestHandler = false

func init() {
	// beego sets it to runtime.NumCPU()
	// Currently none of the contesters does concurrent routing
	runtime.GOMAXPROCS(1)

	// makes logging 'webscale' (ignores them)
	log.SetOutput(new(mockResponseWriter))
	nullLogger = log.New(new(mockResponseWriter), "", 0)

	initBeego()
	initGin()
}

// Common
func httpHandlerFunc(_ http.ResponseWriter, _ *http.Request) {}

func httpHandlerFuncTest(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, r.RequestURI)
}

// Aero
func aeroHandler(c aero.Context) error {
	return nil
}

func aeroHandlerWrite(ctx aero.Context) error {
	io.WriteString(ctx.Response().Internal(), ctx.Get("name"))
	return nil
}
func aeroHandlerTest(ctx aero.Context) error {
	io.WriteString(ctx.Response().Internal(), ctx.Request().Path())
	return nil
}
func loadAero(routes []route) http.Handler {
	var h aero.Handler = aeroHandler
	if loadTestHandler {
		h = aeroHandlerTest
	}
	app := aero.New()
	for _, r := range routes {
		switch r.method {
		case "GET":
			app.Get(r.path, h)
		case "POST":
			app.Post(r.path, h)
		case "PUT":
			app.Put(r.path, h)
		case "PATCH":
			app.Router().Add(http.MethodPatch, r.path, h)
		case "DELETE":
			app.Delete(r.path, h)
		default:
			panic("Unknow HTTP method: " + r.method)
		}
	}
	return app
}
func loadAeroSingle(method, path string, h aero.Handler) http.Handler {
	app := aero.New()
	switch method {
	case "GET":
		app.Get(path, h)
	case "POST":
		app.Post(path, h)
	case "PUT":
		app.Put(path, h)
	case "PATCH":
		app.Router().Add(http.MethodPatch, path, h)
	case "DELETE":
		app.Delete(path, h)
	default:
		panic("Unknow HTTP method: " + method)
	}
	// }
	return app
}

// beego
func beegoHandler(ctx *context.Context) {}

func beegoHandlerWrite(ctx *context.Context) {
	ctx.WriteString(ctx.Input.Param(":name"))
}

func beegoHandlerTest(ctx *context.Context) {
	ctx.WriteString(ctx.Request.RequestURI)
}

func initBeego() {
	beego.BConfig.RunMode = beego.PROD
	beego.BeeLogger.Close()
}

func loadBeego(routes []route) http.Handler {
	h := beegoHandler
	if loadTestHandler {
		h = beegoHandlerTest
	}

	re := regexp.MustCompile(":([^/]*)")
	app := beego.NewControllerRegister()
	for _, route := range routes {
		route.path = re.ReplaceAllString(route.path, ":$1")
		switch route.method {
		case "GET":
			app.Get(route.path, h)
		case "POST":
			app.Post(route.path, h)
		case "PUT":
			app.Put(route.path, h)
		case "PATCH":
			app.Patch(route.path, h)
		case "DELETE":
			app.Delete(route.path, h)
		default:
			panic("Unknow HTTP method: " + route.method)
		}
	}
	return app
}

func loadBeegoSingle(method, path string, handler beego.FilterFunc) http.Handler {
	app := beego.NewControllerRegister()
	switch method {
	case "GET":
		app.Get(path, handler)
	case "POST":
		app.Post(path, handler)
	case "PUT":
		app.Put(path, handler)
	case "PATCH":
		app.Patch(path, handler)
	case "DELETE":
		app.Delete(path, handler)
	default:
		panic("Unknow HTTP method: " + method)
	}
	return app
}

// Gin
func ginHandle(_ *gin.Context) {}

func ginHandleWrite(c *gin.Context) {
	io.WriteString(c.Writer, c.Params.ByName("name"))
}

func ginHandleTest(c *gin.Context) {
	io.WriteString(c.Writer, c.Request.RequestURI)
}

func initGin() {
	gin.SetMode(gin.ReleaseMode)
}

func loadGin(routes []route) http.Handler {
	h := ginHandle
	if loadTestHandler {
		h = ginHandleTest
	}

	router := gin.New()
	for _, route := range routes {
		router.Handle(route.method, route.path, h)
	}
	return router
}

func loadGinSingle(method, path string, handle gin.HandlerFunc) http.Handler {
	router := gin.New()
	router.Handle(method, path, handle)
	return router
}

// gocraft/web
type gocraftWebContext struct{}

func gocraftWebHandler(w web.ResponseWriter, r *web.Request) {}

func gocraftWebHandlerWrite(w web.ResponseWriter, r *web.Request) {
	io.WriteString(w, r.PathParams["name"])
}

func gocraftWebHandlerTest(w web.ResponseWriter, r *web.Request) {
	io.WriteString(w, r.RequestURI)
}

func loadGocraftWeb(routes []route) http.Handler {
	h := gocraftWebHandler
	if loadTestHandler {
		h = gocraftWebHandlerTest
	}

	router := web.New(gocraftWebContext{})
	for _, route := range routes {
		switch route.method {
		case "GET":
			router.Get(route.path, h)
		case "POST":
			router.Post(route.path, h)
		case "PUT":
			router.Put(route.path, h)
		case "PATCH":
			router.Patch(route.path, h)
		case "DELETE":
			router.Delete(route.path, h)
		default:
			panic("Unknow HTTP method: " + route.method)
		}
	}
	return router
}

func loadGocraftWebSingle(method, path string, handler interface{}) http.Handler {
	router := web.New(gocraftWebContext{})
	switch method {
	case "GET":
		router.Get(path, handler)
	case "POST":
		router.Post(path, handler)
	case "PUT":
		router.Put(path, handler)
	case "PATCH":
		router.Patch(path, handler)
	case "DELETE":
		router.Delete(path, handler)
	default:
		panic("Unknow HTTP method: " + method)
	}
	return router
}

// HttpRouter
func httpRouterHandle(_ http.ResponseWriter, _ *http.Request, _ httprouter.Params) {}

func httpRouterHandleWrite(w http.ResponseWriter, _ *http.Request, ps httprouter.Params) {
	io.WriteString(w, ps.ByName("name"))
}

func httpRouterHandleTest(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	io.WriteString(w, r.RequestURI)
}

func loadHttpRouter(routes []route) http.Handler {
	h := httpRouterHandle
	if loadTestHandler {
		h = httpRouterHandleTest
	}

	router := httprouter.New()
	for _, route := range routes {
		router.Handle(route.method, route.path, h)
	}
	return router
}

func loadHttpRouterSingle(method, path string, handle httprouter.Handle) http.Handler {
	router := httprouter.New()
	router.Handle(method, path, handle)
	return router
}

// Usage notice
func main() {
	fmt.Println("Usage: go test -bench=. -timeout=20m")
	os.Exit(1)
}
