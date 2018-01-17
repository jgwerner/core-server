package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

var (
	sessionStore *sessions.CookieStore
	serverPath   string
)

func init() {
	secret := []byte(args.SecretKey)
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

type Transport struct {
	tr http.RoundTripper
}

func (tr *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	reqc := new(http.Request)
	*reqc = *req
	if req.Body != nil {
		var buf bytes.Buffer
		defer req.Body.Close()
		body := io.TeeReader(req.Body, &buf)
		req.Body = ioutil.NopCloser(body)
		reqc.Body = ioutil.NopCloser(&buf)
	}
	resp, err := tr.tr.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusNotFound {
		req.URL.Path = strings.TrimPrefix(req.URL.Path, serverPath)
		resp, err = tr.tr.RoundTrip(reqc)
	}
	return resp, err
}

type RunProxy struct {
	gen *RunGeneric
}

func (rp *RunProxy) Run() error {
	serverPath = fmt.Sprintf("/%s/%s/projects/%s/servers/%s/endpoint/proxy",
		args.Version, args.Namespace, args.ProjectID, args.ServerID)
	go rp.gen.Run()
	err := os.Chdir(args.ResourceDir)
	if err != nil {
		return err
	}
	targetURL, _ := url.Parse("http://localhost:8888")
	proxy := &httputil.ReverseProxy{
		Transport: &Transport{http.DefaultTransport},
		Director: func(req *http.Request) {
			req.URL.Host = targetURL.Host
			req.URL.Scheme = targetURL.Scheme
		},
		ModifyResponse: func(resp *http.Response) error {
			loc, _ := resp.Location()
			if loc != nil && !strings.HasPrefix(loc.Path, serverPath) {
				loc.Host = args.ApiRoot
				loc.Path = serverPath + loc.Path
				resp.Header.Set("Location", loc.String())
			}
			return nil
		},
	}
	server := &http.Server{
		Addr:           ":8080",
		Handler:        handle(proxy),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: http.DefaultMaxHeaderBytes,
	}
	return server.ListenAndServe()
}

func handle(proxy *httputil.ReverseProxy) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := handleWS(w, r)
		if err != nil {
			log.Println(err)
			http.Error(w, "Websocket error", http.StatusInternalServerError)
			return
		}
		sessionName := fmt.Sprintf("session-%s", args.ServerID)
		session, err := sessionStore.Get(r, sessionName)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		var token string
		sessionToken, ok := session.Values["token"]
		if !ok {
			chunks := strings.Split(r.Header.Get("Authorization"), " ")
			if len(chunks) == 2 {
				token = chunks[1]
			} else {
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}
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

func handleWS(w http.ResponseWriter, r *http.Request) error {
	var err error
	connectionHeader := strings.ToLower(r.Header.Get("Connection"))
	upgradeHeader := strings.ToLower(r.Header.Get("Upgrade"))
	if connectionHeader == "upgrade" && upgradeHeader == "websocket" {
		err = hijack(w, r)
	}
	return err
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
