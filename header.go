package cache

// Cacheable HTTP header directives
const (
	HeaderMaxAge       = "max-age"
	HeaderNoCache      = "no-cache"
	HeaderNoStore      = "no-store"
	HeaderMaxStale     = "max-stale"
	HeaderMinFresh     = "min-fresh"
	HeaderNoTransform  = "no-transform"
	HeaderOnlyIfCached = "only-if-cached"

	HeaderPublic               = "public"
	HeaderPrivate              = "private"
	HeaderSMaxAge              = "s-maxage"
	HeaderImmutable            = "immutable"
	HeaderStaleIfError         = "stale-if-error"
	HeaderMustRevalidate       = "must-revalidate"
	HeaderProxyRevalidate      = "proxy-revalidate"
	HeaderStaleWhileRevalidate = "stale-while-revalidate"
)
