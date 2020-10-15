package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/sourcegraph/sourcegraph/internal/env"
)

type Config struct {
	ResetInterval                    time.Duration
	SchedulerInterval                time.Duration
	IndexabilityUpdaterInterval      time.Duration
	JanitorInterval                  time.Duration
	IndexBatchSize                   int
	IndexMinimumTimeSinceLastEnqueue time.Duration
	IndexMinimumSearchCount          int
	IndexMinimumSearchRatio          int
	IndexMinimumPreciseCount         int
	DisableJanitor                   bool
	MaximumTransactions              int
	RequeueDelay                     time.Duration
	CleanupInterval                  time.Duration
	MaximumMissedHeartbeats          int
	FrontendURL                      string
	FrontendURLFromDocker            string
	InternalProxyAuthToken           string
	errs                             []error
}

func (c *Config) Load() {
	c.ResetInterval = c.parseInterval("PRECISE_CODE_INTEL_RESET_INTERVAL", "1m", "How often to reset stalled indexes.")
	c.IndexabilityUpdaterInterval = c.parseInterval("PRECISE_CODE_INTEL_INDEXABILITY_UPDATER_INTERVAL", "30m", "Interval between scheduled indexability updates.")
	c.SchedulerInterval = c.parseInterval("PRECISE_CODE_INTEL_SCHEDULER_INTERVAL", "30m", "Interval between scheduled index updates.")
	c.JanitorInterval = c.parseInterval("PRECISE_CODE_INTEL_JANITOR_INTERVAL", "1m", "Interval between cleanup runs.")
	c.IndexBatchSize = c.parseInt("PRECISE_CODE_INTEL_INDEX_BATCH_SIZE", "100", "Number of indexable repos to consider on each index scheduler update.")
	c.IndexMinimumTimeSinceLastEnqueue = c.parseInterval("PRECISE_CODE_INTEL_INDEX_MINIMUM_TIME_SINCE_LAST_ENQUEUE", "24h", "Interval between indexing runs of the same repo.")
	c.IndexMinimumSearchCount = c.parseInt("PRECISE_CODE_INTEL_INDEX_MINIMUM_SEARCH_COUNT", "50", "Minimum number of search events to trigger indexing for a repo.")
	c.IndexMinimumSearchRatio = c.parsePercent("PRECISE_CODE_INTEL_INDEX_MINIMUM_SEARCH_RATIO", "50", "Minimum ratio of search events to total events to trigger indexing for a repo.")
	c.IndexMinimumPreciseCount = c.parseInt("PRECISE_CODE_INTEL_INDEX_MINIMUM_PRECISE_COUNT", "1", "Minimum number of precise events to trigger indexing for a repo.")
	c.DisableJanitor = c.parseBool("PRECISE_CODE_INTEL_DISABLE_JANITOR", "false", "Set to true to disable the janitor process during system migrations.")
	c.MaximumTransactions = c.parseInt("PRECISE_CODE_INTEL_MAXIMUM_TRANSACTIONS", "10", "Number of index jobs that can be active at once.")
	c.RequeueDelay = c.parseInterval("PRECISE_CODE_INTEL_REQUEUE_DELAY", "1m", "The requeue delay of index jobs assigned to an unreachable indexer.")
	c.CleanupInterval = c.parseInterval("PRECISE_CODE_INTEL_CLEANUP_INTERVAL", "10s", "Interval between cleanup runs.")
	c.MaximumMissedHeartbeats = c.parseInt("PRECISE_CODE_INTEL_MAXIMUM_MISSED_HEARTBEATS", "5", "The number of heartbeats an indexer must miss to be considered unreachable.")
	// TODO - these are new, add them to deploy-sourcegraph* repos
	c.FrontendURL = c.get("PRECISE_CODE_INTEL_EXTERNAL_URL", "", "The external URL of the sourcegraph instance.")
	c.FrontendURLFromDocker = c.get("PRECISE_CODE_INTEL_EXTERNAL_URL_FROM_DOCKER", c.FrontendURL, "The external URL of the sourcegraph instance used form within an index container.")
	c.InternalProxyAuthToken = c.get("PRECISE_CODE_INTEL_INTERNAL_PROXY_AUTH_TOKEN", "", "The auth token supplied to the frontend.")
}

func (c *Config) Validate() error {
	if len(c.errs) == 0 {
		return nil
	}

	err := c.errs[0]
	for i := 1; i < len(c.errs); i++ {
		err = multierror.Append(err, c.errs[i])
	}

	return err
}

func (c *Config) get(name, defaultValue, description string) string {
	rawValue := env.Get(name, defaultValue, description)
	if rawValue == "" {
		c.errs = append(c.errs, fmt.Errorf("invalid value %q for %s: no value supplied", rawValue, name))
		return ""
	}

	return rawValue
}

func (c *Config) parseInt(name, defaultValue, description string) int {
	rawValue := env.Get(name, defaultValue, description)
	i, err := strconv.ParseInt(rawValue, 10, 64)
	if err != nil {
		c.errs = append(c.errs, fmt.Errorf("invalid int %q for %s: %s", rawValue, name, err))
		return 0
	}

	return int(i)
}

func (c *Config) parsePercent(name, defaultValue, description string) int {
	value := c.parseInt(name, defaultValue, description)
	if value < 0 || value > 100 {
		c.errs = append(c.errs, fmt.Errorf("invalid percent %q for %s: must be 0 <= p <= 100", value, name))
		return 0
	}

	return value
}

func (c *Config) parseInterval(name, defaultValue, description string) time.Duration {
	rawValue := env.Get(name, defaultValue, description)
	d, err := time.ParseDuration(rawValue)
	if err != nil {
		c.errs = append(c.errs, fmt.Errorf("invalid duration %q for %s: %s", rawValue, name, err))
		return 0
	}

	return d
}

func (c *Config) parseBool(name, defaultValue, description string) bool {
	rawValue := env.Get(name, defaultValue, description)
	v, err := strconv.ParseBool(rawValue)
	if err != nil {
		c.errs = append(c.errs, fmt.Errorf("invalid bool %q for %s: %s", rawValue, name, err))
		return false
	}

	return v
}
