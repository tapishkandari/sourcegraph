package main

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	"github.com/sourcegraph/sourcegraph/internal/db/basestore"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/apiworker/apiclient"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/apiworker/apiserver"
)

func setupServer(s store.Store, config Config) []goroutine.BackgroundRoutine {
	metadataStore := &storeWrapper{s}
	workerStore := store.WorkerutilIndexStore(s)
	options := apiserver.Options{
		Port:                  Port,
		MaximumTransactions:   config.MaximumTransactions,
		RequeueDelay:          config.RequeueDelay,
		UnreportedIndexMaxAge: config.CleanupInterval * time.Duration(config.MaximumMissedHeartbeats),
		DeathThreshold:        config.CleanupInterval * time.Duration(config.MaximumMissedHeartbeats),
		CleanupInterval:       config.CleanupInterval,
		ToIndex: func(record workerutil.Record) (apiclient.Index, error) {
			return transformRecord(record, config.FrontendURLFromDocker, config.InternalProxyAuthToken)
		},
	}

	return []goroutine.BackgroundRoutine{apiserver.NewServer(metadataStore, workerStore, options)}
}

type storeWrapper struct {
	store store.Store
}

func (sw *storeWrapper) SetIndexLogContents(ctx context.Context, tx basestore.ShareableStore, indexID int, contents string) error {
	return sw.store.With(tx).SetIndexLogContents(ctx, indexID, contents)
}

//
//
//

func transformRecord(record workerutil.Record, frontendURLFromDocker, internalProxyAuthToken string) (apiclient.Index, error) {
	index := record.(store.Index)

	uploadURL, err := url.Parse(frontendURLFromDocker)
	if err != nil {
		return apiclient.Index{}, err
	}
	uploadURL.User = url.UserPassword("indexer", internalProxyAuthToken)

	outfile := "dump.lsif"
	if index.Outfile != "" {
		outfile = index.Outfile
	}

	args := []string{
		"lsif", "upload",
		"-no-progress",
		"-repo", index.RepositoryName,
		"-commit", index.Commit,
		"-upload-route", uploadRoute,
		"-file", outfile,
	}

	dockerSteps := make([]apiclient.DockerStep, 0, len(index.DockerSteps)+2)
	for _, dockerStep := range index.DockerSteps {
		dockerSteps = append(dockerSteps, apiclient.DockerStep{
			Root:     dockerStep.Root,
			Image:    dockerStep.Image,
			Commands: dockerStep.Commands,
			Env:      nil, // TODO - dockerStep.Env,
		})
	}

	if index.Indexer != "" {
		dockerSteps = append(dockerSteps, apiclient.DockerStep{
			Root:     index.Root,
			Image:    index.Indexer,
			Commands: index.IndexerArgs,
			Env:      nil,
		})
	}
	dockerSteps = append(dockerSteps, apiclient.DockerStep{
		Root:     index.Root,
		Image:    uploadImage,
		Commands: args,
		Env:      []string{fmt.Sprintf("SRC_ENDPOINT=%q", uploadURL.String())},
	})

	return apiclient.Index{
		ID:             index.ID,
		Commit:         index.Commit,
		RepositoryName: index.RepositoryName,
		DockerSteps:    dockerSteps,
	}, nil
}
