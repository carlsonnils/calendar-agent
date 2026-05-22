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
func BuildMuxer(a *agent.Agent, conv *agent.Conversation) *http.ServeMux {
	// apply routes to the request multiplexer
    mux := http.NewServeMux()
    mux.HandleFunc("/", MainHandler)
	mux.HandleFunc("GET /login", LoginHandler)
	mux.HandleFunc("POST /login", AuthLoginHandler)
    mux.HandleFunc(
		"POST /api/chat", 
		CheckAuthMiddleware(a.ReplyHandler(context.Background(), conv)))
    // mux.HandleFunc("GET /api/conversations", ListConversationsHandler)

    return mux
}

// run http server to accept http requests to the agent
func StartServer(a *agent.Agent, conv *agent.Conversation) {  
    // build serve muxer
    mux := BuildMuxer(a, conv)

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
func MainHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("server.MainHandler: ", r.Method, r.RequestURI, r.URL.EscapedPath())
	http.ServeFileFS(w, r, os.DirFS(UiPath), r.URL.EscapedPath())
}
