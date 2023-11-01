package webcache

import "net/http"

type HTTPCache interface {
	Get(r *http.Request) (*http.Response, bool)
	Set(r *http.Request, response *http.Response)
	Delete(r *http.Request)
}

type httpCache struct {
	cache Cache[cacheKey, http.Response]
}

func NewHTTPCache(cache Cache[cacheKey, http.Response]) HTTPCache {
	return &httpCache{cache: cache}
}

func (c *httpCache) Get(r *http.Request) (*http.Response, bool) {
	cacheKey := buildCacheKey(r)
	return c.cache.Get(cacheKey)
}

func (c *httpCache) Set(r *http.Request, response *http.Response) {
	cacheKey := buildCacheKey(r)
	c.cache.Set(cacheKey, response)
}

func (c *httpCache) Delete(r *http.Request) {
	cacheKey := buildCacheKey(r)
	c.cache.Delete(cacheKey)
}

func isCached(r *http.Response) bool {
	return r.Header.Get("X-Cache") == "HIT"
}
