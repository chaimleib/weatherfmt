// Package webcache provides the ability
// to download and cache documents from the internet.
package webcache

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/chaimleib/weatherfmt/typedmap"
)

type SwapMap[T any] interface {
	Load(key string) (T, bool)
	Store(key string, val T)
	SwapFunc(
		key string,
		f func(key string, prev T, ok bool) (T, error),
	) error
	Keys() []string
}

// CacheEntry stores the raw XML data and its fetch timestamp.
type CacheEntry struct {
	Data      []byte
	FetchedAt time.Time
}

// Document configures an expiration for a URL.
type Document struct {
	// Method is the HTTP method to use. Defaults to http.MethodGet.
	Method string

	// Body is the HTTP request body to use. Defaults to none.
	Body io.Reader

	// ExpireIn indicates how long a downloaded document can be
	// considered unexpired.
	// Use 0 to indicate that it never expires.
	// (If you never want to cache this document,
	// don't use Agent.
	// Agent exists to cache downloads.)
	ExpireIn time.Duration
}

type UpdateOptions struct {
	// IgnoreExpirations tells update functions to always fetch the sources,
	// even if they have not expired yet.
	IgnoreExpirations bool
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// WebCache is a cache for web documents.
type WebCache struct {
	// Now is a function for fetching the current time.
	// The New function defaults this to time.Now.
	Now func() time.Time

	// HTTPClient provides a Do method.
	// The New function defaults this to an empty http.Client.
	HTTPClient HTTPClient

	// Docs configures URLs with expirations and other fetch options.
	Docs SwapMap[Document]

	// Cache stores document bodies fetched from the web.
	Cache SwapMap[CacheEntry]
}

func New() (*WebCache, error) {
	w := new(WebCache)
	w.Now = time.Now
	w.HTTPClient = new(http.Client)
	w.Cache = typedmap.New[CacheEntry]()
	w.Docs = typedmap.New[Document]()
	return w, nil
}

func (w *WebCache) SetExpiration(u string, exp time.Duration) error {
	return w.Docs.SwapFunc(
		u,
		func(u string, d Document, ok bool) (Document, error) {
			d.ExpireIn = exp
			return d, nil
		},
	)
}

func (w *WebCache) UpdateAll(ctx context.Context, opts *UpdateOptions) error {
	var defaultOpts UpdateOptions
	if opts == nil {
		opts = &defaultOpts
	}

	cancel := func() {}
	ctx, cancel = context.WithCancel(ctx)
	defer cancel()

	errChan := make(chan error)
	var wg sync.WaitGroup

	urls := w.Docs.Keys()
	for _, u := range urls {
		wg.Go(func() {
			if err := w.Update(ctx, u, opts); err != nil {
				cancel()
				errChan <- fmt.Errorf("%s: %w", u, err)
				return
			}
		})
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) != 0 {
		return errors.Join(errs...)
	}
	return nil
}

func (w *WebCache) Update(
	ctx context.Context,
	u string,
	opts *UpdateOptions,
) error {
	var defaultOpts UpdateOptions
	if opts == nil {
		opts = &defaultOpts
	}

	// Check ctx for cancel.
	if err := ctx.Err(); err != nil {
		return err
	}

	// Fetch the Document config.
	d, ok := w.Docs.Load(u)
	if !ok {
		return fmt.Errorf("URL has no expiration set: %q", u)
	}

	// When was is previously cached?
	cache, ok := w.Cache.Load(u)

	// Skip fetching from the web if not expired or IgnoreExpirations.
	if !opts.IgnoreExpirations &&
		!cache.FetchedAt.IsZero() &&
		w.Now().Sub(cache.FetchedAt) < d.ExpireIn {

		return nil
	}

	// Check ctx for cancel.
	if err := ctx.Err(); err != nil {
		return err
	}

	// Fetch from the web.
	method := d.Method
	if method == "" {
		method = http.MethodGet
	}
	rq, err := http.NewRequestWithContext(
		ctx, method, u, d.Body)
	if err != nil {
		return err
	}

	resp, err := w.HTTPClient.Do(rq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status from %s %s: %d %s",
			method, u, resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Write a new CacheEntry.
	ts := w.Now()
	w.Cache.Store(u, CacheEntry{
		Data:      body,
		FetchedAt: ts,
	})
	return nil
}

func (w *WebCache) Get(
	ctx context.Context,
	u string,
) ([]byte, error) {
	err := w.Update(ctx, u, nil)
	if err != nil {
		return nil, err
	}
	cache, _ := w.Cache.Load(u)
	return cache.Data, nil
}
