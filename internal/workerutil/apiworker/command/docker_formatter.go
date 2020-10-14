package command

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
)

type dockerFormatter struct {
	repoDir string
	options HandlerOptions
}

var _ Formatter = &dockerFormatter{}

func NewDockerFormatter(
	repoDir string,
	options HandlerOptions,
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

func (r *dockerFormatter) FormatCommand(cmd *Cmd) []string {
	return flatten(
		"docker", "run", "--rm",
		r.resourceFlags(),
		r.volumeFlags(),
		r.workingdirectoryFlags(cmd.wd),
		r.envFlags(cmd.env),
		cmd.image,
		cmd.command,
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

func (r *dockerFormatter) envFlags(env map[string]string) []string {
	var flattened []string
	for _, key := range orderedKeys(env) {
		flattened = append(flattened, fmt.Sprintf("%s=%s", key, env[key]))
	}

	return intersperse("-e", flattened)
}
