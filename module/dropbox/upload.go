package dropbox

import (
	"bytes"
	"fmt"
	"io"
	"path"
	"sync"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/dustin/go-humanize"
	"github.com/sirupsen/logrus"
)

var mux sync.Mutex

func Upload(name string, size int64, reader io.ReadCloser) (f *files.FileMetadata, err error) {
	defer reader.Close()

	mux.Lock()
	defer mux.Unlock()

	logrus.Debugf("upload %s", name)
	start := time.Now()
	defer func() {
		if err == nil {
			logrus.Infof("upload %s done: %s, %s", name, humanize.IBytes(f.Size), time.Now().Sub(start).Round(time.Millisecond))
		} else {
			logrus.Errorf("upload %s failed: %s", name, err)
		}
	}()

	return upload(name, size, reader)
}
func upload(name string, size int64, reader io.Reader) (*files.FileMetadata, error) {
	name = path.Join("/", name)

	const chunk = humanize.MiByte * 100
	if size > 0 && size < chunk {
		return uploadTiny(name, size, reader)
	}

	var cursor *files.UploadSessionCursor
	var buffer bytes.Buffer
	for {
		n, e := io.CopyN(&buffer, reader, chunk)
		if e != nil && e != io.EOF {
			return nil, fmt.Errorf("read data failed: %s", e)
		}

		switch {
		case cursor == nil:
			if n < chunk {
				return uploadTiny(name, n, &buffer)
			}
			logrus.Debugf("start %s-%s", humanize.IBytes(0), humanize.IBytes(uint64(n)))
			session, err := client.UploadSessionStart(files.NewUploadSessionStartArg(), &buffer)
			if err != nil {
				return nil, fmt.Errorf("UploadSessionStart failed: %s", err)
			}
			cursor = files.NewUploadSessionCursor(session.SessionId, 0)
		case e == io.EOF:
			logrus.Debugf("finish %s-%s", humanize.IBytes(cursor.Offset), humanize.IBytes(cursor.Offset+uint64(n)))
			return client.UploadSessionFinish(files.NewUploadSessionFinishArg(cursor, files.NewCommitInfo(name)), &buffer)
		default:
			logrus.Debugf("append %s-%s", humanize.IBytes(cursor.Offset), humanize.IBytes(cursor.Offset+uint64(n)))
			e := client.UploadSessionAppendV2(files.NewUploadSessionAppendArg(cursor), &buffer)
			if e != nil {
				return nil, fmt.Errorf("UploadSessionAppendV2 failed: %s", e)
			}
		}

		cursor.Offset += uint64(n)
		buffer.Reset()
	}
}
func uploadTiny(name string, size int64, r io.Reader) (*files.FileMetadata, error) {
	logrus.Debugf("tiny %s", humanize.IBytes(uint64(size)))
	return client.Upload(files.NewCommitInfo(name), r)
}
