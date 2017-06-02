package main

import (
	"errors"
	"fmt"
	"strconv"

	"time"

	"strings"

	"encoding/json"

	"github.com/Jeffail/gabs"
	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
)

var (
	apiToken   string
	apiTimeout = time.Second * 30
)

func apiMessageAdd(ctx *fasthttp.RequestCtx) {
	if !checkRequest(ctx) {
		tokenError(ctx)
		return
	}
	fmt.Fprintf(ctx, "hello, %s!\n", ctx.UserValue("name"))
}
func apiMessageUpdate(ctx *fasthttp.RequestCtx) {
	if !checkRequest(ctx) {
		tokenError(ctx)
		return
	}
	fmt.Fprintf(ctx, "hello, %s!\n", ctx.UserValue("name"))
}
func apiMessageDelete(ctx *fasthttp.RequestCtx) {
	if !checkRequest(ctx) {
		tokenError(ctx)
		return
	}
	fmt.Fprintf(ctx, "hello, %s!\n", ctx.UserValue("name"))
}
func apiMessageStatus(ctx *fasthttp.RequestCtx) {
	if !checkRequest(ctx) {
		tokenError(ctx)
		return
	}
	fmt.Fprintf(ctx, "hello, %s!\n", ctx.UserValue("name"))
}
func apiConsumerAdd(ctx *fasthttp.RequestCtx) {
	if !checkRequest(ctx) {
		tokenError(ctx)
		return
	}
	fmt.Fprintf(ctx, "hello, %s!\n", ctx.UserValue("name"))
}
func apiConsumerUpdate(ctx *fasthttp.RequestCtx) {
	if !checkRequest(ctx) {
		tokenError(ctx)
		return
	}
	fmt.Fprintf(ctx, "hello, %s!\n", ctx.UserValue("name"))
}
func apiConsumerDelete(ctx *fasthttp.RequestCtx) {
	if !checkRequest(ctx) {
		tokenError(ctx)
		return
	}
	fmt.Fprintf(ctx, "hello, %s!\n", ctx.UserValue("name"))
}
func apiConsumerStatus(ctx *fasthttp.RequestCtx) {
	if !checkRequest(ctx) {
		tokenError(ctx)
		return
	}
	fmt.Fprintf(ctx, "hello, %s!\n", ctx.UserValue("name"))
}
func apiPublish(ctx *fasthttp.RequestCtx) {
	queryString := string(ctx.QueryArgs().QueryString())
	exchangeName := ctx.UserValue("name").(string)
	msg, _, err := getMessage(exchangeName)
	if err != nil {
		response(ctx, "", err)
		return
	}

	tokenB := ctx.Request.Header.Peek("Token")
	token := string(tokenB)
	if token != msg.Token {
		response(ctx, "", errors.New("token error"))
		return
	}
	routeKeyB := ctx.Request.Header.Peek("RouteKey")
	routeKey := string(routeKeyB)
	method := strings.ToLower(string(ctx.Request.Header.Method()))
	headerMap := make(map[string]string)
	igstr := "Token,Route-Key,Host,Accept-Encoding,Content-Length,Content-Type,Connection"
	ignores := strings.Split(igstr, ",")
	ctx.Request.Header.VisitAll(func(k, v []byte) {
		if in, _ := inArray(string(k), ignores); !in {
			headerMap[string(k)] = string(v)
		}
	})
	mqMessage := gabs.New()
	a, _ := json.Marshal(headerMap)
	mqMessage.Set(string(a), "header")
	mqMessage.Set(ctx.RemoteIP(), "ip")
	mqMessage.Set(string(ctx.Request.Body()), "body")
	mqMessage.Set(method, "method")
	mqMessage.Set(queryString, "args")
	err = publish(mqMessage.String(), exchangeName, routeKey, token)
	response(ctx, "", err)
}
func apiReload(ctx *fasthttp.RequestCtx) {
	if !checkRequest(ctx) {
		tokenError(ctx)
		return
	}
	reload()
	response(ctx, "{\"code\":1}", nil)
}
func apiRestart(ctx *fasthttp.RequestCtx) {
	if !checkRequest(ctx) {
		tokenError(ctx)
		return
	}
	err := restart()
	response(ctx, "", err)
}
func apiConfig(ctx *fasthttp.RequestCtx) {
	if !checkRequest(ctx) {
		tokenError(ctx)
		return
	}
	j, e := config()
	response(ctx, j, e)
}
func tokenError(ctx *fasthttp.RequestCtx) {
	ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
	ctx.Response.SetBodyString("token error")
}
func checkRequest(ctx *fasthttp.RequestCtx) (ok bool) {
	token := ctx.QueryArgs().Peek("token")
	if token == nil || string(token) != apiToken {
		ok = false
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}
	ok = true
	return
}
func timeoutFactory(h fasthttp.RequestHandler) fasthttp.RequestHandler {
	return fasthttp.TimeoutHandler(h, apiTimeout, "timeout")
}
func response(ctx *fasthttp.RequestCtx, data string, err error) {
	callback := ctx.QueryArgs().Peek("callback")
	callbackFunc := ""
	if callback != nil && len(callback) > 0 {
		callbackFunc = string(callback)
	}
	ja := gabs.New()
	if err == nil {
		ja.Set(1, "code")
		ja.Set(data, "data")
	} else {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ja.Set(0, "code")
		ja.Set(err.Error(), "data")
	}
	if callbackFunc == "" {
		fmt.Fprintf(ctx, ja.String())
	} else {
		fmt.Fprintf(ctx, callbackFunc+"("+ja.String()+")")
	}
}
func serveAPI(port int, token string) (err error) {
	apiToken = token
	router := fasthttprouter.New()
	router.POST("/message/add", timeoutFactory(apiMessageAdd))
	router.POST("/message/update", timeoutFactory(apiMessageUpdate))
	router.POST("/message/delete", timeoutFactory(apiMessageDelete))
	router.GET("/message/status", timeoutFactory(apiMessageStatus))
	router.POST("/consumer/add", timeoutFactory(apiConsumerAdd))
	router.POST("/consumer/update", timeoutFactory(apiConsumerUpdate))
	router.POST("/consumer/delete", timeoutFactory(apiConsumerDelete))
	router.GET("/consumer/status", timeoutFactory(apiConsumerStatus))
	router.GET("/reload", timeoutFactory(apiReload))
	router.GET("/restart", timeoutFactory(apiRestart))
	router.GET("/config", timeoutFactory(apiConfig))
	log.Infof("Api service started")
	if fasthttp.ListenAndServe("0.0.0.0:"+strconv.Itoa(port), router.Handler) == nil {
		log.Fatalf("start api fail:%s", err)
	}
	return
}
func servePublish(port int) (err error) {
	router := fasthttprouter.New()
	router.POST("/:name", timeoutFactory(apiPublish))
	router.GET("/:name", timeoutFactory(apiPublish))
	log.Infof("Publish service started")
	if fasthttp.ListenAndServe("0.0.0.0:"+strconv.Itoa(port), router.Handler) == nil {
		log.Fatalf("start publish fail:%s", err)
	}
	return
}
