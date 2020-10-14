package command

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
)

type firecrackerFormatter struct {
	name      string
	repoDir   string
	options   HandlerOptions
	formatter Formatter
}

var _ Formatter = &firecrackerFormatter{}

const FirecrackerRepoDir = "/repo-dir"

func NewFirecrackerCommandFormatter(
	name string,
	repoDir string,
	options HandlerOptions,
) Formatter {
	return &firecrackerFormatter{
		name:      name,
		repoDir:   repoDir,
		options:   options,
		formatter: NewDockerFormatter(FirecrackerRepoDir, options),
	}
}

var commonFirecrackerFlags = []string{
	"--runtime", "docker",
	"--network-plugin", "docker-bridge",
}

func (r *firecrackerFormatter) Setup(ctx context.Context, runner Runner, logger *Logger, images []string) error {
	imageMap := map[string]string{}
	for i, image := range images {
		imageMap[fmt.Sprintf("image%d", i)] = image
	}

	for _, key := range orderedKeys(imageMap) {
		if _, err := os.Stat(r.tarfilePathOnHost(key)); err == nil {
			continue
		} else if !os.IsNotExist(err) {
			return err
		}

		if err := r.saveDockerImage(ctx, runner, logger, key, imageMap[key]); err != nil {
			return err
		}
	}

	startCommand := flatten(
		"ignite", "run",
		commonFirecrackerFlags,
		r.resourceFlags(),
		r.copyfileFlags(imageMap),
		"--ssh",
		"--name", r.name,
		sanitizeImage(r.options.FirecrackerImage),
	)
	if err := runner.Run(ctx, logger, startCommand...); err != nil {
		return errors.Wrap(err, "failed to start firecracker vm")
	}

	for _, key := range orderedKeys(imageMap) {
		loadCommand := flatten(
			"ignite", "exec", r.name, "--",
			"docker", "load",
			"-i", r.tarfilePathInVM(key),
		)
		if err := runner.Run(ctx, logger, loadCommand...); err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to load %s", imageMap[key]))
		}
	}

	// Remove tar files inside of vm to clear scratch space
	for _, key := range orderedKeys(imageMap) {
		rmCommand := flatten(
			"ignite", "exec", r.name, "--",
			"rm", r.tarfilePathInVM(key),
		)
		if err := runner.Run(ctx, logger, rmCommand...); err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to remove tarfile for %s", imageMap[key]))
		}
	}

	return nil
}

func (r *firecrackerFormatter) Teardown(ctx context.Context, runner Runner, logger *Logger) error {
	stopCommand := flatten(
		"ignite", "stop",
		commonFirecrackerFlags,
		r.name,
	)
	if err := runner.Run(ctx, logger, stopCommand...); err != nil {
		log15.Warn("Failed to stop firecracker vm", "name", r.name, "err", err)
	}

	removeCommand := flatten(
		"ignite", "rm", "-f",
		commonFirecrackerFlags,
		r.name,
	)
	if err := runner.Run(ctx, logger, removeCommand...); err != nil {
		log15.Warn("Failed to remove firecracker vm", "name", r.name, "err", err)
	}

	return nil
}

func (r *firecrackerFormatter) FormatCommand(cmd *Cmd) []string {
	return flatten("ignite", "exec", r.name, "--", r.formatter.FormatCommand(cmd))
}

func (r *firecrackerFormatter) saveDockerImage(ctx context.Context, runner Runner, logger *Logger, key, image string) error {
	pullCommand := flatten(
		"docker", "pull",
		image,
	)
	if err := runner.Run(ctx, logger, pullCommand...); err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to pull %s", image))
	}

	saveCommand := flatten(
		"docker", "save",
		"-o", r.tarfilePathOnHost(key),
		image,
	)
	if err := runner.Run(ctx, logger, saveCommand...); err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to save %s", image))
	}

	return nil
}

func (r *firecrackerFormatter) resourceFlags() []string {
	return []string{
		"--cpus", strconv.Itoa(r.options.FirecrackerNumCPUs),
		"--memory", r.options.FirecrackerMemory,
		"--size", r.options.FirecrackerDiskSpace,
	}
}

func (r *firecrackerFormatter) copyfileFlags(images map[string]string) (copyfiles []string) {
	for _, key := range orderedKeys(images) {
		copyfiles = append(copyfiles, fmt.Sprintf(
			"%s:%s",
			r.tarfilePathOnHost(key),
			r.tarfilePathInVM(key),
		))
	}

	return intersperse("--copy-files", append(
		[]string{fmt.Sprintf("%s:%s", r.repoDir, FirecrackerRepoDir)},
		copyfiles...,
	))
}

func (r *firecrackerFormatter) tarfilePathOnHost(key string) string {
	return filepath.Join(r.options.ImageArchivePath, fmt.Sprintf("%s.tar", key))
}

func (r *firecrackerFormatter) tarfilePathInVM(key string) string {
	return fmt.Sprintf("/%s.tar", key)
}

var imagePattern = regexp.MustCompile(`([^:@]+)(?::([^@]+))?(?:@sha256:([a-z0-9]{64}))?`)

// sanitizeImage sanitizes the given docker image for use by ignite. The ignite utility has
// some issue parsing docker tags that include a sha256 hash, so we try to remove it from
// any of the image references before passing it to the ignite command.
func sanitizeImage(image string) string {
	if matches := imagePattern.FindStringSubmatch(image); len(matches) == 4 {
		if matches[2] == "" {
			return matches[1]
		}

		return fmt.Sprintf("%s:%s", matches[1], matches[2])
	}

	return image
}
