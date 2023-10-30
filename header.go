package webcache

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	ErrorInvalidMaxAge  = errors.New("invalid max-age")
	ErrorMaxAgeNotFound = errors.New("max-age not found")

	ErrorInvalidAge  = errors.New("invalid age")
	ErrorAgeNotFound = errors.New("age not found")

	ErrorInvalidResponseDate = errors.New("invalid response date")
	ErrorInvalidExpireDate   = errors.New("invalid expire date")
)

type cacheControlKey string

var (
	cacheControlKeyMaxAge  = cacheControlKey("max-age")
	cacheControlKeyPublic  = cacheControlKey("public")
	cacheControlKeyPrivate = cacheControlKey("private")
)

type CacheControl map[cacheControlKey]string

func (c CacheControl) MaxAge() (int, error) {
	v, ok := c[cacheControlKeyMaxAge]
	if !ok {
		return 0, ErrorMaxAgeNotFound
	}
	maxAge, err := strconv.Atoi(v)
	if err != nil {
		return 0, ErrorInvalidMaxAge
	}
	return maxAge, nil
}

func (c CacheControl) Public() bool {
	_, ok := c[cacheControlKeyPublic]
	return ok
}

func (c CacheControl) Private() bool {
	_, ok := c[cacheControlKeyPrivate]
	return ok
}

func ageFromHeader(h http.Header) (int, error) {
	age, err := strconv.Atoi(h.Get("Age"))
	if err != nil {
		return 0, ErrorInvalidMaxAge
	}
	return age, nil
}

func expiresFromHeader(h http.Header) (time.Time, error) {
	v, err := timeFromHeader(h, "Expires")
	if err != nil {
		return time.Time{}, ErrorInvalidExpireDate
	}
	return v, nil
}

func dateFromHeader(h http.Header) (time.Time, error) {
	v, err := timeFromHeader(h, "Date")
	if err != nil {
		return time.Time{}, ErrorInvalidResponseDate
	}
	return v, nil
}

func timeFromHeader(h http.Header, key string) (time.Time, error) {
	v, err := http.ParseTime(h.Get(key))
	if err != nil {
		return time.Time{}, err
	}
	return v, nil
}

func newCacheControl(h http.Header) CacheControl {
	cc := CacheControl{}
	for k, v := range h {
		if k != "Cache-Control" {
			continue
		}
		for _, vv := range v {
			for _, vvv := range splitCacheControl(vv) {
				kv := splitCacheControlKeyValue(vvv)
				if len(kv) == 2 {
					cc[cacheControlKey(kv[0])] = kv[1]
				}
				if len(kv) == 1 {
					cc[cacheControlKey(kv[0])] = ""
				}
			}
		}
	}
	return cc
}
func splitCacheControl(s string) []string {
	return strings.Split(strings.TrimSpace(s), ",")
}

func splitCacheControlKeyValue(s string) []string {
	return strings.Split(strings.TrimSpace(s), "=")
}
