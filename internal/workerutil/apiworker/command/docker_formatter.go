package command

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
)

type dockerFormatter struct {
	repoDir string
	options Options
}

var _ Formatter = &dockerFormatter{}

func newDockerFormatter(
	repoDir string,
	options Options,
) Formatter {
	return &dockerFormatter{
		repoDir: repoDir,
		options: options,
	}
}

func (r *dockerFormatter) Setup(ctx context.Context, runner Runner, logger *Logger, images []string) error {
	return nil
}

func (r *dockerFormatter) Teardown(ctx context.Context, runner Runner, logger *Logger) error {
	return nil
}

func (r *dockerFormatter) FormatCommand(command DockerCommand) []string {
	return flatten(
		"docker", "run", "--rm",
		r.resourceFlags(),
		r.volumeFlags(),
		r.workingdirectoryFlags(command.WorkingDir),
		r.envFlags(command.Env),
		command.Image,
		command.Arguments,
	)
}

func (r *dockerFormatter) resourceFlags() []string {
	return []string{
		"--cpus", strconv.Itoa(r.options.FirecrackerNumCPUs),
		"--memory", r.options.FirecrackerMemory,
	}
}

func (r *dockerFormatter) volumeFlags() []string {
	return []string{"-v", fmt.Sprintf("%s:/data", r.repoDir)}
}

func (r *dockerFormatter) workingdirectoryFlags(wd string) []string {
	return []string{"-w", filepath.Join("/data", wd)}
}

func (r *dockerFormatter) envFlags(env []string) []string {
	return intersperse("-e", env)
}
