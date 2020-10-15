package command

import "context"

type Formatter interface {
	Setup(ctx context.Context, runner Runner, logger *Logger, images []string) error
	Teardown(ctx context.Context, runner Runner, logger *Logger) error
	FormatCommand(command DockerCommand) []string
}

type Options struct {
	FirecrackerImage     string
	UseFirecracker       bool
	FirecrackerNumCPUs   int
	FirecrackerMemory    string
	FirecrackerDiskSpace string
	ImageArchivePath     string
	IndexerName          string
}

type DockerCommand struct {
	Image      string
	Arguments  []string
	WorkingDir string
	Env        []string
}

func NewFormatter(repoDir string, options Options) Formatter {
	if !options.UseFirecracker {
		return newDockerFormatter(repoDir, options)
	}

	return newFirecrackerFormatter(options.IndexerName, repoDir, options)
}
