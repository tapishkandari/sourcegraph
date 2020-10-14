package apiworker

//go:generate env GOBIN=$PWD/.bin GO111MODULE=on go install github.com/efritz/go-mockgen
//go:generate $PWD/.bin/go-mockgen -f github.com/sourcegraph/sourcegraph/internal/workerutil/apiworker/command -i Runner -o mock_command_runner_test.go
//go:generate $PWD/.bin/go-mockgen -f github.com/sourcegraph/sourcegraph/internal/workerutil/apiworker -i queueClient -o mock_queue_client_test.go
