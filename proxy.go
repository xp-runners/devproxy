package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

const noRoute = "no-route"
const timeFormat = "2006/01/06 15:04:05"

type roundtripper struct {
	rt http.RoundTripper
}

func newResponse(status int, text string, args ...interface{}) *http.Response {
	buf := new(bytes.Buffer)
	fmt.Fprintf(buf, text, args...)
	return &http.Response{
		StatusCode: 404,
		Body:       ioutil.NopCloser(buf),
	}
}

func log(req *http.Request, status int, message string) {
	var color string
	if status < 400 {
		color = "34;1"
	} else {
		color = "31;1"
	}

	fmt.Printf("%s \033[%sm%s\033[0m \033[4m%s\033[0m %s\n", time.Now().Format(timeFormat), color, req.Method, req.URL, message)
}

// RoundTrip implements the RoundTripper interface
func (r roundtripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Scheme == noRoute {
		log(req, 404, "404 Not found")
		return newResponse(404, "No route for %s", req.URL.Path), nil
	}

	res, err := r.rt.RoundTrip(req)
	if err != nil {
		log(req, 502, err.Error())
		return newResponse(502, "502 Proxy error %v", err), nil
	}

	log(req, res.StatusCode, res.Status)
	return res, nil
}

type proxy struct {
	Routes map[string]*url.URL
}

// newProxy returns a new proxy instance.
func newProxy() *proxy {
	return &proxy{make(map[string]*url.URL)}
}

// Handle returns a http.Handler suitable for use with HTTP servers
func (p proxy) Handler() http.Handler {
	director := func(req *http.Request) {
		for prefix, target := range p.Routes {
			if strings.HasPrefix(req.URL.Path, prefix) {

				// Transfer origin host
				req.Header.Add("X-Forwarded-Host", req.URL.Host)
				req.Header.Add("X-Forwarded-Proto", req.URL.Scheme)
				req.Host = target.Host

				// Rewrite URL
				req.URL.Scheme = target.Scheme
				req.URL.Host = target.Host
				req.URL.Path = target.Path + strings.Replace(req.URL.Path, prefix, "", 1)
				if target.RawQuery == "" || req.URL.RawQuery == "" {
					req.URL.RawQuery = target.RawQuery + req.URL.RawQuery
				} else {
					req.URL.RawQuery = target.RawQuery + "&" + req.URL.RawQuery
				}
				return
			}
		}

		// No route matched
		req.URL.Scheme = noRoute
	}

	t := &roundtripper{http.DefaultTransport}
	return &httputil.ReverseProxy{Director: director, Transport: t}
}

// Api returns the API handler
func (p proxy) Api() http.Handler {
	use := make(map[string]*url.URL)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/use") {
			prefix := strings.Replace(r.URL.Path, "/use", "", 1)
			if _, exist := p.Routes[prefix]; exist {

				// Restore from "use" pool
				if route, ok := use[prefix]; ok {
					p.Routes[prefix] = route
					fmt.Printf("Configuration updated for %s: %s\n", prefix, p.String())
				}

				w.WriteHeader(201)
				w.Write([]byte(fmt.Sprintf("Using %s from %s", prefix, p.Routes[prefix])))
				return
			}

			http.Error(w, "No such route "+prefix, 400)
		} else if strings.HasPrefix(r.URL.Path, "/develop") {
			url, err := url.Parse(r.FormValue("at"))
			if err != nil {
				http.Error(w, err.Error(), 400)
				return
			}

			prefix := strings.Replace(r.URL.Path, "/develop", "", 1)
			if route, ok := p.Routes[prefix]; ok {

				// Backup original URL into "use" pool
				if _, exist := use[prefix]; !exist {
					use[prefix] = route
				}

				p.Routes[prefix] = url
				fmt.Printf("Configuration updated for %s: %s\n", prefix, p.String())

				w.WriteHeader(201)
				w.Write([]byte(fmt.Sprintf("Developing %s at %s", prefix, url)))
				return
			}

			http.Error(w, "No such route "+prefix, 400)
		} else {
			w.WriteHeader(200)
			w.Write([]byte("Configuration: " + p.String()))
		}
	})
}

// String implements the Stringer interface
func (p proxy) String() string {
	s := "{\n"
	for prefix, target := range p.Routes {
		s += "  " + prefix + " -> " + target.String() + "\n"
	}
	return s + "}"
}
