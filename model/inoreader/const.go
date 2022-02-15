package inoreader

import (
	"net/url"
	"path"
)

const Api builder = "https://www.inoreader.com"
const TagRead tag = "user/-/state/com.google/read"
const TagStarred tag = "user/-/state/com.google/starred"

type tag string

func (t tag) String() string {
	return string(t)
}

type builder string

func (b builder) String() string {
	return string(b)
}
func (b builder) StreamApi(tag tag, query url.Values) string {
	u, _ := url.Parse(b.String())
	u.Path = path.Join("/reader/api/0/stream/contents", tag.String())
	u.RawQuery = query.Encode()
	return u.String()
}
