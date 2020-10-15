package main

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/logging"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/internal/tracer"
)

const port = 3190
const gitPrefix = "/.internal-code-intel/git"
const indexQueuePrefix = "/.internal-code-intel/index-queue"

var requestMeter = metrics.NewRequestMeter("precise_code_intel_index_manager", "Total number of requests sent.")

// ot.Transport will propagate opentracing spans.
var defaultTransport = &ot.Transport{
	RoundTripper: requestMeter.Transport(&http.Transport{}, func(u *url.URL) string {
		return strings.TrimPrefix(u.Path, indexQueuePrefix)
	}),
}

func main() {
	config := Config{}
	config.Load()

	env.Lock()
	env.HandleHelpFlag()
	logging.Init()
	tracer.Init()
	trace.Init(true)

	if err := config.Validate(); err != nil {
		log.Fatalf("failed to read config: %s", err)
	}

	go debugserver.Start()

	goroutine.MonitorBackgroundRoutines(append(
		setupServer(),
		setupWorker(config)...,
	)...)
}
