package main

import (
	"fmt"
	"github.com/unrolled/render"
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"
	"net/http"
)

var r *render.Render

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

func forms(c web.C, w http.ResponseWriter, req *http.Request) {
	r.HTML(w, http.StatusOK, "forms", map[string]interface{}{
		"Forms": []map[string]string{
			map[string]string{
				"Name": "Test Form",
				"Id":   "asdf123",
			},
			map[string]string{
				"Name": "Test Form 2",
				"Id":   "asdf1234",
			},
		},
	})
}

func newForm(c web.C, w http.ResponseWriter, req *http.Request) {
	url := fmt.Sprintf("/admin/%s", 1)
	http.Redirect(w, req, url, http.StatusTemporaryRedirect)
}

func form(c web.C, w http.ResponseWriter, req *http.Request) {
	formURL := req.URL
	formURL.Scheme = "http"
	if req.TLS != nil {
		formURL.Scheme += "s"
	}
	formURL.Host = req.Host
	formURL.Path = fmt.Sprintf("/s/%s", "asdf123")
	r.HTML(w, http.StatusOK, "form", map[string]interface{}{
		"Name": "Test Form",
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
}

func main() {
	goji.Get("/", index)
	goji.Post("/s/:id", submit)

	admin := web.New()
	admin.Use(middleware.SubRouter)
	admin.Get("/", forms)
	admin.Post("/", newForm)
	admin.Handle("/:id", form)
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
