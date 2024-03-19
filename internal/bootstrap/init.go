package bootstrap

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/intuit/naavik/cmd/options"
	"github.com/intuit/naavik/internal/controller"
	"github.com/intuit/naavik/internal/leasechecker"
	"github.com/intuit/naavik/internal/types/context"
	"github.com/intuit/naavik/pkg/logger"
)

var log = logger.NewLogger()

func InitNaavik(ctx context.Context) {
	// Initialize logger
	log.SetLogLevel(options.GetLogLevel())
	ctx.Log.Info("Starting Naavik")

	// Start profiling if enabled
	if options.IsProfilingEnabled() {
		StartProfiler(ctx)
	}

	httpServer, tlsServer := StartServer()

	// Initialize state checker
	startStateChecker(ctx)

	StartControllers(ctx)

	shutdown(ctx, httpServer, tlsServer)
}

func startStateChecker(ctx context.Context) {
	// Initialize State Checker
	leaseChecker := leasechecker.GetStateChecker(ctx, options.GetStateChecker())
	leasechecker.RunStateCheck(ctx, leaseChecker)
}

func shutdown(ctx context.Context, httpServer *http.Server, tlsServer *http.Server) {
	gracefulStop := make(chan os.Signal, 1)
	signal.Notify(gracefulStop, syscall.SIGINT, syscall.SIGTERM)

	signal := <-gracefulStop
	startTime := time.Now()
	ctx.Log.Infof("Received %s singal, gracefully shutting down all controllers", signal.String())

	// Shutdown server
	ctx.Log.Infof("Shutting down HTTP server on %s", httpServer.Addr)
	go httpServer.Shutdown(ctx.Context)
	if tlsServer != nil {
		ctx.Log.Infof("Shutting down HTTPS server on %s", tlsServer.Addr)
		go tlsServer.Shutdown(ctx.Context)
	}

	// close the remote controllers stop channel
	controller.StopAllControllers()
	ctx.Log.Int(logger.TimeTakenMSKey, int(time.Since(startTime).Milliseconds())).Info("Graceful shutdown completed")
}
