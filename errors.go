package main

import (
	"fmt"
	"net/http"

	ct "github.com/elnormous/contenttype"
	"go.goblog.app/app/pkgs/contenttype"
)

type errorData struct {
	Title   string
	Message string
}

func (a *goBlog) serve404(w http.ResponseWriter, r *http.Request) {
	a.serveError(w, r, fmt.Sprintf("%s was not found", r.RequestURI), http.StatusNotFound)
}

func (a *goBlog) serveNotAllowed(w http.ResponseWriter, r *http.Request) {
	a.serveError(w, r, "", http.StatusMethodNotAllowed)
}

func (a *goBlog) serveError(w http.ResponseWriter, r *http.Request, message string, status int) {
	// Init the first time
	if len(a.errorCheckMediaTypes) == 0 {
		a.errorCheckMediaTypes = append(a.errorCheckMediaTypes, ct.NewMediaType(contenttype.HTML))
	}
	// Check message
	if message == "" {
		message = http.StatusText(status)
	}
	// Check if request accepts HTML
	if mt, _, err := ct.GetAcceptableMediaType(r, a.errorCheckMediaTypes); err != nil || mt.String() != a.errorCheckMediaTypes[0].String() {
		// Request doesn't accept HTML
		http.Error(w, message, status)
		return
	}
	a.renderWithStatusCode(w, r, status, templateError, &renderData{
		Data: &errorData{
			Title:   fmt.Sprintf("%d %s", status, http.StatusText(status)),
			Message: message,
		},
	})
}
