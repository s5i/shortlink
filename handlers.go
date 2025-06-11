package main

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strings"

	"github.com/s5i/goutil/authn"
	"github.com/s5i/shortlink/database"
)

type Database interface {
	GetLink(key string) (database.Link, bool, error)
	PutLink(key, value, owner string, force bool) error
	DeleteLink(key, owner string, force bool) error
	ListLinks(owner string, all bool) ([]database.Link, error)

	AddUser(user string) error
	IsUser(user string) (bool, error)
	ListUsers() ([]string, error)
	DeleteUser(user string, keepLinks bool) error
	IsAdmin(user string) (bool, error)
}

func GetLink(db Database, defaultURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := strings.TrimPrefix(r.URL.Path, "/")
		link, ok, err := db.GetLink(key)
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

func EditLink(a *authn.Authn, db Database) http.HandlerFunc {
	return a.RequireUser(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := a.User(r)
		if !ok {
			http.Error(w, "user not authenticated", http.StatusUnauthorized)
			return
		}
		isUser, err := db.IsUser(user)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to read user whitelist: %v", err), http.StatusInternalServerError)
			return
		}
		isAdmin, err := db.IsAdmin(user)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to read admin list: %v", err), http.StatusInternalServerError)
			return
		}
		if !isUser && !isAdmin {
			http.Error(w, fmt.Sprintf("user %q not whitelisted for operation", user), http.StatusForbidden)
			return
		}

		switch r.Method {
		case "GET":
			key, link := "", ""
			keys := r.URL.Query()["key"]
			if len(keys) == 1 {
				key = keys[0]
				if l, _, err := db.GetLink(key); err == nil {
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
				if err := db.PutLink(key, url.String(), user, isAdmin); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
			} else {
				if err := db.DeleteLink(key, user, isAdmin); err != nil {
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

func ListLinks(a *authn.Authn, db Database) http.HandlerFunc {
	return a.RequireUser(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := a.User(r)
		if !ok {
			http.Error(w, "user not authenticated", http.StatusUnauthorized)
			return
		}
		isUser, err := db.IsUser(user)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to read user whitelist: %v", err), http.StatusInternalServerError)
			return
		}
		isAdmin, err := db.IsAdmin(user)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to read admin list: %v", err), http.StatusInternalServerError)
			return
		}
		if !isUser && !isAdmin {
			http.Error(w, fmt.Sprintf("user %q not whitelisted for operation", user), http.StatusForbidden)
			return
		}

		links, err := db.ListLinks(user, isAdmin)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		listTmpl.Execute(w, links)
	}))
}

func EditUsers(a *authn.Authn, db Database) http.HandlerFunc {
	return a.RequireUser(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := a.User(r)
		if !ok {
			http.Error(w, "user not authenticated", http.StatusUnauthorized)
			return
		}
		isAdmin, err := db.IsAdmin(user)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to read admin list: %v", err), http.StatusInternalServerError)
			return
		}
		if !isAdmin {
			http.Error(w, fmt.Sprintf("user %q not whitelisted for operation", user), http.StatusForbidden)
			return
		}

		switch r.Method {
		case "GET":
			users, err := db.ListUsers()
			if err != nil {
				http.Error(w, fmt.Sprintf("failed to list users: %v", err), http.StatusInternalServerError)
				return
			}
			usersTmpl.Execute(w, users)
			return
		case "POST":
			if err := r.ParseForm(); err != nil {
				http.Error(w, "failed to parse POST form", http.StatusBadRequest)
				return
			}
			user, action := r.FormValue("user"), r.FormValue("action")
			switch action {
			case "CREATE":
				if err := db.AddUser(user); err != nil {
					http.Error(w, fmt.Sprintf("failed to add user: %v", err), http.StatusInternalServerError)
					return
				}
			case "DELETE":
				if err := db.DeleteUser(user, true); err != nil {
					http.Error(w, fmt.Sprintf("failed to delete user: %v", err), http.StatusInternalServerError)
					return
				}
			case "DELETE_WITH_LINKS":
				if err := db.DeleteUser(user, false); err != nil {
					http.Error(w, fmt.Sprintf("failed to delete user: %v", err), http.StatusInternalServerError)
					return
				}
			}
			http.Redirect(w, r, "/admin/users", http.StatusFound)
		default:
			http.Error(w, "disallowed method", http.StatusBadRequest)
		}
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

var usersTmpl = template.Must(template.New("").Parse(`<!DOCTYPE html>
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
<th>User</th>
<th></th><th></th>
</tr>
{{range .}}
<tr>
  <td>{{.}}</td>
  <td>
    <form method="POST">
      <input name="user" type="hidden" value="{{.}}" />
      <input name="action" type="hidden" value="DELETE" />
      <input type="submit" value="delete" />
    </form>
  </td>
  <td>
    <form method="POST">
      <input name="user" type="hidden" value="{{.}}" />
      <input name="action" type="hidden" value="DELETE_WITH_LINKS" />
      <input type="submit" value="delete (including links)" />
    </form>
  </td>
</tr>
{{end}}
<tr>
  <form method="POST">
    <td><input name="user" type="text" value="" /></td>
    <input name="action" type="hidden" value="CREATE" />
    <td><input type="submit" value="create" /></td>
    <td></td>
  </form>
</tr>
</table>
</body>
</html>`))
