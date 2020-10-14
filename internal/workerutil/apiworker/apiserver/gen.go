package apiserver

//go:generate env GOBIN=$PWD/.bin GO111MODULE=on go install github.com/efritz/go-mockgen
//go:generate $PWD/.bin/go-mockgen -f github.com/sourcegraph/sourcegraph/internal/workerutil/apiworker/apiserver -i MetadataStore -o mock_metadata_store_test.go
