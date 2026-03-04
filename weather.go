// Package weatherfmt provides the ability
// to download minute weather forecasts for a given location
// so that the data can be inserted into the output.
//
// It uses XML data from the US National Weather Service.
// To avoid excess network usage, it caches the results every hour.
package weatherfmt

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const urlTemplate = "https://forecast.weather.gov/MapClick.php?lat=%f&lon=%f&FcstType=digitalDWML"

// CacheEntry stores the raw XML data and its fetch timestamp.
type CacheEntry struct {
    Data   []byte // Raw XML content
    FetchedAt time.Time
}

type Weather struct {
	http *http.Client
	timeout time.Duration
	places map[string]*url.URL
	cache map[string]CacheEntry
	// mu protects both maps, but nothing else.
	mu sync.Mutex
}

func New(lat, lon float64) (*Weather, error) {
	w := new(Weather)
	w.http = new(http.Client)
	w.cache = make(map[string]CacheEntry)
	w.places = make(map[string]*url.URL)
	return w, nil
}

func (w *Weather) AddPlace(name string, lat, lon float64) error {
	u, err := url.Parse(fmt.Sprintf(urlTemplate, lat, lon))
	if err != nil {
		return fmt.Errorf(
			"error: AddPlace failed to create update URL: %w",
			err,
		)
	}

	w.mu.Lock()
	w.places[name] = u
	w.mu.Unlock()

	return nil
}

func (w *Weather) UpdateAll(ctx context.Context) error {
	cancel := func() {}
	if w.timeout != 0 {
		ctx, cancel = context.WithTimeout(ctx, w.timeout)
	}
	defer cancel()

	errChan := make(chan error)
	var wg sync.WaitGroup

	w.mu.Lock()
	nameURLs := make(map[string]string, len(w.places))
	for name, u := range w.places {
		nameURLs[name] = u.String()
	}
	w.mu.Unlock()

	for name := range nameURLs {
		wg.Go(func() {
			if err := w.Update(ctx, name); err != nil {
				cancel()
				errChan <- fmt.Errorf("%s: %w", name, err)
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

func (w *Weather) Update(ctx context.Context, name string) error {
	w.mu.Lock()
	u := w.places[name].String()
	w.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	rq, err := http.NewRequestWithContext(
		ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}

	resp, err := w.http.Do(rq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	ts := time.Now()
	w.mu.Lock()
	w.cache[name] = CacheEntry{
		Data: body,
		FetchedAt: ts,
	}
	w.mu.Unlock()
	return nil
}

