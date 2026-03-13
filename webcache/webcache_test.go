package webcache_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/chaimleib/weatherfmt/typedmap"
	"github.com/chaimleib/weatherfmt/webcache"
)

func TestNew(t *testing.T) {
	w, err := webcache.New()
	if err != nil {
		t.Errorf("unexpected err: %v", err)
	}
	if w == nil {
		t.Errorf("unexpected nil from New")
	}
}

type ClientResult struct {
	R   *http.Response
	Err error
}

type MapClient map[string]ClientResult

var _ webcache.HTTPClient = MapClient(nil)

func (c MapClient) Do(req *http.Request) (*http.Response, error) {
	u := req.URL.String()
	res, ok := c[u]
	if !ok {
		return nil, fmt.Errorf("test: ClientResult not configured for %s", u)
	}
	return res.R, res.Err
}

func NewBody(s string) io.ReadCloser {
	return io.NopCloser(strings.NewReader(s))
}

type ErrorBody struct{}

var _ io.ReadCloser = (*ErrorBody)(nil)

func (b ErrorBody) Read([]byte) (int, error) {
	return 0, errors.New("read error")
}

func (b ErrorBody) Close() error { return nil }

type HookCache struct {
	*typedmap.TypedMap[webcache.CacheEntry]
	loadHook func(
		u string,
		load func(string) (webcache.CacheEntry, bool),
	) (webcache.CacheEntry, bool)
}

func (m *HookCache) Load(u string) (webcache.CacheEntry, bool) {
	return m.loadHook(u, m.TypedMap.Load)
}

func TestWebCache(t *testing.T) {
	type Case struct {
		name          string
		cache         webcache.SwapMap[webcache.CacheEntry]
		method        string
		doErr         string
		doStatus      int
		respBody      io.ReadCloser
		getUpdateCtx  func(ctx context.Context) context.Context // defaults to t.Context()
		getGetCtx     func(ctx context.Context) context.Context // defaults to t.Context()
		skipRegister  bool
		wantUpdateErr string
		wantGetErr    string
		wantData      string
	}
	cases := []Case{
		{
			name:     "success",
			respBody: NewBody("hi"),
			wantData: "hi",
		},
		{
			name:          "fetch fail",
			doErr:         "GET failed",
			wantUpdateErr: "GET failed",
			wantGetErr:    "GET failed",
		},
		{
			name:          "fetch status 400",
			respBody:      NewBody("hi"),
			doStatus:      http.StatusBadRequest,
			wantUpdateErr: "unexpected status from GET https://example.com/doc: 400 Bad Request",
			wantGetErr:    "unexpected status from GET https://example.com/doc: 400 Bad Request",
		},
		{
			name:          "invalid method",
			method:        ":INVALID",
			wantUpdateErr: `net/http: invalid method ":INVALID"`,
			wantGetErr:    `net/http: invalid method ":INVALID"`,
		},
		{
			name:          "read body failure",
			respBody:      ErrorBody{},
			wantUpdateErr: "read error",
			wantGetErr:    "read error",
		},
		{
			name:          "unregistered fail",
			respBody:      NewBody("hi"),
			skipRegister:  true,
			wantUpdateErr: `URL has no expiration set: "https://example.com/doc"`,
			wantGetErr:    `URL has no expiration set: "https://example.com/doc"`,
		},
		{
			name: "done update context fail",
			getUpdateCtx: func(ctx context.Context) context.Context {
				ctx, cancel := context.WithCancel(ctx)
				cancel()
				return ctx
			},
			respBody:      NewBody("hi"),
			wantUpdateErr: "context canceled",
			wantData:      "hi",
		},
		{
			name: "done get context fail",
			getGetCtx: func(ctx context.Context) context.Context {
				ctx, cancel := context.WithCancel(ctx)
				cancel()
				return ctx
			},
			respBody:   NewBody("hi"),
			wantGetErr: "context canceled",
		},
		func() (c Case) {
			c.name = "canceled get context fail before web fetch"

			// Create a cancelable updateCtx.
			var updateCtx context.Context
			var cancel context.CancelFunc
			c.getUpdateCtx = func(ctx context.Context) context.Context {
				updateCtx, cancel = context.WithCancel(ctx)
				return updateCtx
			}

			// Set up the cache to cancel updateCtx after the first Load() call.
			c.cache = &HookCache{
				TypedMap: typedmap.New[webcache.CacheEntry](),
				loadHook: func(
					u string,
					load func(string) (webcache.CacheEntry, bool),
				) (webcache.CacheEntry, bool) {
					val, ok := load(u)
					if updateCtx.Err() == nil { // if not canceled yet
						cancel()
					}
					return val, ok
				},
			}

			c.respBody = NewBody("hi")
			c.wantUpdateErr = "context canceled"
			c.wantData = "hi"
			return c
		}(),
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// Create a WebCache.
			w, err := webcache.New()
			if err != nil {
				t.Fatal(err)
			}

			// Set up WebCache mocks.
			nowFunc := func() time.Time {
				return time.Date(2026, time.May, 18, 13, 0, 0, 0, time.UTC)
			}
			w.Now = nowFunc
			client := make(MapClient)
			w.HTTPClient = client
			if c.cache != nil {
				w.Cache = c.cache
			}

			// Add a test document.
			u := "https://example.com/doc"
			if !c.skipRegister {
				w.SetExpiration(u, time.Hour)
			}
			if c.method != "" {
				err := w.Docs.SwapFunc(u, func(
					u string,
					prev webcache.Document,
					ok bool,
				) (webcache.Document, error) {
					prev.Method = c.method
					return prev, nil
				})
				if err != nil {
					t.Errorf("unexpected error while setting up method: %v", err)
				}
			}

			if c.doErr != "" {
				client[u] = ClientResult{Err: errors.New(c.doErr)}
			} else {
				status := http.StatusOK
				if c.doStatus != 0 {
					status = c.doStatus
				}
				client[u] = ClientResult{R: &http.Response{
					Status:     http.StatusText(status),
					StatusCode: status,
					Body:       c.respBody,
				}}
			}

			// Set up the context.
			updateCtx := t.Context()
			if c.getUpdateCtx != nil {
				updateCtx = c.getUpdateCtx(updateCtx)
			}

			// Fetch initially.
			err = w.Update(updateCtx, u, nil)
			// Check update error.
			if err != nil {
				if c.wantUpdateErr == "" {
					t.Errorf("update err should have been nil, got:\n%v", err)
				} else if err.Error() != c.wantUpdateErr {
					t.Errorf("update err did not match, want:\n%s\ngot:\n%v",
						c.wantUpdateErr, err)
				}
			} else if c.wantUpdateErr != "" {
				t.Errorf("update err was nil, want:\n%s", c.wantUpdateErr)
			}

			// Set up the context.
			getCtx := t.Context()
			if c.getGetCtx != nil {
				getCtx = c.getGetCtx(getCtx)
			}

			// Test the Get function for a cache hit.
			data, err := w.Get(getCtx, u)
			// Check get error.
			if err != nil {
				if c.wantGetErr == "" {
					t.Errorf("get err should have been nil, got:\n%v", err)
				} else if err.Error() != c.wantGetErr {
					t.Errorf("get err did not match, want:\n%s\ngot:\n%v",
						c.wantGetErr, err)
				}
			} else if c.wantGetErr != "" {
				t.Errorf("get err was nil, want:\n%s", c.wantGetErr)
			}
			// Check data.
			if string(data) != c.wantData {
				t.Errorf("fetched data did not match, want:\n%s\ngot:\n%s",
					c.wantData, string(data))
			}
		})
	}
}
