package server

import (
	"log"
	"net/http"
	"os"
	
	"calendar/internal/agent"
)

var (
	uiPath = "ui"
)

// run http server to accept http requests to the agent
func StartServer(a *agent.Agent) {
    mux := http.NewServeMux()
    mux.HandleFunc("/", HomeHandler)
    mux.HandleFunc("POST /api/chat", a.ReplyHandler)

    go serveHTTP(mux)
    // go serveHTTPS(mux)
    select {}
}

// start a standard socket HTTP server
func serveHTTP(muxer *http.ServeMux) {
	server := &http.Server{
		Addr: ":80",
		Handler: muxer,
	}

	log.Println("starting HTTP server")

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

// start a SSL socket HTTP server
func serveHTTPS(muxer *http.ServeMux) {
    server := &http.Server{
        Addr: ":443",
        Handler: muxer,
    }

    log.Println("starting HTTP server")

    err := server.ListenAndServeTLS(
                "path to .pem file",
                "path to .key.pem file",
        )
    if err != nil {
        log.Fatal(err)
    }
}

// return files from ui directory for home path
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFileFS(w, r, os.DirFS(uiPath), r.URL.EscapedPath())
}
