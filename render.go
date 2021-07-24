package main

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"go.goblog.app/app/pkgs/contenttype"
)

const (
	templatesDir = "templates"
	templatesExt = ".gohtml"

	templateBase               = "base"
	templatePost               = "post"
	templateError              = "error"
	templateIndex              = "index"
	templateTaxonomy           = "taxonomy"
	templateSearch             = "search"
	templateSummary            = "summary"
	templatePhotosSummary      = "photosummary"
	templateEditor             = "editor"
	templateEditorFiles        = "editorfiles"
	templateLogin              = "login"
	templateStaticHome         = "statichome"
	templateBlogStats          = "blogstats"
	templateBlogStatsTable     = "blogstatstable"
	templateComment            = "comment"
	templateCaptcha            = "captcha"
	templateCommentsAdmin      = "commentsadmin"
	templateNotificationsAdmin = "notificationsadmin"
	templateWebmentionAdmin    = "webmentionadmin"
	templateBlogroll           = "blogroll"
	templateGeoMap             = "geomap"
	templateContact            = "contact"
)

func (a *goBlog) initRendering() error {
	a.templates = map[string]*template.Template{}
	templateFunctions := template.FuncMap{
		"md":   a.safeRenderMarkdownAsHTML,
		"html": wrapStringAsHTML,
		// Post specific
		"p":            firstPostParameter,
		"ps":           postParameter,
		"hasp":         postHasParameter,
		"content":      a.postHtml,
		"summary":      a.postSummary,
		"translations": a.postTranslations,
		"shorturl":     a.shortPostURL,
		// Others
		"dateformat": dateFormat,
		"isodate":    isoDateFormat,
		"unixtodate": unixToLocalDateString,
		"now":        localNowString,
		"asset":      a.assetFileName,
		"string":     a.ts.GetTemplateStringVariantFunc(),
		"include":    a.includeRenderedTemplate,
		"urlize":     urlize,
		"sort":       sortedStrings,
		"absolute":   a.getFullAddress,
		"mentions":   a.db.getWebmentionsByAddress,
		"geotitle":   a.geoTitle,
		"geolink":    geoOSMLink,
		"opensearch": openSearchUrl,
		"mbytes":     mBytesString,
	}
	baseTemplate, err := template.New("base").Funcs(templateFunctions).ParseFiles(path.Join(templatesDir, templateBase+templatesExt))
	if err != nil {
		return err
	}
	err = filepath.Walk(templatesDir, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Mode().IsRegular() && path.Ext(p) == templatesExt {
			if name := strings.TrimSuffix(path.Base(p), templatesExt); name != templateBase {
				if a.templates[name], err = template.Must(baseTemplate.Clone()).New(name).ParseFiles(p); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

type renderData struct {
	BlogString                 string
	Canonical                  string
	TorAddress                 string
	Blog                       *configBlog
	User                       *configUser
	Data                       interface{}
	LoggedIn                   bool
	CommentsEnabled            bool
	WebmentionReceivingEnabled bool
	TorUsed                    bool
	EasterEgg                  bool
}

func (a *goBlog) render(w http.ResponseWriter, r *http.Request, template string, data *renderData) {
	a.renderWithStatusCode(w, r, http.StatusOK, template, data)
}

func (a *goBlog) renderWithStatusCode(w http.ResponseWriter, r *http.Request, statusCode int, template string, data *renderData) {
	// Check render data
	a.checkRenderData(r, data)
	// Set content type
	w.Header().Set(contentType, contenttype.HTMLUTF8)
	// Minify and write response
	var tw bytes.Buffer
	err := a.templates[template].ExecuteTemplate(&tw, template, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(statusCode)
	_, err = a.min.Write(w, contenttype.HTML, tw.Bytes())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (a *goBlog) checkRenderData(r *http.Request, data *renderData) {
	// User
	if data.User == nil {
		data.User = a.cfg.User
	}
	// Blog
	if data.Blog == nil {
		if data.BlogString == "" {
			data.BlogString = a.cfg.DefaultBlog
		}
		data.Blog = a.cfg.Blogs[data.BlogString]
	}
	if data.BlogString == "" {
		for s, b := range a.cfg.Blogs {
			if b == data.Blog {
				data.BlogString = s
				break
			}
		}
	}
	// Tor
	if a.cfg.Server.Tor && a.torAddress != "" {
		data.TorAddress = fmt.Sprintf("http://%v%v", a.torAddress, r.RequestURI)
	}
	if torUsed, ok := r.Context().Value(torUsedKey).(bool); ok && torUsed {
		data.TorUsed = true
	}
	// Check login
	if a.isLoggedIn(r) {
		data.LoggedIn = true
	}
	// Check if comments enabled
	data.CommentsEnabled = data.Blog.Comments != nil && data.Blog.Comments.Enabled
	// Check if able to receive webmentions
	data.WebmentionReceivingEnabled = a.cfg.Webmention == nil || !a.cfg.Webmention.DisableReceiving
	// Easter egg
	if ee := a.cfg.EasterEgg; ee != nil && ee.Enabled {
		data.EasterEgg = true
	}
	// Data
	if data.Data == nil {
		data.Data = map[string]interface{}{}
	}
}

func (a *goBlog) includeRenderedTemplate(templateName string, data ...interface{}) (template.HTML, error) {
	if l := len(data); l < 1 || l > 2 {
		return "", errors.New("wrong argument count")
	}
	if rd, ok := data[0].(*renderData); ok {
		if len(data) == 2 {
			nrd := *rd
			nrd.Data = data[1]
			rd = &nrd
		}
		var buf bytes.Buffer
		err := a.templates[templateName].ExecuteTemplate(&buf, templateName, rd)
		return template.HTML(buf.String()), err
	}
	return "", errors.New("wrong arguments")
}
