package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dustin/randbo"
	"github.com/garyburd/redigo/redis"
	"github.com/unrolled/render"
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"
)

type Form struct {
	Id          string
	Name        string
	RedirectURL string
}

var r *render.Render
var rp *redis.Pool

func key(args ...string) string {
	args = append([]string{"submit"}, args...)
	return strings.Join(args, ":")
}

func genId() string {
	p := make([]byte, 8)
	randbo.New().Read(p)
	return fmt.Sprintf("%x", p)
}

func getForm(rc redis.Conn, id string, form *Form) error {
	v, err := redis.Values(
		rc.Do("HGETALL", key("form", id)),
	)
	if err != nil {
		return err
	}
	redis.ScanStruct(v, form)
	return nil
}

func index(c web.C, w http.ResponseWriter, req *http.Request) {
	r.HTML(w, http.StatusOK, "index", nil)
}

func showForms(c web.C, w http.ResponseWriter, req *http.Request) {
	var (
		forms []Form
		err   error
	)
	rc := rp.Get()

	fids, err := redis.Strings(rc.Do(
		"SMEMBERS",
		key("forms"),
	))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, fid := range fids {
		var form Form
		err = getForm(rc, fid, &form)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		forms = append(forms, form)
	}

	r.HTML(w, http.StatusOK, "forms", map[string]interface{}{
		"Forms": forms,
	})
}

func createForm(c web.C, w http.ResponseWriter, req *http.Request) {
	var (
		formName    string
		redirectURL string
		err         error
	)
	rc := rp.Get()

	defer func() {
		if err == nil {
			id := genId()

			rc.Do("HMSET", key("form", id),
				"Id", id,
				"Name", formName,
				"RedirectURL", redirectURL,
			)

			rc.Do("SADD", key("forms"), id)

			url := fmt.Sprintf("/admin/%s", id)
			http.Redirect(w, req, url, http.StatusFound)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}()

	if err = req.ParseForm(); err != nil {
		return
	}

	formName = req.PostForm.Get("formName")
	if formName == "" {
		err = errors.New("Form name can't be empty")
		return
	}

	redirectURL = req.PostForm.Get("redirectURL")
	if redirectURL == "" {
		err = errors.New("Redirect URL can't be empty")
		return
	}
}

func showForm(c web.C, w http.ResponseWriter, req *http.Request) {
	var (
		form    Form
		entries []interface{}
		err     error
	)
	rc := rp.Get()

	err = getForm(rc, c.URLParams["id"], &form)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if form == (Form{}) {
		http.Error(w, "Form doesn't exist", http.StatusNotFound)
		return
	}

	formURL := req.URL
	formURL.Scheme = "http"
	if req.TLS != nil {
		formURL.Scheme += "s"
	}
	formURL.Host = req.Host
	formURL.Path = fmt.Sprintf("/s/%s", form.Id)

	fields, err := redis.Strings(rc.Do(
		"SMEMBERS",
		key("form", form.Id, "fields"),
	))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	eids, err := redis.Strings(rc.Do(
		"SMEMBERS",
		key("form", form.Id, "entries"),
	))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, eid := range eids {
		v, err := redis.Strings(
			rc.Do("HGETALL", key("form", form.Id, "entry", eid)),
		)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		entry := make(map[string]string)
		for i := 0; i < len(v); i += 2 {
			entry[v[i]] = v[i+1]
		}

		entries = append(entries, entry)
	}

	r.HTML(w, http.StatusOK, "form", map[string]interface{}{
		"Name":    form.Name,
		"URL":     formURL.String(),
		"Fields":  fields,
		"Entries": entries,
	})
}

func submitEntry(c web.C, w http.ResponseWriter, req *http.Request) {
	var (
		form Form
		err  error
	)
	rc := rp.Get()

	err = getForm(rc, c.URLParams["id"], &form)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if form == (Form{}) {
		http.Error(w, "Form doesn't exist", http.StatusNotFound)
		return
	}

	if err := req.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	eid := genId()

	entry := []interface{}{key("form", form.Id, "entry", eid)}
	for field := range req.PostForm {
		entry = append(entry, field, req.PostForm.Get(field))
		rc.Do("SADD", key("form", form.Id, "fields"), field)
	}
	rc.Do("HMSET", entry...)

	rc.Do("SADD", key("form", form.Id, "entries"), eid)

	http.Redirect(w, req, form.RedirectURL, http.StatusFound)
}

func init() {
	r = render.New(render.Options{
		Layout:        "layout",
		IsDevelopment: true,
	})
	rh := os.Getenv("REDIS_HOST")
	if rh == "" {
		rh = "localhost"
	}
	rp = &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", fmt.Sprintf(
				"%s:6379",
				rh,
			))
			if err != nil {
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func main() {
	goji.Get("/", index)
	goji.Post("/s/:id", submitEntry)

	admin := web.New()
	admin.Use(middleware.SubRouter)
	admin.Get("/", showForms)
	admin.Post("/", createForm)
	admin.Get("/:id", showForm)
	goji.Handle("/admin/*", admin)

	goji.Get("/static/lib/*", http.StripPrefix(
		"/static/lib/",
		http.FileServer(
			http.Dir("./bower_components"),
		),
	))
	goji.Get("/static/*", http.StripPrefix(
		"/static/",
		http.FileServer(
			http.Dir("./static"),
		),
	))

	goji.Serve()
}
