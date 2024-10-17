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
	ErrorDebug  error
	finalBody   *bytes.Reader
	finalHeader http.Header
	finalRes    Response
	resBodyByte []byte
	transport   *http.Transport
	finalUri    *url.URL
	finalMethod string
	finalTo     time.Duration
	req         *http.Request
	res         *http.Response
}

func (o *Fetch) defaultConfig(c Config) (err error) {

	o.finalUri, err = url.ParseRequestURI(c.Uri)
	if err != nil {
		err = fmt.Errorf("error parse uri : %s", err.Error())
		return
	}

	if o.finalUri.Hostname() == "" {
		err = fmt.Errorf("error parse uri : no host detected")
		return
	}

	if c.Method == "" {
		o.finalMethod = get
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
		o.finalTo = 25000 * time.Millisecond
	} else {
		o.finalTo = c.Timeout * time.Millisecond

	}

	if c.WithForceSSL {
		defaultTransport := http.DefaultTransport.(*http.Transport).Clone()
		defaultTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: c.WithForceSSL}
		o.transport = defaultTransport
	}

	var proxyUrl *url.URL
	if c.WithProxy != "" {
		proxyUrl, err = url.Parse(c.WithProxy)
		if err != nil {
			return
		}
		o.transport.Proxy = http.ProxyURL(proxyUrl)
	}

	return

}

func (o *Fetch) Fetch(ctx context.Context, c Config) *Fetch {

	o.Error = o.defaultConfig(c)
	if o.Error != nil {
		o.ErrorDebug = o.Error
		return o
	}

	o.req, o.Error = http.NewRequestWithContext(ctx, o.finalMethod, o.finalUri.String(), o.finalBody)
	if o.Error != nil {
		o.Error = fmt.Errorf("snorlax: request")
		o.ErrorDebug = o.Error
		return o
	}

	o.req.Header = o.finalHeader

	client := &http.Client{Timeout: o.finalTo, Transport: o.transport}

	o.res, o.Error = ctxhttp.Do(ctx, client, o.req)
	ne, ok := o.Error.(net.Error)
	if ok && ne.Timeout() {
		o.Error = fmt.Errorf("snorlax: timeout")
		o.ErrorDebug = o.Error
		return o
	}
	if o.Error != nil {
		o.Error = fmt.Errorf("snorlax: get request")
		o.ErrorDebug = o.Error
		return o
	}

	defer o.res.Body.Close()

	o.finalRes.Header = o.res.Header

	o.resBodyByte, o.Error = io.ReadAll(o.res.Body)
	if o.Error != nil {
		return o
	}

	o.finalRes.Code = o.res.StatusCode

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
