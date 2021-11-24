package reader

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"

	"github.com/gonejack/inostar-heroku/config"
	"github.com/gonejack/inostar-heroku/model"
	"github.com/gonejack/inostar-heroku/reader/constant"
	"github.com/gonejack/inostar-heroku/util"
)

type reader struct {
	context context.Context
	cancel  context.CancelFunc
	client  *http.Client

	info Info

	sync.Mutex
	sync.WaitGroup
}

func (r *reader) reset(tok *oauth2.Token) {
	logrus.Infof("reset with token %s", util.JsonDump(tok))

	r.Lock()
	defer r.Unlock()

	if r.cancel != nil {
		r.cancel()
		r.Wait()
	}

	r.context, r.cancel = context.WithCancel(context.TODO())
	r.client = config.OAuth2.Client(r.context, tok)

	go func() {
		r.Add(1)
		logrus.Info("reader start")
		r.mainRoutine()
		logrus.Info("reader stop")
		r.Done()
	}()
}
func (r *reader) mainRoutine() {
	for {
		r.handle()
		select {
		case <-r.context.Done():
			return
		case <-time.After(time.Minute * 3):
		}
	}
}
func (r *reader) handle() {
	logrus.Infof("query stars")

	r.info.Query()

	var query = r.queryStarsParams()
	var stars []Item
loop:
	for {
		resp, err := r.queryStars(query)
		if err != nil {
			logrus.Errorf("get star failed: %s", err)
			break
		}

		for _, s := range resp.Items {
			if s.StarTime().After(r.info.LastStarTime()) {
				stars = append(stars, s)
				logrus.Infof("add [%s][%s]", s.Title, s.StarTime())
			} else {
				logrus.Debugf("skip [%s][%s]", s.Title, s.StarTime())
				break loop
			}
		}

		if resp.Continuation == "" {
			break
		} else {
			query.Set("c", resp.Continuation)
		}
	}

	var delay *time.Timer
	for i := len(stars) - 1; i >= 0; i-- {
		s := stars[i]
		a := &model.Article{
			FeedTitle:   s.Origin.Title,
			Title:       s.Title,
			StarTimeRaw: s.StarTime().UTC().Format("01/02/2006 3:04:05 PM"),
			Href:        s.Link(),
			Article:     s.Summary.Content,
		}
		_, err := util.SaveArticle(a)
		if err == nil {
			r.info.SetLastStarTime(s.StarTime())

			switch {
			case delay == nil:
				delay = time.AfterFunc(time.Second*20, func() { r.info.Save() })
			default:
				if !delay.Stop() {
					<-delay.C
				}
				delay.Reset(time.Second * 20)
			}
		} else {
			logrus.Errorf("save post %s failed: %s", a.Title, err)
		}
	}
}
func (r *reader) queryStars(query url.Values) (data StreamItem, err error) {
	api := constant.TagStarred.WithQuery(query)
	logrus.Debugf("request %s", api)

	rsp, err := r.client.Get(api)
	if err != nil {
		err = fmt.Errorf("send request failed: %w", err)
		return
	}
	defer rsp.Body.Close()

	err = data.From(rsp.Body)
	if err != nil {
		err = fmt.Errorf("decode failed: %w", err)
		return
	}

	return
}

// https://www.inoreader.com/developers/stream-contents
func (r *reader) queryStarsParams() (q url.Values) {
	q = make(url.Values)
	// n - Number of items to return (default 20, max 1000).
	q.Set("n", "50")
	// it - Include Target - You can query for a certain label with this.
	// Accepted values: user/-/state/com.google/starred, user/-/state/com.google/like.
	q.Set("it", constant.TagStarred.String())
	return q
}
