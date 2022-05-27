package dropbox

import (
	"path"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
)

func Exist(name string) bool {
	name = path.Join("/", name)
	arg := &files.GetMetadataArg{Path: name}
	_, err := client.GetMetadata(arg)
	return err == nil
}
