package utils

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/Sirupsen/logrus"

	"gopkg.in/yaml.v2"
)

// InParallel holds a pool and a waitgroup to execute tasks in parallel and to be able
// to wait for completion of all tasks.
type InParallel struct {
	wg   sync.WaitGroup
	pool sync.Pool
}

// Add adds runs the specified task in parallel and add it to the waitGroup.
func (i *InParallel) Add(task func() error) {
	i.wg.Add(1)

	go func() {
		defer i.wg.Done()
		err := task()
		if err != nil {
			i.pool.Put(err)
		}
	}()
}

// Wait waits for all tasks to complete and returns the latests error encountered if any.
func (i *InParallel) Wait() error {
	i.wg.Wait()
	obj := i.pool.Get()
	if err, ok := obj.(error); ok {
		return err
	}
	return nil
}

// ConvertByJSON converts a struct (src) to another one (target) using json marshalling/unmarshalling.
// If the structure are not compatible, this will throw an error as the unmarshalling will fail.
func ConvertByJSON(src, target interface{}) error {
	newBytes, err := json.Marshal(src)
	if err != nil {
		return err
	}

	err = json.Unmarshal(newBytes, target)
	if err != nil {
		logrus.Errorf("Failed to unmarshall: %v\n%s", err, string(newBytes))
	}
	return err
}

// Convert converts a struct (src) to another one (target) using yaml marshalling/unmarshalling.
// If the structure are not compatible, this will throw an error as the unmarshalling will fail.
func Convert(src, target interface{}) error {
	newBytes, err := yaml.Marshal(src)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(newBytes, target)
	if err != nil {
		logrus.Errorf("Failed to unmarshall: %v\n%s", err, string(newBytes))
	}
	return err
}

// FilterString returns a json representation of the specified map
// that is used as filter for docker.
func FilterString(data map[string][]string) string {
	// I can't imagine this would ever fail
	bytes, _ := json.Marshal(data)
	return string(bytes)
}

// LabelFilter returns a label json representation of the specifed couple (key,value)
// that is used as filter for docker.
func LabelFilter(key, value string) string {
	return FilterString(map[string][]string{
		"label": {fmt.Sprintf("%s=%s", key, value)},
	})
}

// Contains checks if the specified string (key) is present in the specified collection.
func Contains(collection []string, key string) bool {
	for _, value := range collection {
		if value == key {
			return true
		}
	}

	return false
}
