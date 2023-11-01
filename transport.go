package webcache

import (
	"context"
	"net/http"
)

type Transport struct {
	clock            Clock
	cache            HTTPCache
	rt               http.RoundTripper
	freshnessChecker freshnessChecker
}

type TransportOption func(*Transport)

func WithClock(c Clock) TransportOption {
	return func(t *Transport) {
		t.clock = c
	}
}

// NewRoundTripper
func NewTransport(cache Cache[string, []byte], rt http.RoundTripper, opts ...TransportOption) *Transport {
	t := &Transport{
		cache: NewHTTPCache(cache),
		rt:    rt,
		clock: NewClock(),
	}
	for _, o := range opts {
		o(t)
	}
	t.freshnessChecker = newFreshnerChecker(t.clock)
	return t
}

func (t *Transport) RoundTrip(r *http.Request) (*http.Response, error) {
	// check if we have this request in the cache
	ctx := r.Context()
	response, ok := t.cache.Get(r)
	if ok {
		return t.handleCachedResponse(ctx, response, r)
	}

	return nil, nil
}

func (t *Transport) handleCachedResponse(ctx context.Context, response *http.Response, r *http.Request) (*http.Response, error) {
	cacheControl := newCacheControl(response.Header)

	// we check if the response is still fresh, if it is, we return it
	freshness, err := t.freshnessChecker.Freshness(ctx, response.Header, cacheControl)
	if err != nil {
		return nil, err
	}

	switch freshness {
	case FreshnessFresh:
		response.Header = withCacheHitHeader(response.Header)
		return response, nil

	case FreshnessStale:
		// if the response is stale, we check if we can validate it
		validator := newResponseValidator(t.rt)
		response, err := validator.Validate(response, r)
		if err != nil {
			return nil, err
		}

		// if caching is not allowed, we delete the response from the cache
		if cacheControl.NoStore() {
			t.cache.Delete(r)
			return response, nil
		}

		// if the validator returned a cached response, we return it
		if isCached(response) {
			return response, nil
		}

		// otherwise, we cache the response and return it
		t.cache.Set(r, response)
		return response, nil

	default:
		return t.rt.RoundTrip(r)
	}
}
