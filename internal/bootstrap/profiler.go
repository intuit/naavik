package bootstrap

import (
	"fmt"

	"github.com/grafana/pyroscope-go"
	"github.com/intuit/naavik/cmd/options"
	"github.com/intuit/naavik/internal/types"
	"github.com/intuit/naavik/internal/types/context"
	"github.com/intuit/naavik/pkg/logger"
)

func StartProfiler(ctx context.Context) {
	ctx.Log.Infof("Starting pyroscope profiler %s", logger.Log.GetLogLevel())
	profilerEndpoint := options.GetProfilerEndpoint()
	config := pyroscope.Config{
		ApplicationName: types.NaavikName,
		ServerAddress:   fmt.Sprintf("http://%s", profilerEndpoint),
		Logger:          ctx.Log,
		ProfileTypes: []pyroscope.ProfileType{
			pyroscope.ProfileCPU,
			pyroscope.ProfileInuseObjects,
			pyroscope.ProfileAllocObjects,
			pyroscope.ProfileInuseSpace,
			pyroscope.ProfileAllocSpace,
			pyroscope.ProfileGoroutines,
			pyroscope.ProfileMutexCount,
			pyroscope.ProfileMutexDuration,
			pyroscope.ProfileBlockCount,
			pyroscope.ProfileBlockDuration,
		},
	}

	// All profiles are enabled by default
	_, err := pyroscope.Start(config)
	if err != nil {
		ctx.Log.Errorf("Error starting pyroscope profiler: %v", err)
	}
}
