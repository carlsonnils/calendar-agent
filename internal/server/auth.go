package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"fake.com/nilspcarlson/internal/dal"
	"fake.com/nilspcarlson/internal/jwt"
)

// Login UI handler
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("server.LoginHandler: ", r.Method, r.RequestURI, r.URL.EscapedPath())

	path := strings.TrimPrefix(r.URL.EscapedPath(), "/")
	if path == "login" {
		http.ServeFileFS(w, r, os.DirFS(UiPath), "login/index.html")
		return
	}

	http.ServeFileFS(w, r, os.DirFS(UiPath), r.URL.EscapedPath())
}

// Check the provided username and password and reply with the jwt token
// that has the encrypted signature to check the server created the jwt
func AuthLoginHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("server.AuthLoginHandler: ", r.Method, r.URL.RequestURI())

	// extract username and password from body
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	err := json.NewDecoder(r.Body).Decode(&body)	
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message": "username or password is incorrect"}`))
		return
	}

	// query database for user password
	user, err := dal.GetUser(context.Background(), body.Username)
	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message": "username or password is incorrect"}`))
		return
	} 
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "error querying database for user"}`))
		return
	}

	// check if passwords match
	if body.Password != user.Password {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message": "username or password is incorrect"}`))
		return
	}
	
	// create new jwt token
	token, err := jwt.Make(
		jwt.Header{Alg: "hs256", Typ: "jwt"}, 
		jwt.Body{Username: body.Username},
	)
	if err != nil {
		w.Write([]byte(`{"message": "error making jwt token"}`))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// create http cookie with the token
	c := &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		MaxAge:   31536000, // 1 year in seconds
		HttpOnly: true,     // Not accessible via JavaScript
		// Secure:   true,        // Only sent over HTTPS
		SameSite: http.SameSiteStrictMode,
	}
	
	// set the cookie and reply with status ok (200)
	http.SetCookie(w, c)
	w.WriteHeader(http.StatusOK)
}

// Check to see if the request is coming from an authenticated
// client that has the jwt auth token
// if the auth fails return the login ui
func CheckAuthMiddleware(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("server.CheckAuthMiddleware: ", r.Method, r.RequestURI)
		// check for cookie
		jwtCookie, err := r.Cookie("auth_token")
		if err == http.ErrNoCookie {
			log.Println("server.CheckAuthMiddleware: no auth jwt cookie")
			w.WriteHeader(http.StatusUnauthorized)
			return
		} else if err != nil {
			log.Println("server.CheckAuthMiddleware: error getting auth jwt: ", err)
			w.WriteHeader(http.StatusUnauthorized)
			return  
		}

		// check the signature
		if jwt.CheckSignature(jwtCookie.Value) != true {
			log.Println("server.CheckAuthMiddleware: jwt signature check failed")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		handler(w, r)
	}
}
