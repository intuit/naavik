package server

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/intuit/naavik/cmd/options"
	"github.com/intuit/naavik/internal/types"
	"github.com/intuit/naavik/pkg/logger"

	// Initialize swagger.
	"github.com/intuit/naavik/internal/server/swagger"
)

func ListenAndServe() (httpServer *http.Server, tlsServer *http.Server) {
	// Logging setup for gin
	if options.GetEnvironment() == types.EnvDev {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	if !options.GetLogColor() {
		gin.DisableConsoleColor()
	}
	// Initialize swagger
	swagger.Initialize()

	router := SetupRouter()
	httpportstr := os.Getenv("MESH_TRAFFIC_PORT")
	httpport, err := strconv.Atoi(httpportstr)
	if len(httpportstr) == 0 || err != nil {
		httpport = 8090
	}
	httpServer = startHTTPServer("", httpport, router)

	env := options.GetEnvironment()
	if env != types.EnvDev {
		tlsServer = startTLSServer("", 8443, router, "./ssl/certificate.crt", "./ssl/certificate.key")
	}
	return httpServer, tlsServer
}

func startHTTPServer(host string, port int, router *gin.Engine) *http.Server {
	hostname := fmt.Sprintf("%s:%d", host, port)
	srvr := &http.Server{
		ReadHeaderTimeout: 5 * time.Second,
		Addr:              hostname,
		Handler:           router,
	}
	logger.Log.Infof("Starting HTTP server on %s", hostname)
	go func(hostname string) {
		err := srvr.ListenAndServe()
		if err != nil {
			logger.Log.Warnf("HTTP server. error=%s", err.Error())
		}
	}(hostname)

	return srvr
}

func startTLSServer(host string, port int, router *gin.Engine, certFile string, keyFile string) *http.Server {
	hostname := fmt.Sprintf("%s:%d", host, port)
	tlsServer := &http.Server{
		ReadHeaderTimeout: 5 * time.Second,
		Addr:              hostname,
		Handler:           router,
	}
	if os.Getenv("ENFORCE_TLS_12") != "false" {
		tlsServer.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	logger.Log.Infof("Starting HTTPS server on %s", hostname)
	go func() {
		tlsErr := tlsServer.ListenAndServeTLS(certFile, keyFile)
		if tlsErr != nil {
			logger.Log.Warnf("HTTPS server. error=%s", tlsErr.Error())
		}
	}()
	return tlsServer
}
