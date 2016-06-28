package glue

import (
	"encoding/json"
	"io"
	"os"

	"github.com/pkg/errors"
)

func isZero(path string) bool {
	stat, err := os.Stat(path)
	if os.IsNotExist(err) {
		return true
	}
	if err != nil {
		return true
	}

	return stat.Size() == 0
}

func copyToExistingFile(to, from string) error {
	src, err := os.Open(from)
	if err != nil {
		return errors.Wrap(err, "opening file "+from)
	}
	defer src.Close()

	dest, err := os.OpenFile(to, os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		return errors.Wrap(err, "opening file "+to)
	}
	defer dest.Close()

	_, err = io.Copy(dest, src)
	return err
}

func readJSONFile(file string, obj interface{}) error {
	f, err := os.Open(file)
	if err != nil {
		return errors.Wrap(err, "opening "+file)
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(obj); err != nil {
		return errors.Wrap(err, "unmarshaling "+file)
	}
	return nil
}
