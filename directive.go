package cache

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/textproto"
	"strconv"
	"strings"
)

var (
	ErrMaxAgeDeltaSeconds               = errors.New("invalid delta-seconds value in `max-age` directive")
	ErrSMaxAgeDeltaSeconds              = errors.New("invalid delta-seconds value in `s-maxage` directive")
	ErrMaxStaleDeltaSeconds             = errors.New("invalid delta-seconds value in `max-stale` directive")
	ErrMinFreshDeltaSeconds             = errors.New("invalid delta-seconds value in `min-fresh` directive")
	ErrStaleIfErrorDeltaSeconds         = errors.New("invalid delta-seconds value in `stale-if-error` directive")
	ErrStaleWhileRevalidateDeltaSeconds = errors.New("invalid delta-seconds value in `stale-while-revalidate` directive")

	ErrPublicDirectiveValue          = errors.New("public directive does not accept a value")
	ErrNoCacheDirectiveValue         = errors.New("no-cache directive does not accept a value")
	ErrNoStoreDirectiveValue         = errors.New("no-store directive does not accept a value")
	ErrImmutableDirectiveValue       = errors.New("immutable directive does not accept a value")
	ErrNoTransformDirectiveValue     = errors.New("no-transform directive does not accept a value")
	ErrOnlyIfCachedDirectiveValue    = errors.New("only-if-cached directive does not accept a value")
	ErrMustRevalidateDirectiveValue  = errors.New("must-revalidate directive does not accept a value")
	ErrProxyRevalidateDirectiveValue = errors.New("proxy-revalidate directive does not accept a value")
)

type directive interface {
	setToken(token string) error
	setPair(key, val string) error
}

func NewRequestCacheDirective(value string) (*RequestCacheDirective, error) {
	directive := &RequestCacheDirective{MaxAge: -1, MaxStale: -1, MinFresh: -1}
	if err := parseCacheControlv(directive, value); err != nil {
		return nil, err
	}
	return directive, nil
}

type RequestCacheDirective struct {
	// max-age
	// MaxAge is the maximum time in seconds that a response can be considered fresh.
	// The client is not willing to accept a stale response if the max-stale directive
	// is not present.
	MaxAge int32

	// max-stale
	// MaxStale is the maximum time in seconds that a client is willing to accept a
	// stale response that has exceeded its freshness lifetime.
	MaxStale int32

	// min-fresh
	// MinFresh is the minimum time in seconds that a response must remain fresh,
	// calculated as the difference between its freshness lifetime and its current age.
	MinFresh int32

	// no-cache
	// NoCache is a boolean value that indicates whether a cache should validate
	// the response with the origin server before using a stored response to satisfy
	// the request.
	NoCache bool

	// no-store
	// NoStore is a boolean value that indicates whether a cache should store
	// any part of the request or response.
	NoStore bool

	// only-if-cached
	// OnlyIfCached is a boolean value that indicates whether the client only
	// wants to obtain a stored response and not send a request to the origin server.
	OnlyIfCached bool

	// Extensions is a list of cache-extension tokens with optional values
	// that can be used to extend the Cache-Control header field.
	Extensions []string
}

func (directive *RequestCacheDirective) setToken(token string) error {
	switch token {
	case HeaderMaxAge:
		return ErrMaxAgeDeltaSeconds
	case HeaderMaxStale:
		return ErrMaxStaleDeltaSeconds
	case HeaderMinFresh:
		return ErrMinFreshDeltaSeconds
	}

	switch token {
	case HeaderNoCache:
		directive.NoCache = true
	case HeaderNoStore:
		directive.NoStore = true
	case HeaderOnlyIfCached:
		directive.OnlyIfCached = true
	default:
		directive.Extensions = append(directive.Extensions, token)
	}
	return nil
}

func (directive *RequestCacheDirective) setPair(key, val string) error {
	switch key {
	case HeaderNoCache:
		return ErrNoCacheDirectiveValue
	case HeaderNoStore:
		return ErrNoStoreDirectiveValue
	case HeaderOnlyIfCached:
		return ErrOnlyIfCachedDirectiveValue
	}

	switch key {
	case HeaderMaxAge:
		deltaSec, err := validateDeltaSeconds(val)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrMaxAgeDeltaSeconds, err)
		}
		directive.MaxAge = deltaSec
	case HeaderMaxStale:
		deltaSec, err := validateDeltaSeconds(val)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrMaxStaleDeltaSeconds, err)
		}
		directive.MaxStale = deltaSec
	case HeaderMinFresh:
		deltaSec, err := validateDeltaSeconds(val)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrMinFreshDeltaSeconds, err)
		}
		directive.MinFresh = deltaSec
	default:
		directive.Extensions = append(directive.Extensions, key+"="+val)
	}
	return nil
}

func NewResponseCacheDirective(value string) (*ResponseCacheDirective, error) {
	directive := &ResponseCacheDirective{}
	if err := parseCacheControlv(directive, value); err != nil {
		return nil, err
	}
	return directive, nil
}

type ResponseCacheDirective struct {
	// MustRevalidate is a boolean value that indicates whether a cache
	// must revalidate a stored response on every request.
	MustRevalidate bool

	// NoCache is a map of field-name to boolean values that indicates whether
	// a cache should not use a stored response to satisfy a request if any of the
	// request-header field names are present in the list.
	NoCache map[string]bool

	// NoCachePresent is a boolean value that indicates whether the no-cache
	// directive was present in the response.
	NoCachePresent bool

	// NoStore is a boolean value that indicates whether a cache should not
	// store any part of the response or request.
	NoStore bool

	// NoTransform is a boolean value that indicates whether a cache should
	// not transform the response payload.
	NoTransform bool

	// Public is a boolean value that indicates whether the response is
	// considered public, meaning it can be cached by any cache.
	Public bool

	// Private is a map of field-name to boolean values that indicates whether
	// the response is considered private, meaning it can only be cached by a cache
	// that is specific to a particular user.
	Private map[string]bool

	// PrivatePresent is a boolean value that indicates whether the private
	// directive was present in the response.
	PrivatePresent bool

	// ProxyRevalidate is a boolean value that indicates whether a cache must
	// revalidate a stored response on every request when the response was obtained
	// from a proxy cache.
	ProxyRevalidate bool

	// MaxAge is the maximum time in seconds that a response can be considered fresh.
	MaxAge int32

	// SMaxAge is the maximum time in seconds that a shared cache can consider
	// a response to be fresh.
	SMaxAge int32

	// Immutable is a boolean value that indicates whether the response payload
	// is considered immutable and can be cached indefinitely.
	Immutable bool

	// StaleIfError is the maximum time in seconds that a cache can serve a stale
	// response when an error occurs.
	StaleIfError int32

	// StaleWhileRevalidate is the maximum time in seconds that a cache can serve
	// a stale response while a background revalidation is being performed.
	StaleWhileRevalidate int32

	// Extensions is a list of cache-extension tokens with optional values
	// that can be used to extend the Cache-Control header field.
	Extensions []string
}

func (directive *ResponseCacheDirective) setToken(token string) error {
	switch token {
	case HeaderMaxAge:
		return ErrMaxAgeDeltaSeconds
	case HeaderSMaxAge:
		return ErrSMaxAgeDeltaSeconds
	case HeaderStaleIfError:
		return ErrStaleIfErrorDeltaSeconds
	case HeaderStaleWhileRevalidate:
		return ErrStaleWhileRevalidateDeltaSeconds
	}

	switch token {
	case HeaderPublic:
		directive.Public = true
	case HeaderNoStore:
		directive.NoStore = true
	case HeaderImmutable:
		directive.Immutable = true
	case HeaderNoTransform:
		directive.NoTransform = true
	case HeaderPrivate:
		directive.PrivatePresent = true
	case HeaderMustRevalidate:
		directive.MustRevalidate = true
	case HeaderNoCache:
		directive.NoCachePresent = true
	case HeaderProxyRevalidate:
		directive.ProxyRevalidate = true
	default:
		directive.Extensions = append(directive.Extensions, token)
	}
	return nil
}

func (directive *ResponseCacheDirective) setPair(key, val string) error {
	switch key {
	case HeaderMustRevalidate:
		return ErrMustRevalidateDirectiveValue
	case HeaderNoStore:
		return ErrNoStoreDirectiveValue
	case HeaderNoTransform:
		return ErrNoTransformDirectiveValue
	case HeaderPublic:
		return ErrPublicDirectiveValue
	case HeaderProxyRevalidate:
		return ErrProxyRevalidateDirectiveValue
	case HeaderImmutable:
		return ErrImmutableDirectiveValue
	}

	switch key {
	case HeaderNoCache:
		directive.NoCachePresent = true

		if directive.NoCache == nil {
			directive.NoCache = make(map[string]bool)
		}

		vals := strings.Split(val, ",")
		for _, v := range vals {
			k := http.CanonicalHeaderKey(textproto.TrimString(v))
			directive.NoCache[k] = true
		}
	case HeaderPrivate:
		directive.PrivatePresent = true

		if directive.Private == nil {
			directive.Private = make(map[string]bool)
		}

		vals := strings.Split(val, ",")
		for _, v := range vals {
			k := http.CanonicalHeaderKey(textproto.TrimString(v))
			directive.Private[k] = true
		}
	case HeaderMaxAge:
		deltaSec, err := validateDeltaSeconds(val)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrMaxAgeDeltaSeconds, err)
		}
		directive.MaxAge = deltaSec
	case HeaderSMaxAge:
		deltaSec, err := validateDeltaSeconds(val)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrSMaxAgeDeltaSeconds, err)
		}
		directive.SMaxAge = deltaSec
	case HeaderStaleIfError:
		deltaSec, err := validateDeltaSeconds(val)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrStaleIfErrorDeltaSeconds, err)
		}
		directive.StaleIfError = deltaSec
	case HeaderStaleWhileRevalidate:
		deltaSec, err := validateDeltaSeconds(val)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrStaleWhileRevalidateDeltaSeconds, err)
		}
		directive.StaleWhileRevalidate = deltaSec
	default:
		directive.Extensions = append(directive.Extensions, key+"="+val)
	}

	return nil
}

func validateDeltaSeconds(delta string) (int32, error) {
	deltaSec, err := strconv.ParseUint(delta, 10, 32)
	if err != nil {
		switch err.(*strconv.NumError).Err {
		case strconv.ErrRange:
			return math.MaxInt32, nil
		default:
			return -1, err
		}
	}

	if deltaSec > math.MaxInt32 {
		return math.MaxInt32, nil
	}
	return int32(deltaSec), nil
}
