package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"path/filepath"

	"os"

	"github.com/Jeffail/gabs"
	"github.com/buaazp/fasthttprouter"
	"github.com/nu7hatch/gouuid"
	logger "github.com/snail007/mini-logger"
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
	Name := string(ctx.QueryArgs().Peek("Name"))
	Comment := string(ctx.QueryArgs().Peek("Comment"))
	DurableS := string(ctx.QueryArgs().Peek("Durable"))
	IsNeedTokenS := string(ctx.QueryArgs().Peek("IsNeedToken"))
	Mode := string(ctx.QueryArgs().Peek("Mode"))
	Token := string(ctx.QueryArgs().Peek("Token"))
	if Name == "" || DurableS == "" || IsNeedTokenS == "" || Token == "" {
		response(ctx, "", errors.New("args required"))
		return
	}
	if _, _, err := getMessage(Name); err == nil {
		response(ctx, "", errors.New("message exists"))
		return
	}
	Durable := false
	if DurableS == "1" {
		Durable = true
	}
	IsNeedToken := false
	if DurableS == "1" {
		IsNeedToken = true
	}
	if Mode != "fanout" && Mode != "topic" && Mode != "direct" {
		response(ctx, "", errors.New("args required"))
		return
	}
	m := message{
		Name:        Name,
		Durable:     Durable,
		Mode:        Mode,
		IsNeedToken: IsNeedToken,
		Token:       Token,
		Comment:     Comment,
		Consumers:   []consumer{},
	}
	err := addMessage(m)
	if err == nil {
		err = writeMessagesToFile(messages, cfg.GetString("consume.DataFile"))
	}
	response(ctx, err, err)
}
func apiMessageUpdate(ctx *fasthttp.RequestCtx) {
	if !checkRequest(ctx) {
		tokenError(ctx)
		return
	}
	Name := string(ctx.QueryArgs().Peek("Name"))
	Comment := string(ctx.QueryArgs().Peek("Comment"))
	DurableS := string(ctx.QueryArgs().Peek("Durable"))
	IsNeedTokenS := string(ctx.QueryArgs().Peek("IsNeedToken"))
	Mode := string(ctx.QueryArgs().Peek("Mode"))
	Token := string(ctx.QueryArgs().Peek("Token"))
	if Name == "" || DurableS == "" || IsNeedTokenS == "" || Token == "" {
		response(ctx, "", errors.New("args required"))
		return
	}
	if _, _, err := getMessage(Name); err != nil {
		response(ctx, "", errors.New("message not found"))
		return
	}
	Durable := false
	if DurableS == "1" {
		Durable = true
	}
	IsNeedToken := false
	if DurableS == "1" {
		IsNeedToken = true
	}
	if Mode != "fanout" && Mode != "topic" && Mode != "direct" {
		response(ctx, "", errors.New("args required"))
		return
	}
	m := message{
		Name:        Name,
		Durable:     Durable,
		Mode:        Mode,
		IsNeedToken: IsNeedToken,
		Token:       Token,
		Comment:     Comment,
	}
	err := updateMessage(m)
	if err == nil {
		err = writeMessagesToFile(messages, cfg.GetString("consume.DataFile"))
	}
	response(ctx, err, err)
}
func apiMessageDelete(ctx *fasthttp.RequestCtx) {
	if !checkRequest(ctx) {
		tokenError(ctx)
		return
	}
	exchangeName := string(ctx.QueryArgs().Peek("Name"))
	msg, _, err := getMessage(exchangeName)
	if err != nil {
		response(ctx, "", errors.New("message not found"))
		return
	}
	err = deleteMessage(*msg)
	if err == nil {
		err = writeMessagesToFile(messages, cfg.GetString("consume.DataFile"))
	}
	response(ctx, err, err)
}
func apiMessageStatus(ctx *fasthttp.RequestCtx) {
	if !checkRequest(ctx) {
		tokenError(ctx)
		return
	}
	exchangeName := string(ctx.QueryArgs().Peek("Name"))
	j, err := statusMessage(exchangeName)
	if err != nil {
		response(ctx, "", err)
		return
	}
	ctx.WriteString("{\"code\":1,\"data\":" + j + "}")
}
func apiConsumerAdd(ctx *fasthttp.RequestCtx) {
	if !checkRequest(ctx) {
		tokenError(ctx)
		return
	}
	exchangeName := string(ctx.QueryArgs().Peek("Name"))
	IDUUID, _ := uuid.NewV4()
	Comment := string(ctx.QueryArgs().Peek("Comment"))
	CodeS := string(ctx.QueryArgs().Peek("Code"))
	CheckCodeS := string(ctx.QueryArgs().Peek("CheckCode"))
	RouteKey := string(ctx.QueryArgs().Peek("RouteKey"))
	TimeoutS := string(ctx.QueryArgs().Peek("Timeout"))
	URL := string(ctx.QueryArgs().Peek("URL"))

	if exchangeName == "" || CodeS == "" || CheckCodeS == "" || TimeoutS == "" || URL == "" {
		response(ctx, "", errors.New("args required"))
		return
	}
	msg, _, err := getMessage(exchangeName)
	if err != nil {
		response(ctx, "", err)
		return
	}
	ID := IDUUID.String()
	CheckCode := false
	if CheckCodeS == "1" {
		CheckCode = true
	}
	if ok, err := regexp.Match(`[1-9]\d{1,2}`, []byte(CodeS)); !ok || err != nil {
		response(ctx, "", errors.New("args required"))
		return
	}
	if ok, err := regexp.Match(`[1-9]\d*`, []byte(TimeoutS)); !ok || err != nil {
		response(ctx, "", errors.New("args required"))
		return
	}
	CheckCode = true
	codeI, _ := strconv.Atoi(CodeS)
	TimeoutI, _ := strconv.Atoi(TimeoutS)
	c := consumer{
		ID:        ID,
		Comment:   Comment,
		CheckCode: CheckCode,
		Code:      float64(codeI),
		Timeout:   float64(TimeoutI),
		URL:       URL,
		RouteKey:  RouteKey,
	}
	err = addConsumer(*msg, c)
	if err == nil {
		err = writeMessagesToFile(messages, cfg.GetString("consume.DataFile"))
	}
	response(ctx, err, err)
}
func apiConsumerUpdate(ctx *fasthttp.RequestCtx) {
	if !checkRequest(ctx) {
		tokenError(ctx)
		return
	}
	exchangeName := string(ctx.QueryArgs().Peek("Name"))
	ID := string(ctx.QueryArgs().Peek("ID"))
	Comment := string(ctx.QueryArgs().Peek("Comment"))
	CodeS := string(ctx.QueryArgs().Peek("Code"))
	CheckCodeS := string(ctx.QueryArgs().Peek("CheckCode"))
	RouteKey := string(ctx.QueryArgs().Peek("RouteKey"))
	TimeoutS := string(ctx.QueryArgs().Peek("Timeout"))
	URL := string(ctx.QueryArgs().Peek("URL"))

	if exchangeName == "" || CodeS == "" || CheckCodeS == "" || TimeoutS == "" || URL == "" {
		response(ctx, "", errors.New("args required"))
		return
	}
	msg, _, err := getMessage(exchangeName)
	if err != nil {
		response(ctx, "", errors.New("message not found"))
		return
	}
	c, _, _, err := getConsumer(exchangeName, ID)
	if err != nil {
		response(ctx, "", errors.New("consumer not found"))
		return
	}
	CheckCode := false
	if CheckCodeS == "1" {
		CheckCode = true
	}
	if ok, err := regexp.Match(`[1-9]\d{1,2}`, []byte(CodeS)); !ok || err != nil {
		response(ctx, "", errors.New("args required"))
		return
	}
	if ok, err := regexp.Match(`[1-9]\d*`, []byte(TimeoutS)); !ok || err != nil {
		response(ctx, "", errors.New("args required"))
		return
	}
	CheckCode = true
	codeI, _ := strconv.Atoi(CodeS)
	TimeoutI, _ := strconv.Atoi(TimeoutS)
	c0 := consumer{
		ID:        c.ID,
		Comment:   Comment,
		CheckCode: CheckCode,
		Code:      float64(codeI),
		Timeout:   float64(TimeoutI),
		URL:       URL,
		RouteKey:  RouteKey,
	}
	err = updateConsumer(*msg, c0)
	if err == nil {
		err = writeMessagesToFile(messages, cfg.GetString("consume.DataFile"))
	}
	response(ctx, err, err)
}
func apiConsumerDelete(ctx *fasthttp.RequestCtx) {
	if !checkRequest(ctx) {
		tokenError(ctx)
		return
	}
	exchangeName := string(ctx.QueryArgs().Peek("Name"))
	ID := string(ctx.QueryArgs().Peek("ID"))
	msg, _, err := getMessage(exchangeName)
	if err != nil {
		response(ctx, "", errors.New("message not found"))
		return
	}
	c, _, _, err := getConsumer(exchangeName, ID)
	if err != nil {
		response(ctx, "", errors.New("consumer not found"))
		return
	}
	err = deleteConsumer(*msg, *c)
	if err == nil {
		err = writeMessagesToFile(messages, cfg.GetString("consume.DataFile"))
	}
	response(ctx, err, err)
}
func apiConsumerStatus(ctx *fasthttp.RequestCtx) {
	if !checkRequest(ctx) {
		tokenError(ctx)
		return
	}
	exchangeName := string(ctx.QueryArgs().Peek("Name"))
	consumerID := string(ctx.QueryArgs().Peek("ID"))
	j, e := statusConsumer(exchangeName, consumerID)
	d := ""
	if e == nil {
		d = j.String()
		ctx.WriteString("{\"code\":1,\"data\":" + d + "}")
	} else {
		response(ctx, e, e)
	}
}
func apiPublish(ctx *fasthttp.RequestCtx) {
	queryString := string(ctx.QueryArgs().QueryString())
	exchangeName := ctx.UserValue("name").(string)
	msg, _, err := getMessage(exchangeName)
	if err != nil {
		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.WriteString(err.Error())
		return
	}

	tokenB := ctx.Request.Header.Peek("Token")
	token := string(tokenB)
	if msg.IsNeedToken && token != msg.Token {
		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.WriteString("token error")
		return
	}
	routeKeyB := ctx.Request.Header.Peek("RouteKey")
	routeKey := string(routeKeyB)
	method := strings.ToLower(string(ctx.Request.Header.Method()))
	headerMap := make(map[string]string)
	ignores := cfg.GetStringSlice("publish.IgnoreHeaders")
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
	if err == nil {
		ctx.Response.SetStatusCode(fasthttp.StatusNoContent)
		return
	}
	ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
	ctx.WriteString(err.Error())
	return
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
func apiLogList(ctx *fasthttp.RequestCtx) {
	if !checkRequest(ctx) {
		tokenError(ctx)
		return
	}
	f, _ := filepath.Abs(cfg.GetString("log.dir"))
	fs, _ := filepath.Glob(f + "/*")
	var list []string
	for _, v := range fs {
		list = append(list, filepath.Base(v))
	}
	response(ctx, list, nil)
}
func apiLogFile(ctx *fasthttp.RequestCtx) {
	if !checkRequest(ctx) {
		tokenError(ctx)
		return
	}
	filename := string(ctx.QueryArgs().Peek("file"))
	if filename == "" || strings.ContainsAny(filename, "/\\") {
		response(ctx, "", errors.New("args required"))
		return
	}
	f, _ := filepath.Abs(cfg.GetString("log.dir"))

	file := filepath.Join(f, filename)
	if _, e := os.Stat(file); os.IsNotExist(e) {
		response(ctx, "", errors.New("file not found"))
		return
	}
	ctx.Response.Header.Set("Content-Type", "application/force-download")
	ctx.Response.Header.Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	ctx.Response.Header.Set("Content-Transfer-Encoding", "binary")
	ctx.SendFile(file)
}
func apiLog(ctx *fasthttp.RequestCtx) {
	if !checkRequest(ctx) {
		tokenError(ctx)
		return
	}
	keyword := string(ctx.QueryArgs().Peek("keyword"))
	logType := string(ctx.QueryArgs().Peek("type"))
	if logType == "" {
		response(ctx, "", errors.New("args required"))
		return
	}
	file, _ := filepath.Abs(filepath.Join(cfg.GetString("log.dir"), logType) + ".log")
	commandStr := ""
	if keyword == "" {
		commandStr = fmt.Sprintf("tail -n 100 -f %s", file)
	} else {
		commandStr = fmt.Sprintf("grep \"%s\" %s |tail -n 100", keyword, file)
	}
	cmd := exec.Command("bash", "-c", commandStr)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()
	response(ctx, string(stdout.Bytes()), err)

}
func tokenError(ctx *fasthttp.RequestCtx) {
	ctx.Response.SetBodyString("{code:0,data:\"token error\"}")
}
func checkRequest(ctx *fasthttp.RequestCtx) (ok bool) {
	token := ctx.QueryArgs().Peek("api-token")
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
func response(ctx *fasthttp.RequestCtx, data interface{}, err error) {
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
		ja.Set(0, "code")
		ja.Set(err.Error(), "data")
	}
	if callbackFunc == "" {
		fmt.Fprintf(ctx, ja.String())
	} else {
		fmt.Fprintf(ctx, callbackFunc+"("+ja.String()+")")
	}
}

func serveAPI(listen, token string) (err error) {
	ctx := log.With(logger.Fields{"func": "serveAPI"})
	apiToken = token
	router := fasthttprouter.New()
	router.GET("/message/add", timeoutFactory(apiMessageAdd))
	router.GET("/message/update", timeoutFactory(apiMessageUpdate))
	router.GET("/message/delete", timeoutFactory(apiMessageDelete))
	router.GET("/message/status", timeoutFactory(apiMessageStatus))
	router.GET("/consumer/add", timeoutFactory(apiConsumerAdd))
	router.GET("/consumer/update", timeoutFactory(apiConsumerUpdate))
	router.GET("/consumer/delete", timeoutFactory(apiConsumerDelete))
	router.GET("/consumer/status", timeoutFactory(apiConsumerStatus))
	router.GET("/reload", timeoutFactory(apiReload))
	router.GET("/restart", timeoutFactory(apiRestart))
	router.GET("/config", timeoutFactory(apiConfig))
	router.GET("/log", timeoutFactory(apiLog))
	router.GET("/log/file", apiLogFile)
	router.GET("/log/list", timeoutFactory(apiLogList))
	ctx.Infof("Api service started")
	var h = func(ctx *fasthttp.RequestCtx) {
		defer access(ctx)
		router.Handler(ctx)
	}
	if fasthttp.ListenAndServe(listen, h) == nil {
		ctx.Fatalf("start api fail:%s", err)
	}
	return
}
func servePublish(listen string) (err error) {
	ctx := log.With(logger.Fields{"func": "servePublish"})
	router := fasthttprouter.New()
	router.POST("/:name", timeoutFactory(apiPublish))
	router.GET("/:name", timeoutFactory(apiPublish))
	ctx.Infof("Publish service started")
	var h = func(ctx *fasthttp.RequestCtx) {
		defer access(ctx)
		router.Handler(ctx)
	}
	if fasthttp.ListenAndServe(listen, h) == nil {
		ctx.Fatalf("start publish fail:%s", err)
	}
	return
}
func access(ctx *fasthttp.RequestCtx) {
	fields := logger.Fields{
		"code":       strconv.Itoa(ctx.Response.StatusCode()),
		"uri":        string(ctx.RequestURI()),
		"remoteAddr": strings.Split(ctx.RemoteAddr().String(), ":")[0],
		"method":     string(ctx.Method()),
		"host":       string(ctx.Request.Host()),
		"referer":    string(ctx.Request.Header.Referer()),
		"userAgent":  string(ctx.Request.Header.UserAgent()),
		"response":   string(ctx.Response.Body()),
	}
	accessLog.With(fields).Info("")
}
