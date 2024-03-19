package bootstrap

import (
	"net/http"

	"github.com/intuit/naavik/internal/server"
)

func StartServer() (httpServer *http.Server, tlsServer *http.Server) {
	return server.ListenAndServe()
}
