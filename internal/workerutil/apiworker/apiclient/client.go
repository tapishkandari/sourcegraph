package apiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"golang.org/x/net/context/ctxhttp"
)

// Client is the interface to the precise-code-intel-index-manager.
type Client struct {
	options    Options
	userAgent  string
	httpClient *http.Client
}

type Options struct {
	IndexerName       string
	FrontendURL       string
	FrontendAuthToken string
	Prefix            string
	Transport         http.RoundTripper
	OperationName     string
}

// NewClient creates a new Client with the given unique name targetting hte given external frontend API.
func New(options Options) *Client {
	return &Client{
		options:    options,
		userAgent:  filepath.Base(os.Args[0]),
		httpClient: &http.Client{Transport: options.Transport},
	}
}

// Dequeue returns a queued index record for processing. This record can be marked as completed
// or failed by calling Complete with the same identifier. While processing, the identifier of
// the record must appear in all heartbeat requests.
func (c *Client) Dequeue(ctx context.Context, index *Index) (bool, error) {
	url, err := makeIndexManagerURL(c.options.FrontendURL, c.options.FrontendAuthToken, c.options.Prefix, "dequeue")
	if err != nil {
		return false, err
	}

	payload, err := marshalPayload(DequeueRequest{
		IndexerName: c.options.IndexerName,
	})
	if err != nil {
		return false, err
	}

	hasContent, body, err := c.do(ctx, "POST", url, payload)
	if err != nil {
		return false, err
	}
	if !hasContent {
		return false, nil
	}
	defer body.Close()

	if err := json.NewDecoder(body).Decode(&index); err != nil {
		return false, err
	}

	return true, nil
}

// SetLogContents updates a currently processing index record with the given log contents.
func (c *Client) SetLogContents(ctx context.Context, indexID int, contents string) error {
	url, err := makeIndexManagerURL(c.options.FrontendURL, c.options.FrontendAuthToken, c.options.Prefix, "setlog")
	if err != nil {
		return err
	}

	payload, err := marshalPayload(SetLogRequest{
		IndexerName: c.options.IndexerName,
		IndexID:     indexID,
		Contents:    contents,
	})
	if err != nil {
		return err
	}

	return c.doAndDrop(ctx, "POST", url, payload)
}

// Complete marks the target index record as complete or errored depending on the existence of an
// error message.
func (c *Client) Complete(ctx context.Context, indexID int, indexErr error) error {
	url, err := makeIndexManagerURL(c.options.FrontendURL, c.options.FrontendAuthToken, c.options.Prefix, "complete")
	if err != nil {
		return err
	}

	rawPayload := CompleteRequest{
		IndexerName: c.options.IndexerName,
		IndexID:     indexID,
	}
	if indexErr != nil {
		rawPayload.ErrorMessage = indexErr.Error()
	}

	payload, err := marshalPayload(rawPayload)
	if err != nil {
		return err
	}

	return c.doAndDrop(ctx, "POST", url, payload)
}

// Heartbeat hints to the index manager that the indexer system is has not been lost and should not
// release any of the index records assigned to the indexer.
func (c *Client) Heartbeat(ctx context.Context, indexIDs []int) error {
	url, err := makeIndexManagerURL(c.options.FrontendURL, c.options.FrontendAuthToken, c.options.Prefix, "heartbeat")
	if err != nil {
		return err
	}

	payload, err := marshalPayload(HeartbeatRequest{
		IndexerName: c.options.IndexerName,
		IndexIDs:    indexIDs,
	})
	if err != nil {
		return err
	}

	return c.doAndDrop(ctx, "POST", url, payload)
}

// doAndDrop performs an HTTP request to the frontend and ignores the body contents.
func (c *Client) doAndDrop(ctx context.Context, method string, url *url.URL, payload io.Reader) error {
	hasContent, body, err := c.do(ctx, method, url, payload)
	if err != nil {
		return err
	}
	if hasContent {
		body.Close()
	}
	return nil
}

// do performs an HTTP request to the frontend and returns the body content as a reader.
func (c *Client) do(ctx context.Context, method string, url *url.URL, body io.Reader) (hasContent bool, _ io.ReadCloser, err error) {
	span, ctx := ot.StartSpanFromContext(ctx, "do")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()

	req, err := http.NewRequest(method, url.String(), body)
	if err != nil {
		return false, nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	req = req.WithContext(ctx)

	req, ht := nethttp.TraceRequest(
		span.Tracer(),
		req,
		nethttp.OperationName(c.options.OperationName),
		nethttp.ClientTrace(false),
	)
	defer ht.Finish()

	resp, err := ctxhttp.Do(req.Context(), c.httpClient, req)
	if err != nil {
		return false, nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()

		if resp.StatusCode == http.StatusNoContent {
			return false, nil, nil
		}

		return false, nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	return true, resp.Body, nil
}

func makeIndexManagerURL(baseURL, authToken, prefix, op string) (*url.URL, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	base.User = url.UserPassword("indexer", authToken)

	return base.ResolveReference(&url.URL{Path: path.Join(prefix, op)}), nil
}

func marshalPayload(payload interface{}) (io.Reader, error) {
	content, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(content), nil
}
