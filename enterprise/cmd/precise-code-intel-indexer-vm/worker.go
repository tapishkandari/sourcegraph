package main

import (
	"github.com/google/uuid"
	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/apiworker"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/apiworker/apiclient"
)

func setupWorker(config Config) []goroutine.BackgroundRoutine {
	queueClient := apiclient.New(apiclient.Options{
		IndexerName:       uuid.New().String(),
		FrontendURL:       config.FrontendURL,
		FrontendAuthToken: config.InternalProxyAuthToken,
		Prefix:            indexQueuePrefix,
		Transport:         defaultTransport,
		OperationName:     "Code Intel Index Manager Client",
	})

	options := apiworker.Options{
		NumHandlers:           config.NumContainers,
		PollInterval:          config.IndexerPollInterval,
		HeartbeatInterval:     config.IndexerHeartbeatInterval,
		WorkerMetrics:         workerMetrics(),
		FrontendURL:           config.FrontendURL,
		FrontendURLFromDocker: config.FrontendURLFromDocker,
		AuthToken:             config.InternalProxyAuthToken,
		FirecrackerImage:      config.FirecrackerImage,
		UseFirecracker:        config.UseFirecracker,
		FirecrackerNumCPUs:    config.FirecrackerNumCPUs,
		FirecrackerMemory:     config.FirecrackerMemory,
		FirecrackerDiskSpace:  config.FirecrackerDiskSpace,
		ImageArchivePath:      config.ImageArchivePath,
		Prefix:                gitPrefix,
	}

	return []goroutine.BackgroundRoutine{apiworker.NewIndexer(queueClient, options)}
}

func workerMetrics() workerutil.WorkerMetrics {
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

	return workerutil.WorkerMetrics{
		HandleOperation: observationContext.Operation(observation.Op{
			Name:         "Processor.Process",
			MetricLabels: []string{"process"},
			Metrics:      metrics,
		}),
	}
}
