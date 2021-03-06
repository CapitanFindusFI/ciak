package server

import (
	"fmt"
	"github.com/GaruGaru/ciak/server/auth"
	"github.com/GaruGaru/ciak/utils"
	"github.com/gorilla/sessions"
	"github.com/sirupsen/logrus"
	"html/template"
	"net/http"
)

var UnauthenticatedUrls = []string{
	"/login",
	"/probe",
	"/api/login",
}

type LoginPage struct {
	Title string
}

var loginPageTemplate = template.Must(template.ParseFiles("static/base.html", "static/login.html"))

var store = sessions.NewCookieStore([]byte("test"))

func (s CiakServer) LoginApiHandler(w http.ResponseWriter, r *http.Request) {

	username := r.FormValue("username")
	password := r.FormValue("password")

	authUser, err := s.Authenticator.Authenticate(username, password)
	if err == nil {
		s.createSession(w, r, authUser)
		http.Redirect(w, r, "/", http.StatusFound)
	} else {
		http.Redirect(w, r, "/login", 302)
	}

}

func (s CiakServer) createSession(w http.ResponseWriter, r *http.Request, user auth.User) {
	session, err := store.Get(r, "user")

	if err != nil {
		logrus.Warn("Error creating the session: ", err)
		return
	}

	session.Values["username"] = user.Name
	store.Save(r, w, session)
}

func (s CiakServer) LoginPageHandler(w http.ResponseWriter, r *http.Request) {
	loginPageTemplate.Execute(w, LoginPage{
		Title: "Login",
	})
}

func (s CiakServer) SessionAuthMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if !s.Config.AuthenticationEnabled || utils.StringIn(r.URL.Path, UnauthenticatedUrls) {
			next.ServeHTTP(w, r)
			return
		}

		session, err := store.Get(r, "user")

		if err != nil {
			fmt.Println("Session error ", err)
			return
		}

		if !session.IsNew {
			next.ServeHTTP(w, r)
		} else {
			http.Redirect(w, r, "/login", http.StatusFound)
		}

	})
}
