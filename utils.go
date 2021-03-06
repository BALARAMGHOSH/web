// Copyright 2013 Jamie Hall. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package web

import (
	"net/http"
	"sync"
	"time"
)

// DoNotCache uses the Cache-Control, Pragma, and Expires HTTP headers
// to advise the client not to cache the response.
func DoNotCache(w http.ResponseWriter) {
	header := w.Header()
	header.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	header.Set("Pragma", "no-cache")
	header.Set("Expires", "0")
}

// Cache uses the Last-Modified, Expires, and Vary HTTP headers to
// advise the client to cache the response for the given duration.
func Cache(w http.ResponseWriter, modTime time.Time, duration time.Duration) {
	header := w.Header()
	if !modTime.IsZero() {
		header.Set("Last-Modified", modTime.UTC().Format(http.TimeFormat))
	}
	header.Set("Expires", time.Now().Add(duration).UTC().Format(http.TimeFormat))
	header.Set("Vary", "Accept-Encoding")
}

// May be useful in cache durations.
// Slightly less than one year, to conform to RFC 2616.
var OneYear time.Duration = time.Hour * 24 * 364

// RedirectToHTTPS takes an HTTP request and redirects it to the same
// page, but using HTTPS. Be careful not to use in serving HTTPS, or
// an infinite redirection loop will occur.
func RedirectToHTTPS(w http.ResponseWriter, r *http.Request) {
	url := r.URL
	url.Scheme = "https"
	url.Host = r.Host
	http.Redirect(w, r, url.String(), 301)
}

// RedirectToHttpsHandler can be used as an http.Handler which uses
// RedirectToHTTPS above.
var RedirectToHttpsHandler = Handler(RedirectToHTTPS)

// RedirectToHTTP takes an HTTPS request and redirects it to the same
// page, but using HTTP. Be careful not to use in serving HTTP, or
// an infinite redirection loop will occur.
func RedirectToHTTP(w http.ResponseWriter, r *http.Request) {
	url := r.URL
	url.Scheme = "http"
	url.Host = r.Host
	http.Redirect(w, r, url.String(), 301)
}

// RedirectToHttpHandler can be used as an http.Handler which uses
// RedirectToHTTP above.
var RedirectToHttpHandler = Handler(RedirectToHTTP)

// Redirect can be used as an http.Handler which redirects all requests
// to the enclosed string.
//
//	site := web.NewSite("example.com", 80, nil)
//	site.Always(Redirect("www.example.com")) // Redirects all requests to this URL.
//
type Redirect string

func (s Redirect) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, string(s), 301)
}

// UsePath creates a Handler which will call the given
// PathHandler with a fixed path, allowing multiple URLs
// to refer to the same content more simply.
//
//	site := web.NewSite("example.com", 80, nil)
//
//	// Requests to / and /index.html will both be served
//	// by calling serveHTML with the http.ResponseWriter,
//	// the *http.Request, and the path "content/index.html".
//	site.Equals(web.UsePath("content/index.html", serveHTML), "/", "/index.html")
//
func UsePath(path string, handler PathHandler) http.Handler {
	return Handler(func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, path)
	})
}

// UsePrefix works similarly to UsePath, but will simply
// prepend the request path with the given prefix.
//
//	site := web.NewSite("example.com", 80, nil)
//
//	// Requests ending in .jpg or .png will all be served
//	// by calling serveHTML with the http.ResponseWriter,
//	// the *http.Request, and the path "images" + request.URL.Path.
//	site.HasSuffix(web.UsePrefix{"images", serveImage}, ".jpg", ".png")
//
func UsePrefix(prefix string, handler PathHandler) http.Handler {
	return Handler(func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, prefix+r.URL.Path)
	})
}

// PathHandler represents a handler which takes a string
// describing the filepath to the resource to serve.
type PathHandler func(http.ResponseWriter, *http.Request, string)

// Handler can be used as a shorter http.HandlerFunc.
type Handler func(http.ResponseWriter, *http.Request)

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h(w, r)
}

// PageViews is a simple structure
// for recording page view counts
// in a thread-safe manner.
type PageViews struct {
	sync.Mutex
	count int64
}

// Add increments the count.
func (p *PageViews) Add() {
	p.Lock()
	p.count++
	p.Unlock()
}

// Count returns the number of page views.
func (p *PageViews) Count() (count int64) {
	p.Lock()
	count = p.count
	p.Unlock()
	return count
}
