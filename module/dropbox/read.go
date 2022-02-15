package dropbox

import (
	"io"
	"path"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
)

func Read(name string) (res *files.FileMetadata, content io.ReadCloser, err error) {
	name = path.Join("/", name)
	return client.Download(files.NewDownloadArg(name))
}
