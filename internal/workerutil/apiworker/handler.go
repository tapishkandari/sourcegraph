package apiworker

import (
	"context"
	"os"
	"sort"

	"github.com/google/uuid"
	"github.com/hashicorp/go-multierror"
	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/apiworker/apiclient"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/apiworker/command"
)

type Handler struct {
	idSet         *IDSet
	commandRunner command.Runner
	options       Options
	uuidGenerator func() (uuid.UUID, error)
}

var _ workerutil.Handler = &Handler{}

// Handle clones the target code into a temporary directory, invokes the target indexer in a fresh
// docker container, and uploads the results to the external frontend API.
func (h *Handler) Handle(ctx context.Context, s workerutil.Store, record workerutil.Record) error {
	index := record.(apiclient.Index)
	// TODO - needs to be some specific type of record here
	// (but can have an additional payload)

	// 🚨 SECURITY: The job logger must be supplied with all sensitive values that may appear
	// in a command constructed and run in the following function. Note that the command and
	// its output may both contain sensitive values, but only values which we directly
	// interpolate into the command. No command that we run on the host leaks environment
	// variables, and the user-specified commands (which could leak their environment) are
	// run in a clean VM.
	logger := command.NewLogger(h.options.AuthToken)

	defer func() {
		type SetLogContents interface {
			SetLogContents(ctx context.Context, id int, contents string) error
		}
		if setLogContents, ok := s.(SetLogContents); ok {
			if err := setLogContents.SetLogContents(ctx, index.ID, logger.String()); err != nil {
				log15.Warn("Failed to upload log for index job", "id", index.ID, "err", err)
			}
		}
	}()

	h.idSet.Add(index.ID)
	defer h.idSet.Remove(index.ID)

	repoDir, err := h.fetchRepository(ctx, h.commandRunner, logger, index.RepositoryName, index.Commit)
	if err != nil {
		return err
	}
	defer func() {
		_ = os.RemoveAll(repoDir)
	}()

	imageMap := map[string]struct{}{}
	for _, dockerStep := range index.DockerSteps {
		imageMap[dockerStep.Image] = struct{}{}
	}

	images := make([]string, 0, len(imageMap))
	for image := range imageMap {
		images = append(images, image)
	}
	sort.Strings(images)

	name, err := h.uuidGenerator()
	if err != nil {
		return err
	}

	commandFormatter := command.NewFormatter(repoDir, command.Options{
		FirecrackerImage:     h.options.FirecrackerImage,
		UseFirecracker:       h.options.UseFirecracker,
		FirecrackerNumCPUs:   h.options.FirecrackerNumCPUs,
		FirecrackerMemory:    h.options.FirecrackerMemory,
		FirecrackerDiskSpace: h.options.FirecrackerDiskSpace,
		ImageArchivePath:     h.options.ImageArchivePath,
		IndexerName:          name.String(),
	})

	if err := commandFormatter.Setup(ctx, h.commandRunner, logger, images); err != nil {
		return err
	}
	defer func() {
		if teardownErr := commandFormatter.Teardown(ctx, h.commandRunner, logger); teardownErr != nil {
			err = multierror.Append(err, teardownErr)
		}
	}()

	for _, dockerStep := range index.DockerSteps {
		dockerStepCommand := command.DockerCommand{
			Image:      dockerStep.Image,
			Arguments:  dockerStep.Commands,
			WorkingDir: dockerStep.Root,
			Env:        dockerStep.Env,
		}

		if err := h.commandRunner.Run(ctx, logger, commandFormatter.FormatCommand(dockerStepCommand)...); err != nil {
			return errors.Wrap(err, "failed to perform docker step")
		}
	}

	return nil
}
