package server

import (
	"context"
	"log"
	"net/http"
	"os"
	
	"fake.com/nilspcarlson/internal/agent"
)

var (
	UiPath string
)

// Build muxer
func BuildMuxer(a *agent.Agent) *http.ServeMux {
	// one conversation per server
	conv := agent.NewConversation()

	// apply routes to the request multiplexer
    mux := http.NewServeMux()
    mux.HandleFunc("/", HomeHandler)
    mux.HandleFunc("POST /api/chat", a.ReplyHandler(context.Background(), conv))
    mux.HandleFunc("GET /api/conversations", ListConversationsHandler)

    return mux
}

// run http server to accept http requests to the agent
func StartServer(a *agent.Agent) {  
    // build serve muxer
    mux := BuildMuxer(a)

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
	log.Println("server.HomeHandler request path: ", r.RequestURI)
	http.ServeFileFS(w, r, os.DirFS(UiPath), r.URL.EscapedPath())
}
