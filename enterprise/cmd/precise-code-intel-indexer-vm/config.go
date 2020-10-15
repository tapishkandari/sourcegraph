package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/sourcegraph/sourcegraph/internal/env"
)

type Config struct {
	FrontendURL              string
	FrontendURLFromDocker    string
	InternalProxyAuthToken   string
	IndexerPollInterval      time.Duration
	IndexerHeartbeatInterval time.Duration
	NumContainers            int
	FirecrackerImage         string
	UseFirecracker           bool
	FirecrackerNumCPUs       int
	FirecrackerMemory        string
	FirecrackerDiskSpace     string
	ImageArchivePath         string
	errs                     []error
}

func (c *Config) Load() {
	c.FrontendURL = c.get("PRECISE_CODE_INTEL_EXTERNAL_URL", "", "The external URL of the sourcegraph instance.")
	c.FrontendURLFromDocker = c.get("PRECISE_CODE_INTEL_EXTERNAL_URL_FROM_DOCKER", c.FrontendURL, "The external URL of the sourcegraph instance used form within an index container.")
	c.InternalProxyAuthToken = c.get("PRECISE_CODE_INTEL_INTERNAL_PROXY_AUTH_TOKEN", "", "The auth token supplied to the frontend.")
	c.IndexerPollInterval = c.parseInterval("PRECISE_CODE_INTEL_INDEXER_POLL_INTERVAL", "1s", "Interval between queries to the precise-code-intel-index-manager.")
	c.IndexerHeartbeatInterval = c.parseInterval("PRECISE_CODE_INTEL_INDEXER_HEARTBEAT_INTERVAL", "1s", "Interval between heartbeat requests.")
	c.NumContainers = c.parseInt("PRECISE_CODE_INTEL_MAXIMUM_CONTAINERS", "1", "Number of virtual machines or containers that can be running at once.")
	c.FirecrackerImage = c.get("PRECISE_CODE_INTEL_FIRECRACKER_IMAGE", "sourcegraph/ignite-ubuntu:insiders", "The base image to use for virtual machines.")
	c.UseFirecracker = c.parseBool("PRECISE_CODE_INTEL_USE_FIRECRACKER", "true", "Whether to isolate index containers in virtual machines.")
	c.FirecrackerNumCPUs = c.parseInt("PRECISE_CODE_INTEL_FIRECRACKER_NUM_CPUS", "4", "How many CPUs to allocate to each virtual machine or container.")
	c.FirecrackerMemory = c.get("PRECISE_CODE_INTEL_FIRECRACKER_MEMORY", "12G", "How much memory to allocate to each virtual machine or container.")
	c.FirecrackerDiskSpace = c.get("PRECISE_CODE_INTEL_FIRECRACKER_DISK_SPACE", "20G", "How much disk space to allocate to each virtual machine or container.")
	c.ImageArchivePath = c.get("PRECISE_CODE_INTEL_IMAGE_ARCHIVE_PATH", "", "Where to store tar archives of docker images shared by virtual machines.")
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
