package dropbox

import (
	"path"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/sirupsen/logrus"
)

func Exist(name string) bool {
	name = path.Join("/", name)
	arg := &files.GetMetadataArg{
		Path: name,
	}
	res, err := client.GetMetadata(arg)
	logrus.Infof("print %v, %s", res, err)
	return err == nil
}
