package main

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpserver"
	"github.com/sourcegraph/sourcegraph/internal/logging"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/apiworker"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/apiworker/apiclient"
)

const Port = 3190

func main() {
	env.Lock()
	env.HandleHelpFlag()
	logging.Init()
	trace.Init(false)

	var (
		frontendURL              = mustGet(rawFrontendURL, "PRECISE_CODE_INTEL_EXTERNAL_URL")
		frontendURLFromDocker    = mustGet(rawFrontendURLFromDocker, "PRECISE_CODE_INTEL_EXTERNAL_URL_FROM_DOCKER")
		internalProxyAuthToken   = mustGet(rawInternalProxyAuthToken, "PRECISE_CODE_INTEL_INTERNAL_PROXY_AUTH_TOKEN")
		indexerPollInterval      = mustParseInterval(rawIndexerPollInterval, "PRECISE_CODE_INTEL_INDEXER_POLL_INTERVAL")
		indexerHeartbeatInterval = mustParseInterval(rawIndexerHeartbeatInterval, "PRECISE_CODE_INTEL_INDEXER_HEARTBEAT_INTERVAL")
		numContainers            = mustParseInt(rawMaxContainers, "PRECISE_CODE_INTEL_MAXIMUM_CONTAINERS")
		firecrackerImage         = mustGet(rawFirecrackerImage, "PRECISE_CODE_INTEL_FIRECRACKER_IMAGE")
		useFirecracker           = mustParseBool(rawUseFirecracker, "PRECISE_CODE_INTEL_USE_FIRECRACKER")
		firecrackerNumCPUs       = mustParseInt(rawFirecrackerNumCPUs, "PRECISE_CODE_INTEL_FIRECRACKER_NUM_CPUS")
		firecrackerMemory        = mustGet(rawFirecrackerMemory, "PRECISE_CODE_INTEL_FIRECRACKER_MEMORY")
		firecrackerDiskSpace     = mustGet(rawFirecrackerDiskSpace, "PRECISE_CODE_INTEL_FIRECRACKER_DISK_SPACE")
		imageArchivePath         = mustGet(rawImageArchivePath, "PRECISE_CODE_INTEL_IMAGE_ARCHIVE_PATH")
	)

	if frontendURLFromDocker == "" {
		frontendURLFromDocker = frontendURL
	}

	prefix := "/.internal-code-intel/index-queue"
	requestMeter := metrics.NewRequestMeter("precise_code_intel_index_manager", "Total number of requests sent.")

	// ot.Transport will propagate opentracing spans.
	defaultTransport := &ot.Transport{
		RoundTripper: requestMeter.Transport(&http.Transport{}, func(u *url.URL) string {
			return strings.TrimPrefix(u.Path, prefix)
		}),
	}

	queueClient := apiclient.New(apiclient.Options{
		IndexerName:       uuid.New().String(),
		FrontendURL:       frontendURL,
		FrontendAuthToken: internalProxyAuthToken,
		Prefix:            prefix,
		Transport:         defaultTransport,
		OperationName:     "Code Intel Index Manager Client",
	})

	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	metrics := metrics.NewOperationMetrics(
		observationContext.Registerer,
		"index_queue_processor",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of records processed"),
	)

	workerMetrics := workerutil.WorkerMetrics{
		HandleOperation: observationContext.Operation(observation.Op{
			Name:         "Processor.Process",
			MetricLabels: []string{"process"},
			Metrics:      metrics,
		}),
	}

	server := httpserver.New(Port, func(router *mux.Router) {})
	indexer := apiworker.NewIndexer(queueClient, apiworker.Options{
		NumHandlers:           numContainers,
		PollInterval:          indexerPollInterval,
		HeartbeatInterval:     indexerHeartbeatInterval,
		WorkerMetrics:         workerMetrics,
		FrontendURL:           frontendURL,
		FrontendURLFromDocker: frontendURLFromDocker,
		AuthToken:             internalProxyAuthToken,
		FirecrackerImage:      firecrackerImage,
		UseFirecracker:        useFirecracker,
		FirecrackerNumCPUs:    firecrackerNumCPUs,
		FirecrackerMemory:     firecrackerMemory,
		FirecrackerDiskSpace:  firecrackerDiskSpace,
		ImageArchivePath:      imageArchivePath,
	},
	)

	go debugserver.Start()
	goroutine.MonitorBackgroundRoutines(server, indexer)
}
