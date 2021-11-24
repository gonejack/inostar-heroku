package reader

import (
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/gonejack/inostar-heroku/dbx"
	"github.com/gonejack/inostar-heroku/util"
)

type Info struct {
	LastStarTimeRaw string `json:"last_star_time"`
}

func (i *Info) From(r io.Reader) {
	_ = json.NewDecoder(r).Decode(i)
}
func (i *Info) String() string {
	return util.JsonDump(i)
}
func (i *Info) LastStarTime() (t time.Time) {
	t, err := time.Parse(time.RFC3339, i.LastStarTimeRaw)
	if i.LastStarTimeRaw == "" || err != nil {
		t = time.Now()
		i.SetLastStarTime(t)
	}
	return
}
func (i *Info) SetLastStarTime(t time.Time) {
	i.LastStarTimeRaw = t.Format(time.RFC3339)
}

const info_name = "inostar.json"

func (i *Info) Query() {
	_, body, err := dbx.Read(info_name)
	if err != nil {
		if strings.Contains(err.Error(), "not_found") {
			logrus.Debugf("read %s failed: %s", info_name, err)
		} else {
			logrus.Errorf("read %s failed: %s", info_name, err)
		}
		return
	}
	defer body.Close()

	i.From(body)

	if logrus.IsLevelEnabled(logrus.DebugLevel) {
		logrus.Debugf("query info: %s", i.String())
	}
}
func (i *Info) Save() {
	_, err := dbx.Delete(info_name)
	if err != nil {
		if !strings.Contains(err.Error(), "not_found") {
			logrus.Errorf("remove %s failed: %s", info_name, err)
		}
	}

	if logrus.IsLevelEnabled(logrus.DebugLevel) {
		logrus.Debugf("save info: %s", i.String())
	}
	_, err = dbx.Upload(info_name, 0, io.NopCloser(strings.NewReader(i.String())))
	if err != nil {
		logrus.Errorf("save %s failed: %s", info_name, err)
	}
}
