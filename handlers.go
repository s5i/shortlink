package main

import (
	"html/template"
	"net/http"
	"net/url"
	"strings"

	"github.com/s5i/shortlink/auth"
	"github.com/s5i/shortlink/database"
)

type Database interface {
	Get(key string) (database.Link, bool, error)
	Put(key, value, owner string, force bool) error
	Delete(key, owner string, force bool) error
	List(owner string, all bool) ([]database.Link, error)
}

func GetLink(db Database, defaultURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := strings.TrimPrefix(r.URL.Path, "/")
		link, ok, err := db.Get(key)
		if err != nil || !ok {
			if defaultURL == "" {
				http.Error(w, "", http.StatusNotFound)
				return
			}
			http.Redirect(w, r, defaultURL, http.StatusFound)
			return
		}
		http.Redirect(w, r, link.Value, http.StatusFound)
	}
}

func EditLink(a *auth.Auth, db Database) http.HandlerFunc {
	return a.RequireUser(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := auth.User(r)
		switch r.Method {
		case "GET":
			key, link := "", ""
			keys := r.URL.Query()["key"]
			if len(keys) == 1 {
				key = keys[0]
				if l, _, err := db.Get(key); err == nil {
					link = l.Value
				}
			}
			editTmpl.Execute(w, struct{ Key, Link string }{key, link})
			return
		case "POST":
			if err := r.ParseForm(); err != nil {
				http.Error(w, "failed to parse POST form", http.StatusBadRequest)
				return
			}
			key, link := r.FormValue("key"), r.FormValue("link")
			if link != "" {
				url, err := url.Parse(link)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				if !url.IsAbs() {
					url.Scheme = "http"
				}
				if err := db.Put(key, url.String(), user, false); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
			} else {
				if err := db.Delete(key, user, false); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
			}
			http.Redirect(w, r, "/admin/list", http.StatusFound)
		default:
			http.Error(w, "disallowed method", http.StatusBadRequest)
		}
	}))
}

func ListLinks(a *auth.Auth, db Database) http.HandlerFunc {
	return a.RequireUser(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := auth.User(r)
		links, err := db.List(user, false)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		listTmpl.Execute(w, links)
	}))
}

var editTmpl = template.Must(template.New("").Parse(`<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8" />
</head>
<body>
<div>
  <form method="POST">
      <label>Key: </label><input name="key" type="text" value="{{.Key}}" />
      <br>
      <label>Link: </label><input name="link" type="text" value="{{.Link}}" />
      <br>
      <input type="submit" value="submit" />
  </form>
</div>
</body>
</html>`))

var listTmpl = template.Must(template.New("").Parse(`<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8" />
<style>
table, th, td {
  border: 1px solid;
  border-collapse: collapse;
  padding: 2px;
  text-align: left;
}
tr:nth-child(even) {background-color: #f2f2f2;}
</style>
</head>
<body>
<table>
<tr>
<th>Key</th>
<th>Link</th>
<th>Owner</th>
</tr>
{{range .}}
<tr>
  <td><a href="/admin/edit?key={{.Key}}">{{.Key}}</a></td>
  <td><a href="{{.Value}}">{{.Value}}</a></td>
  <td>{{.Owner}}</td>
</tr>
{{end}}
</table>
</body>
</html>`))
