package dbx

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"sync"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/dustin/go-humanize"
	"github.com/sirupsen/logrus"
)

var (
	config = dropbox.Config{
		Token:    os.Getenv("DROPBOX_TOKEN"),
		LogLevel: dropbox.LogInfo, // if needed, set the desired logging level. Default is off
	}
	client = files.New(config)

	lock sync.Mutex
)

func Upload(name string, size int64, r io.ReadCloser) (f *files.FileMetadata, err error) {
	defer r.Close()

	lock.Lock()
	defer lock.Unlock()

	logrus.Debugf("upload %s", name)
	start := time.Now()
	defer func() {
		if err == nil {
			logrus.Infof("upload %s done: %s, %s", name, humanize.IBytes(f.Size), time.Now().Sub(start).Round(time.Millisecond))
		} else {
			logrus.Errorf("upload %s failed: %s", name, err)
		}
	}()

	return upload(name, size, r)
}
func upload(name string, size int64, r io.Reader) (*files.FileMetadata, error) {
	name = path.Join("/", name)

	const chunk = humanize.MiByte * 100
	if size > 0 && size < chunk {
		return uploadTiny(name, size, r)
	}

	var c *files.UploadSessionCursor
	var b bytes.Buffer
	for {
		n, e := io.CopyN(&b, r, chunk)
		if e != nil && e != io.EOF {
			return nil, fmt.Errorf("read data failed: %s", e)
		}

		switch {
		case c == nil:
			if n < chunk {
				return uploadTiny(name, n, &b)
			}
			logrus.Debugf("start %s-%s", humanize.IBytes(0), humanize.IBytes(uint64(n)))
			ss, err := client.UploadSessionStart(files.NewUploadSessionStartArg(), &b)
			if err != nil {
				return nil, fmt.Errorf("UploadSessionStart failed: %s", err)
			}
			c = files.NewUploadSessionCursor(ss.SessionId, 0)
		case e == io.EOF:
			logrus.Debugf("finish %s-%s", humanize.IBytes(c.Offset), humanize.IBytes(c.Offset+uint64(n)))
			return client.UploadSessionFinish(files.NewUploadSessionFinishArg(c, files.NewCommitInfo(name)), &b)
		default:
			logrus.Debugf("append %s-%s", humanize.IBytes(c.Offset), humanize.IBytes(c.Offset+uint64(n)))
			e := client.UploadSessionAppendV2(files.NewUploadSessionAppendArg(c), &b)
			if e != nil {
				return nil, fmt.Errorf("UploadSessionAppendV2 failed: %s", e)
			}
		}

		c.Offset += uint64(n)
		b.Reset()
	}
}
func uploadTiny(name string, size int64, r io.Reader) (*files.FileMetadata, error) {
	logrus.Debugf("tiny %s", humanize.IBytes(uint64(size)))
	return client.Upload(files.NewCommitInfo(name), r)
}

func Read(name string) (res *files.FileMetadata, content io.ReadCloser, err error) {
	name = path.Join("/", name)
	return client.Download(files.NewDownloadArg(name))
}
func Delete(name string) (res *files.DeleteResult, err error) {
	name = path.Join("/", name)
	return client.DeleteV2(files.NewDeleteArg(name))
}
