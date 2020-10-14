package command

import "context"

type Formatter interface {
	Setup(ctx context.Context, runner Runner, logger *Logger, images []string) error
	Teardown(ctx context.Context, runner Runner, logger *Logger) error
	FormatCommand(cmd *Cmd) []string
}

// TODO - rename
type HandlerOptions struct {
	FirecrackerImage     string
	UseFirecracker       bool
	FirecrackerNumCPUs   int
	FirecrackerMemory    string
	FirecrackerDiskSpace string
	ImageArchivePath     string
}
