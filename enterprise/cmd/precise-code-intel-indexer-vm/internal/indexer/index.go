package indexer

import "time"

type Index struct {
	// TODO - get rid of what we don't want
	ID             int          `json:"id"`
	Commit         string       `json:"commit"`
	QueuedAt       time.Time    `json:"queuedAt"`
	State          string       `json:"state"`
	FailureMessage *string      `json:"failureMessage"`
	StartedAt      *time.Time   `json:"startedAt"`
	FinishedAt     *time.Time   `json:"finishedAt"`
	ProcessAfter   *time.Time   `json:"processAfter"`
	NumResets      int          `json:"numResets"`
	NumFailures    int          `json:"numFailures"`
	RepositoryID   int          `json:"repositoryId"`
	RepositoryName string       `json:"repositoryName"`
	DockerSteps    []DockerStep `json:"docker_steps"`
	Root           string       `json:"root"`
	Indexer        string       `json:"indexer"`
	IndexerArgs    []string     `json:"indexer_args"`
	Outfile        string       `json:"outfile"`
}

type DockerStep struct {
	Root     string   `json:"root"`
	Image    string   `json:"image"`
	Commands []string `json:"commands"`
}

func (i Index) RecordID() int {
	return i.ID
}
