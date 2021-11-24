package constant

import (
	"net/url"
	"path"
)

const (
	Host   = "https://www.inoreader.com"
	prefix = "/reader/api/0/stream/contents"

	TagRead    Endpoint = "user/-/state/com.google/read"
	TagStarred Endpoint = "user/-/state/com.google/starred"
)

type Endpoint string

func (e Endpoint) WithQuery(query url.Values) string {
	u, _ := url.Parse(Host)
	u.Path = path.Join(prefix, e.String())
	u.RawQuery = query.Encode()
	return u.String()
}

func (e Endpoint) String() string {
	return string(e)
}
