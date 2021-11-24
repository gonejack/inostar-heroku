package reader

import (
	"encoding/json"
	"io"
	"time"

	"github.com/spf13/cast"
)

type StreamItem struct {
	Direction   string `json:"direction"`
	Id          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Self        struct {
		Href string `json:"href"`
	} `json:"self"`
	Updated      int    `json:"updated"`
	UpdatedUsec  string `json:"updatedUsec"`
	Items        []Item `json:"items"`
	Continuation string `json:"continuation"`
}

func (d *StreamItem) From(r io.Reader) (err error) {
	return json.NewDecoder(r).Decode(d)
}

type Item struct {
	CrawlTimeMsec string   `json:"crawlTimeMsec"`
	TimestampUsec string   `json:"timestampUsec"`
	Id            string   `json:"id"`
	Categories    []string `json:"categories"`
	Title         string   `json:"title"`
	Published     int      `json:"published"`
	Updated       int      `json:"updated"`
	Canonical     []struct {
		Href string `json:"href"`
	} `json:"canonical"`
	Alternate []struct {
		Href string `json:"href"`
		Type string `json:"type"`
	} `json:"alternate"`
	Summary struct {
		Direction string `json:"direction"`
		Content   string `json:"content"`
	} `json:"summary"`
	Author      string        `json:"author"`
	LikingUsers []interface{} `json:"likingUsers"`
	Comments    []interface{} `json:"comments"`
	CommentsNum int           `json:"commentsNum"`
	Annotations []struct {
		Id                 int    `json:"id"`
		Start              int    `json:"start"`
		End                int    `json:"end"`
		AddedOn            int    `json:"added_on"`
		Text               string `json:"text"`
		Note               string `json:"note"`
		UserId             int    `json:"user_id"`
		UserName           string `json:"user_name"`
		UserProfilePicture string `json:"user_profile_picture"`
	} `json:"annotations"`
	Origin struct {
		StreamId string `json:"streamId"`
		Title    string `json:"title"`
		HtmlUrl  string `json:"htmlUrl"`
	} `json:"origin"`
}

func (i *Item) StarTime() time.Time {
	return time.UnixMicro(cast.ToInt64(i.TimestampUsec)).Round(time.Second)
}

func (i *Item) Link() string {
	if len(i.Canonical) > 0 {
		return i.Canonical[0].Href
	}
	return i.Origin.HtmlUrl
}
