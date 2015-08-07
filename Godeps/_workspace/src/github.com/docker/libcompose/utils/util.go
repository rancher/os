package utils

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/Sirupsen/logrus"

	"gopkg.in/yaml.v2"
)

type InParallel struct {
	wg   sync.WaitGroup
	pool sync.Pool
}

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

func (i *InParallel) Wait() error {
	i.wg.Wait()
	obj := i.pool.Get()
	if err, ok := obj.(error); ok {
		return err
	} else {
		return nil
	}
}

func ConvertByJson(src, target interface{}) error {
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

func ConvertToInterfaceMap(input map[string]string) map[string]interface{} {
	result := map[string]interface{}{}
	for k, v := range input {
		result[k] = v
	}

	return result
}

func FilterString(data map[string][]string) string {
	// I can't imagine this would ever fail
	bytes, _ := json.Marshal(data)
	return string(bytes)
}

func LabelFilter(key, value string) string {
	return FilterString(map[string][]string{
		"label": {fmt.Sprintf("%s=%s", key, value)},
	})
}

func Contains(collection []string, key string) bool {
	for _, value := range collection {
		if value == key {
			return true
		}
	}

	return false
}
