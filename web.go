package main

import (
	"encoding/json"
	"fmt"
	"github.com/codegangsta/negroni"
	"github.com/goincremental/negroni-oauth2"
	"github.com/goincremental/negroni-sessions"
	"github.com/goincremental/negroni-sessions/cookiestore"
	"github.com/gorilla/mux"
	goauth "golang.org/x/oauth2"
	"io/ioutil"
	"net/http"
)

var githubConf = &oauth2.Config{
	ClientID:     "c65e55f08cc310a2804f",
	ClientSecret: "8d3358ea56d08213dbc9d08753d8ee1278058f3e",
	RedirectURL:  "http://app-3a214f62-5196-4ca9-aa63-9f2f90298127.cleverapps.io/oauth2callback",
	Scopes:       []string{"user"},
}

const (
	keyToken = "oauth2_token"
)

func cl(conf *goauth.Config, t goauth.Token) *http.Client {
	return conf.Client(goauth.NoContext, &t)
}

func Restrict(w http.ResponseWriter, req *http.Request) {
	s := sessions.GetSession(req)
	if s.Get(keyToken) == nil {
		return
	}

	data := s.Get(keyToken).([]byte)
	var tk goauth.Token
	json.Unmarshal(data, &tk)

	token := oauth2.GetToken(req)
	client := cl((*goauth.Config)(githubConf), tk)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		fmt.Fprintf(w, "NO: %s", err.Error())
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(w, "NO: %s", err.Error())
		return
	}

	fmt.Fprintf(w, "OK: %s, --> %s", token.Access(), body)
}

func main() {
	secureMux := mux.NewRouter()

	// Routes that require a logged in user
	// can be protected by using a separate route handler
	// If the user is not authenticated, they will be
	// redirected to the login path.
	secureMux.HandleFunc("/restrict", Restrict)

	secure := negroni.New()
	secure.Use(oauth2.LoginRequired())
	secure.UseHandler(secureMux)

	n := negroni.New()
	n.Use(sessions.Sessions("my_session", cookiestore.New([]byte("secret123"))))
	n.Use(oauth2.Github(githubConf))

	router := mux.NewRouter()

	//routes added to mux do not require authentication
	router.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		token := oauth2.GetToken(req)
		if token == nil || !token.Valid() {
			fmt.Fprintf(w, "not logged in, or the access token is expired")
			return
		}
		fmt.Fprintf(w, "logged in")
		return
	})

	//There is probably a nicer way to handle this than repeat the restricted routes again
	//of course, you could use something like gorilla/mux and define prefix / regex etc.
	router.Handle("/restrict", secure)

	n.UseHandler(router)

	n.Run("0.0.0.0:8080")
}
