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
	proxy := &httputil.ReverseProxy{Director: director}
	log.Print("Proxy: ")
	log.Println(proxy)
	r := mux.NewRouter()
	log.Print("r: ")
	log.Println(r)
	log.Println("About to call Handle")
	r.Handle("/{version}/{namespace}/projects/{projectID}/servers/{serverID}/endpoint/{service}{path:.*}", handle(proxy))
	log.Println("Back from handle")
	server := &http.Server{
		Addr:           ":8080",
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: http.DefaultMaxHeaderBytes,
	}
	log.Println("Created server")
	return server.ListenAndServe()
}

func director(req *http.Request) {
	targetURL, _ := url.Parse("http://localhost:8888")
	log.Print("targetURL: ")
	log.Println(targetURL)
	log.Println(req)
	vars := mux.Vars(req)
	req.URL.Host = targetURL.Host
	req.URL.Scheme = targetURL.Scheme
	req.URL.Path = vars["path"]
}

func handle(proxy *httputil.ReverseProxy) http.Handler {
	log.Println("In handle function")
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
		log.Print("Token from session: ")
		log.Println(sessionToken)
		if !ok {
			token = r.URL.Query().Get("access_token")
			log.Print("Token from URL Query Parm: ")
			log.Println(token)
		} else {
			token = sessionToken.(string)
		}
		log.Println("About to check token")
		log.Println(args.ApiRoot)
		if checkToken(args.ApiRoot, token) {
			log.Println("checkToken was successful")
			session.Values["token"] = token
			session.Save(r, w)
			proxy.ServeHTTP(w, r)
			return
		}
		log.Println("checkToken was unsuccessful")
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
