package worker

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"

	"github.com/gonejack/inostar-heroku/config"
	"github.com/gonejack/inostar-heroku/model"
	"github.com/gonejack/inostar-heroku/model/inoreader"
	"github.com/gonejack/inostar-heroku/module/kit"
	"github.com/gonejack/inostar-heroku/util"
)

type worker struct {
	state   State
	client  *http.Client
	context context.Context
	cancel  context.CancelFunc

	sync.Mutex
	sync.WaitGroup
}

func (w *worker) reset(tok *oauth2.Token) {
	logrus.Infof("reset with token %s", util.JsonDump(tok))

	w.Lock()
	defer w.Unlock()

	if w.cancel != nil {
		w.cancel()
		w.Wait()
	}

	w.context, w.cancel = context.WithCancel(context.TODO())
	w.client = config.OAuth2.Client(w.context, tok)

	go func() {
		w.Add(1)
		logrus.Info("worker start")
		w.mainRoutine()
		logrus.Info("worker stop")
		w.Done()
	}()
}
func (w *worker) mainRoutine() {
	for {
		w.handle()
		select {
		case <-w.context.Done():
			return
		case <-time.After(time.Minute * 3):
			continue
		}
	}
}
func (w *worker) handle() {
	logrus.Infof("query stars")
	stars := w.fetchStars()
	if stars.Len() > 0 {
		defer w.state.WriteOut()
	}
	for stars.Len() > 0 {
		star, _ := stars.Pop()
		a := &model.Article{
			FeedTitle:   star.Origin.Title,
			Title:       star.Title,
			StarTimeRaw: star.StarTime().UTC().Format("01/02/2006 3:04:05 PM"),
			Href:        star.Link(),
			Article:     star.Summary.Content,
		}
		a.Refine()
		_, err := kit.SaveAsEmail(a)
		switch {
		case err == nil:
			w.state.SetLastStarTime(star.StarTime())
		case strings.Contains(err.Error(), "conflict"):
			logrus.Warnf("save post %s failed: %s", a.Title, err)
		default:
			logrus.Errorf("save post %s as email failed: %s", a.Title, err)
			_, err = kit.SaveAsHTML(a)
			if err != nil {
				logrus.Errorf("save post %s as html failed: %s", a.Title, err)
			}
			return
		}
	}
}
func (w *worker) fetchStars() (stars *stack) {
	w.state.ReadIn()

	stars = NewStack()
	query := w.queryStarsParams()
	for {
		rsp, err := w.queryStars(query)
		if err != nil {
			logrus.Errorf("query star failed: %s", err)
			return
		}
		for _, star := range rsp.Items {
			if star.StarTime().After(w.state.LastStarTime()) {
				stars.Push(star)
				logrus.Infof("push [%s][%s]", star.Title, star.StarTime())
			} else {
				logrus.Debugf("skip [%s][%s]", star.Title, star.StarTime())
				return
			}
		}
		if rsp.Continuation == "" {
			return
		} else {
			query.Set("c", rsp.Continuation)
		}
	}
}
func (w *worker) queryStars(query url.Values) (data StreamItem, err error) {
	api := inoreader.Api.StreamApi(inoreader.TagStarred, query)
	logrus.Debugf("request %s", api)

	rsp, err := w.client.Get(api)
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
func (w *worker) queryStarsParams() (q url.Values) {
	// https://www.inoreader.com/developers/stream-contents
	q = make(url.Values)
	// n - Number of items to return (default 20, max 1000).
	q.Set("n", "50")
	// it - Include Target - You can query for a certain label with this.
	// Accepted values: user/-/state/com.google/starred, user/-/state/com.google/like.
	q.Set("it", inoreader.TagStarred.String())
	return q
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
