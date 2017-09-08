package api

import (
	"encoding/json"
	"errors"
	"reflect"

	"github.com/rancher/go-rancher/client"
	v2client "github.com/rancher/go-rancher/v2"
)

func toMap(obj interface{}) (map[string]interface{}, error) {
	result := map[string]interface{}{}
	bytes, err := json.Marshal(obj)
	if err != nil {
		return result, err
	}

	err = json.Unmarshal(bytes, &result)
	return result, err
}

func getEmbedded(obj interface{}, checkType reflect.Type) interface{} {
	val := reflect.ValueOf(obj)
	for val.Kind() == reflect.Ptr || val.Kind() == reflect.Interface {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		panic("Passed type is not a struct got: " + val.Kind().String())
	}

	t := val.Type()
	for i := 0; i < val.NumField(); i++ {
		if t.Field(i).Anonymous && t.Field(i).Type == checkType {
			if !val.Field(i).CanAddr() {
				panic("Field " + t.Field(i).Name + " is not addressable, pass pointer")
			}
			return val.Field(i).Addr().Interface()
		}
	}

	return nil
}

func getCollection(obj interface{}) interface{} {
	val := getEmbedded(obj, reflect.TypeOf(client.Collection{}))
	if val == nil {
		val = getEmbedded(obj, reflect.TypeOf(v2client.Collection{}))
		if val == nil {
			return nil
		}
	}
	return val
}

func getResource(obj interface{}) (*client.Resource, *v2client.Resource) {
	r, ok := obj.(*client.Resource)
	if ok {
		return r, nil
	}
	rObj, ok := obj.(client.Resource)
	if ok {
		return &rObj, nil
	}
	v2r, ok := obj.(*v2client.Resource)
	if ok {
		return nil, v2r
	}
	v2rObj, ok := obj.(v2client.Resource)
	if ok {
		return nil, &v2rObj
	}
	val := getEmbedded(obj, reflect.TypeOf(client.Resource{}))
	if val == nil {
		val = getEmbedded(obj, reflect.TypeOf(v2client.Resource{}))
		if val == nil {
			return nil, nil
		}
	}
	v1Resource, ok := val.(*client.Resource)
	if ok {
		return v1Resource, nil
	}
	return nil, val.(*v2client.Resource)
}

func CollectionToMap(obj interface{}, schemas *client.Schemas) (map[string]interface{}, []map[string]interface{}, error) {
	result := map[string]interface{}{}
	data := []map[string]interface{}{}
	if obj == nil {
		return result, data, nil
	}

	c := getCollection(obj)
	if c == nil {
		return result, data, errors.New("value is not a Collection")
	}

	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	val = val.FieldByName("Data")
	for i := 0; i < val.Len(); i++ {
		obj := val.Index(i)
		if obj.Kind() != reflect.Ptr && obj.Kind() != reflect.Interface {
			obj = obj.Addr()
		}
		dataObj, err := ResourceToMap(obj.Interface(), schemas)
		if err != nil {
			return result, data, err
		}

		data = append(data, dataObj)
	}

	collectionMap, err := toMap(obj)
	if err != nil {
		return result, data, err
	}

	collectionMap["data"] = data
	return collectionMap, data, nil
}

func ResourceToMap(obj interface{}, schemas *client.Schemas) (map[string]interface{}, error) {
	var resourceType string
	result := map[string]interface{}{}
	if obj == nil {
		return result, nil
	}

	v1resource, v2resource := getResource(obj)
	if v1resource == nil && v2resource == nil {
		return result, errors.New("value is not a Resource")
	}

	if v1resource != nil {
		resourceType = v1resource.Type
	} else {
		resourceType = v2resource.Type
	}

	objMap, err := toMap(obj)
	if err != nil {
		return result, err
	}

	schema := schemas.Schema(resourceType)
	for k, v := range objMap {
		_, ok := schema.CheckField(k)
		if !ok {
			continue
		}
		result[k] = v
	}

	return result, nil
}
