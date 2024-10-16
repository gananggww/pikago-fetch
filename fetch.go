package snorlax

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
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
	Header map[string]string

	/* with proxy url */
	WithProxy string

	/* ignore ssl certification? */
	WithForceSSL bool

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
	/* show error all */
	Error       error
	finalBody   *bytes.Reader
	finalHeader http.Header
	finalRes    Response
	resBodyByte []byte
	transport   *http.Transport
}

func (o *Fetch) defaultConfig(c Config) {
	if c.Method == "" {
		c.Method = get
	}

	var b []byte

	if c.Method == get || c.Method == del {
		b = nil
	}

	if c.Method == post || c.Method == put || c.Method == patch {
		b, _ = json.Marshal(c.Body)
	}

	o.finalBody = bytes.NewReader(b)

	if c.ContentType == "" {
		c.ContentType = appjson
	}

	o.finalHeader = http.Header{
		contype: []string{c.ContentType},
	}

	for key, val := range c.Header {
		o.finalHeader[key] = []string{val}
	}

	if c.Timeout == 0 {
		c.Timeout = 25
	}

	var proxyUrl *url.URL
	if c.WithProxy != "" {
		proxyUrl, o.Error = url.Parse(c.WithProxy)
		o.transport.Proxy = http.ProxyURL(proxyUrl)
	}

	if c.WithForceSSL {
		defaultTransport := http.DefaultTransport.(*http.Transport).Clone()
		defaultTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: c.WithForceSSL}
		o.transport = defaultTransport
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

	client := &http.Client{Timeout: c.Timeout * time.Second, Transport: o.transport}
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
