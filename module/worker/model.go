package worker

import (
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"

	"github.com/gonejack/inostar-heroku/module/dropbox"
	"github.com/gonejack/inostar-heroku/util"
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

func (s *StreamItem) From(r io.Reader) (err error) {
	return json.NewDecoder(r).Decode(s)
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

func (it *Item) StarTime() time.Time {
	return time.UnixMicro(cast.ToInt64(it.TimestampUsec))
}
func (it *Item) Link() string {
	if len(it.Canonical) > 0 {
		return it.Canonical[0].Href
	}
	return it.Origin.HtmlUrl
}

const stateName = "inostar.json"

type State struct {
	LastStarTimeRaw string `json:"last_star_time"`
}

func (s *State) From(r io.Reader) {
	_ = json.NewDecoder(r).Decode(s)
}
func (s *State) String() string {
	return util.JsonDump(s)
}
func (s *State) SetLastStarTime(t time.Time) {
	s.LastStarTimeRaw = t.Format(time.RFC3339Nano)
}
func (s *State) LastStarTime() (t time.Time) {
	t, err := time.Parse(time.RFC3339Nano, s.LastStarTimeRaw)
	if s.LastStarTimeRaw == "" || err != nil {
		t = time.Now()
		s.SetLastStarTime(t)
	}
	return
}
func (s *State) Read() {
	_, body, err := dropbox.Read(stateName)
	if err != nil {
		if strings.Contains(err.Error(), "not_found") {
			logrus.Debugf("read %s failed: %s", stateName, err)
		} else {
			logrus.Errorf("read %s failed: %s", stateName, err)
		}
		return
	}
	defer body.Close()

	s.From(body)

	if logrus.IsLevelEnabled(logrus.DebugLevel) {
		logrus.Debugf("query state: %s", s.String())
	}
}
func (s *State) Save() {
	_, err := dropbox.Delete(stateName)
	if err != nil {
		if !strings.Contains(err.Error(), "not_found") {
			logrus.Errorf("remove %s failed: %s", stateName, err)
		}
	}

	if logrus.IsLevelEnabled(logrus.DebugLevel) {
		logrus.Debugf("save state: %s", s.String())
	}

	_, err = dropbox.Upload(stateName, 0, io.NopCloser(strings.NewReader(s.String())))
	if err != nil {
		logrus.Errorf("save %s failed: %s", stateName, err)
	}
}

type stack struct {
	arr []Item
}

func (s *stack) Len() int {
	return len(s.arr)
}
func (s *stack) Push(v Item) {
	s.arr = append(s.arr, v)
}
func (s *stack) Pop() (v Item, err error) {
	if n := len(s.arr); n == 0 {
		err = errors.New("stack is empty")
	} else {
		v, s.arr = s.arr[n-1], s.arr[:n-1]
	}
	return
}

func NewStack() *stack {
	return &stack{}
}
