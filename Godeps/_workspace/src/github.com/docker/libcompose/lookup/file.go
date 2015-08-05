package lookup

import (
	"io/ioutil"
	"path"
	"strings"

	"github.com/Sirupsen/logrus"
)

type FileConfigLookup struct {
}

func (f *FileConfigLookup) Lookup(file, relativeTo string) ([]byte, string, error) {
	if strings.HasPrefix(file, "/") {
		logrus.Debugf("Reading file %s", file)
		bytes, err := ioutil.ReadFile(file)
		return bytes, file, err
	}

	fileName := path.Join(path.Dir(relativeTo), file)
	logrus.Debugf("Reading file %s relative to %s", fileName, relativeTo)
	bytes, err := ioutil.ReadFile(fileName)
	return bytes, fileName, err
}
