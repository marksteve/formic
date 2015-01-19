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
	Id   string
	Name string
}

var r *render.Render
var rp *redis.Pool

func key(args ...string) string {
	args = append([]string{"submit"}, args...)
	return strings.Join(args, ":")
}

func index(c web.C, w http.ResponseWriter, req *http.Request) {
	r.HTML(w, http.StatusOK, "index", nil)
}

func submit(c web.C, w http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, "Form: %s", c.URLParams["id"])
	for name := range req.PostForm {
		fmt.Fprintf(w, "%s: %s", name, req.PostForm.Get(name))
	}
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
		v, err := redis.Values(
			rc.Do("HGETALL", key("form", fid)),
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		redis.ScanStruct(v, &form)
		forms = append(forms, form)
	}

	r.HTML(w, http.StatusOK, "forms", map[string]interface{}{
		"Forms": forms,
	})
}

func createForm(c web.C, w http.ResponseWriter, req *http.Request) {
	var (
		formName string
		err      error
	)
	rc := rp.Get()

	defer func() {
		if err == nil {
			p := make([]byte, 8)
			randbo.New().Read(p)
			id := fmt.Sprintf("%x", p)

			rc.Do("HMSET", key("form", id),
				"Id", id,
				"Name", formName,
			)
			rc.Do("SADD", key("forms"), id)

			url := fmt.Sprintf("/admin/%s", id)
			http.Redirect(w, req, url, http.StatusTemporaryRedirect)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}()

	if err = req.ParseForm(); err != nil {
		return
	}

	formName = req.PostForm.Get("name")
	if formName == "" {
		err = errors.New("Form name can't be empty")
		return
	}
}

func showForm(c web.C, w http.ResponseWriter, req *http.Request) {
	var (
		form Form
		err  error
	)
	rc := rp.Get()

	v, err := redis.Values(
		rc.Do("HGETALL", key("form", c.URLParams["id"])),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	redis.ScanStruct(v, &form)

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

	r.HTML(w, http.StatusOK, "form", map[string]interface{}{
		"Name": form.Name,
		"URL":  formURL.String(),
		"Entries": []map[string]string{
			map[string]string{
				"Name":  "Mark Steve Samson",
				"Email": "hello@marksteve.com",
			},
			map[string]string{
				"Name":  "Mark Steve Samson",
				"Email": "marksteve@insynchq.com",
			},
		},
	})
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
	goji.Post("/s/:id", submit)

	admin := web.New()
	admin.Use(middleware.SubRouter)
	admin.Get("/", showForms)
	admin.Post("/", createForm)
	admin.Handle("/:id", showForm)
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
