package main

import (
	"fmt"
	"github.com/codegangsta/negroni"
	"github.com/goincremental/negroni-oauth2"
	"github.com/goincremental/negroni-sessions"
	"github.com/goincremental/negroni-sessions/cookiestore"
	"github.com/gorilla/mux"
	"net/http"
	"os"
)

func main() {
	var (
		clientID     = "c65e55f08cc310a2804f"
		clientSecret = "42bf39bb1ed11ece4e26430d4b57fc3e5d4f8631"
		redirectURL  = "https://quiet-depths-2219.herokuapp.com/"
	)

	secureMux := mux.NewRouter()

	// Routes that require a logged in user
	// can be protected by using a separate route handler
	// If the user is not authenticated, they will be
	// redirected to the login path.
	secureMux.HandleFunc("/restrict", func(w http.ResponseWriter, req *http.Request) {
		token := oauth2.GetToken(req)
		fmt.Fprintf(w, "OK: %s", token.Access())
	})

	secure := negroni.New()
	secure.Use(oauth2.LoginRequired())
	secure.UseHandler(secureMux)

	n := negroni.New()
	n.Use(sessions.Sessions("my_session", cookiestore.New([]byte("secret123"))))
	n.Use(oauth2.Github(&oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"user"},
	}))

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

	n.Run(":" + os.Getenv("PORT"))
}
