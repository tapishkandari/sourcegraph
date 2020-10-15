package main

import (
	"github.com/inconshreveable/log15"
	"github.com/prometheus/client_golang/prometheus"
	indexabilityupdater "github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-indexer/internal/indexability_updater"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-indexer/internal/janitor"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-indexer/internal/resetter"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-indexer/internal/scheduler"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func setupSchedulers(s store.Store, config Config) (routines []goroutine.BackgroundRoutine) {
	routines = append(routines, resetter.NewIndexResetter(
		s,
		config.ResetInterval,
		resetter.NewResetterMetrics(prometheus.DefaultRegisterer),
	))

	routines = append(routines, indexabilityupdater.NewUpdater(
		s,
		gitserver.DefaultClient,
		config.IndexabilityUpdaterInterval,
		indexabilityupdater.NewUpdaterMetrics(prometheus.DefaultRegisterer),
		config.IndexMinimumSearchCount,
		float64(config.IndexMinimumSearchRatio)/100,
		config.IndexMinimumPreciseCount,
	))

	routines = append(routines, scheduler.NewScheduler(
		s,
		gitserver.DefaultClient,
		config.SchedulerInterval,
		config.IndexBatchSize,
		config.IndexMinimumTimeSinceLastEnqueue,
		config.IndexMinimumSearchCount,
		float64(config.IndexMinimumSearchRatio)/100,
		config.IndexMinimumPreciseCount,
		scheduler.NewSchedulerMetrics(prometheus.DefaultRegisterer),
	))

	if config.DisableJanitor {
		log15.Warn("Janitor process is disabled.")
		return routines
	}

	return append(routines, janitor.New(
		s,
		config.JanitorInterval,
		janitor.NewJanitorMetrics(prometheus.DefaultRegisterer),
	))
}
