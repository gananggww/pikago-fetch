package snorlax

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"golang.org/x/net/context/ctxhttp"
)

type Config struct {
	/*
		full uri like ganangxcoding.com/code-for/betterworld
	*/
	Uri string
	/*
		method supported like GET, POST, DELETE, PUT, PATCH
	*/
	Method string

	/*
		fill body with object or struct json type
	*/
	Body interface{}

	/* timout in secod default is application/json */
	ContentType string
	/* map your header */
	Header    map[string]string
	withProxy string

	/* timout in secod default is 25 second */
	Timeout time.Duration
}

func NewFetch() *Fetch {
	return &Fetch{}
}

type Response struct {
	Header http.Header
	Code   int
}

type Fetch struct {
	Error       error
	finalBody   *bytes.Reader
	finalHeader http.Header
	finalRes    Response
	resBodyByte []byte
}

func (o *Fetch) defaultConfig(c Config) {
	if c.Method == "" {
		c.Method = "GET"
	}

	var b []byte

	if c.Method == "GET" || c.Method == "DELETE" {
		b = nil
	}

	if c.Method == "POST" || c.Method == "UPDATE" || c.Method == "PUT" || c.Method == "PATCH" {
		b, _ = json.Marshal(c.Body)
	}

	o.finalBody = bytes.NewReader(b)

	if c.ContentType == "" {
		c.ContentType = "application/json"
	}

	o.finalHeader = http.Header{
		"Content-Type": []string{c.ContentType},
	}

	for key, val := range c.Header {
		o.finalHeader[key] = []string{val}
	}

	if c.Timeout == 0 {
		c.Timeout = 25
	}

}

func (o *Fetch) Fetch(ctx context.Context, c Config) *Fetch {
	o.defaultConfig(c)

	var req *http.Request
	var res *http.Response

	req, o.Error = http.NewRequestWithContext(nil, c.Method, c.Uri, o.finalBody)
	if o.Error != nil {
		return o
	}

	for k, v := range c.Header {
		req.Header.Add(k, v)
	}

	client := &http.Client{Timeout: c.Timeout * time.Second}
	// tracingClient := apmhttp.WrapClient(client)

	res, o.Error = ctxhttp.Do(ctx, client, req)

	ne, ok := o.Error.(net.Error)
	if ok && ne.Timeout() {
		o.Error = fmt.Errorf("error timeout to call party | %s", o.Error.Error())
		return o
	}
	if o.Error != nil {
		return o
	}

	defer res.Body.Close()

	o.finalRes.Header = res.Header

	o.resBodyByte, o.Error = io.ReadAll(res.Body)
	if o.Error != nil {
		return o
	}

	o.finalRes.Code = res.StatusCode

	return o

}

/*
generate body response, should Fetch() first
*/
func (o *Fetch) Catch(object interface{}) *Fetch {
	json.Unmarshal(o.resBodyByte, &object)
	if o.Error != nil {
		return o
	}

	return o
}

/*
generate code and header response, should Fetch() first
*/

func (o *Fetch) CatchOther() (res Response) {

	res.Code = o.finalRes.Code
	res.Header = o.finalRes.Header

	return
}
