package dropbox

import (
	"path"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
)

func Delete(name string) (res *files.DeleteResult, err error) {
	name = path.Join("/", name)
	return client.DeleteV2(files.NewDeleteArg(name))
}
