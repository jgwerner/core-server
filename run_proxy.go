package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

var sessionStore *sessions.CookieStore

func init() {
	secret := securecookie.GenerateRandomKey(32)
	sessionStore = &sessions.CookieStore{
		Codecs: securecookie.CodecsFromPairs(secret),
		Options: &sessions.Options{
			Path:     "/",
			MaxAge:   86400 * 7, // 7 days
			HttpOnly: true,
		},
	}
	sessionStore.MaxAge(sessionStore.Options.MaxAge)
}

type RunProxy struct {
	gen *RunGeneric
}

func (rp *RunProxy) Run() error {
	go rp.gen.Run()
	err := os.Chdir(args.ResourceDir)
	if err != nil {
		return err
	}
	targetURL, _ := url.Parse("http://localhost:8888")
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	r := mux.NewRouter()
	r.Handle("/{version}/{namespace}/projects/{projectID}/servers/{serverID}/endpoint/{service}{path:.*}", handle(proxy))
	server := &http.Server{
		Addr:           ":8080",
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: http.DefaultMaxHeaderBytes,
	}
	return server.ListenAndServe()
}

func handle(proxy *httputil.ReverseProxy) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		connectionHeader := strings.ToLower(r.Header.Get("Connection"))
		upgradeHeader := strings.ToLower(r.Header.Get("Upgrade"))
		if connectionHeader == "upgrade" && upgradeHeader == "websocket" {
			err := hijack(w, r)
			if err != nil {
				log.Println(err)
				http.Error(w, "Websocket error", http.StatusInternalServerError)
			}
			return
		}
		sessionName := fmt.Sprintf("session-%s", args.ServerID)
		session, err := sessionStore.Get(r, sessionName)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		var token string
		sessionToken, ok := session.Values["token"]
		if !ok {
			token = r.URL.Query().Get("access_token")
		} else {
			token = sessionToken.(string)
		}
		if checkToken(args.ApiRoot, token) {
			session.Values["token"] = token
			session.Save(r, w)
			proxy.ServeHTTP(w, r)
			return
		}
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
	})
}

func hijack(w http.ResponseWriter, r *http.Request) error {
	hijacker := w.(http.Hijacker)
	conn, err := net.Dial("tcp", ":8888")
	if err != nil {
		return err
	}
	hconn, _, err := hijacker.Hijack()
	if err != nil {
		return err
	}
	defer hconn.Close()
	defer conn.Close()

	err = r.Write(conn)
	if err != nil {
		return err
	}
	errChan := make(chan error, 2)
	cpy := func(dst io.Writer, src io.Reader) {
		_, err := io.Copy(dst, src)
		errChan <- err
	}
	go cpy(conn, hconn)
	go cpy(hconn, conn)
	return <-errChan
}
