package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
)

const (
	SELF       = "self"
	COLLECTION = "collection"
)

type ClientOpts struct {
	Url       string
	AccessKey string
	SecretKey string
}

type ApiError struct {
	StatusCode int
	Url        string
	Msg        string
	Status     string
	Body       string
}

func (e *ApiError) Error() string {
	return e.Msg
}

func newApiError(resp *http.Response, url string) *ApiError {
	contents, err := ioutil.ReadAll(resp.Body)
	var body string
	if err != nil {
		body = "Unreadable body."
	} else {
		body = string(contents)
	}
	formattedMsg := fmt.Sprintf("Bad response from [%s], statusCode [%d]. Status [%s]. Body: [%s]",
		url, resp.StatusCode, resp.Status, body)
	return &ApiError{
		Url:        url,
		Msg:        formattedMsg,
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Body:       body,
	}
}

func contains(array []string, item string) bool {
	for _, check := range array {
		if check == item {
			return true
		}
	}

	return false
}

func appendFilters(urlString string, filters map[string]interface{}) (string, error) {
	if len(filters) == 0 {
		return urlString, nil
	}

	u, err := url.Parse(urlString)
	if err != nil {
		return "", err
	}

	q := u.Query()
	for k, v := range filters {
		q.Add(k, fmt.Sprintf("%v", v))
	}

	u.RawQuery = q.Encode()
	return u.String(), nil
}

func setupRancherBaseClient(rancherClient *RancherBaseClient, opts *ClientOpts) error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", opts.Url, nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(opts.AccessKey, opts.SecretKey)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return newApiError(resp, opts.Url)
	}

	schemasUrls := resp.Header.Get("X-API-Schemas")
	if len(schemasUrls) == 0 {
		return errors.New("Failed to find schema at [" + opts.Url + "]")
	}

	if schemasUrls != opts.Url {
		req, err = http.NewRequest("GET", schemasUrls, nil)
		req.SetBasicAuth(opts.AccessKey, opts.SecretKey)
		if err != nil {
			return err
		}

		resp, err = client.Do(req)
		if err != nil {
			return err
		}

		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return newApiError(resp, opts.Url)
		}
	}

	var schemas Schemas
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, &schemas)
	if err != nil {
		return err
	}

	rancherClient.Opts = opts
	rancherClient.Schemas = &schemas

	for _, schema := range schemas.Data {
		rancherClient.Types[schema.Id] = schema
	}

	return nil
}

func NewListOpts() *ListOpts {
	return &ListOpts{
		Filters: map[string]interface{}{},
	}
}

func (rancherClient *RancherBaseClient) setupRequest(req *http.Request) {
	req.SetBasicAuth(rancherClient.Opts.AccessKey, rancherClient.Opts.SecretKey)
}

func (rancherClient *RancherBaseClient) newHttpClient() *http.Client {
	return &http.Client{}
}

func (rancherClient *RancherBaseClient) doDelete(url string) error {
	client := rancherClient.newHttpClient()
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	rancherClient.setupRequest(req)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return newApiError(resp, url)
	}

	return nil
}

func (rancherClient *RancherBaseClient) doGet(url string, opts *ListOpts, respObject interface{}) error {
	if opts == nil {
		opts = NewListOpts()
	}
	url, err := appendFilters(url, opts.Filters)
	if err != nil {
		return err
	}

	client := rancherClient.newHttpClient()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	rancherClient.setupRequest(req)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return newApiError(resp, url)
	}

	byteContent, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(byteContent, respObject)
}

func (rancherClient *RancherBaseClient) doList(schemaType string, opts *ListOpts, respObject interface{}) error {
	schema, ok := rancherClient.Types[schemaType]
	if !ok {
		return errors.New("Unknown schema type [" + schemaType + "]")
	}

	if !contains(schema.CollectionMethods, "GET") {
		return errors.New("Resource type [" + schemaType + "] is not listable")
	}

	collectionUrl, ok := schema.Links[COLLECTION]
	if !ok {
		return errors.New("Failed to find collection URL for [" + schemaType + "]")
	}

	return rancherClient.doGet(collectionUrl, opts, respObject)
}

func (rancherClient *RancherBaseClient) doModify(method string, url string, createObj interface{}, respObject interface{}) error {
	bodyContent, err := json.Marshal(createObj)
	if err != nil {
		return err
	}

	client := rancherClient.newHttpClient()
	req, err := http.NewRequest(method, url, bytes.NewBuffer(bodyContent))
	if err != nil {
		return err
	}

	rancherClient.setupRequest(req)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Length", string(len(bodyContent)))

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return newApiError(resp, url)
	}

	byteContent, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if len(byteContent) > 0 {
		return json.Unmarshal(byteContent, respObject)
	}
	return nil
}

func (rancherClient *RancherBaseClient) doCreate(schemaType string, createObj interface{}, respObject interface{}) error {
	if createObj == nil {
		createObj = map[string]string{}
	}
	schema, ok := rancherClient.Types[schemaType]
	if !ok {
		return errors.New("Unknown schema type [" + schemaType + "]")
	}

	if !contains(schema.CollectionMethods, "POST") {
		return errors.New("Resource type [" + schemaType + "] is not creatable")
	}

	var collectionUrl string
	collectionUrl, ok = schema.Links[COLLECTION]
	if !ok {
		// return errors.New("Failed to find collection URL for [" + schemaType + "]")
		// This is a hack to address https://github.com/rancherio/cattle/issues/254
		re := regexp.MustCompile("schemas.*")
		collectionUrl = re.ReplaceAllString(schema.Links[SELF], schema.PluralName)
	}

	return rancherClient.doModify("POST", collectionUrl, createObj, respObject)
}

func (rancherClient *RancherBaseClient) doUpdate(schemaType string, existing *Resource, updates interface{}, respObject interface{}) error {
	if existing == nil {
		return errors.New("Existing object is nil")
	}

	selfUrl, ok := existing.Links[SELF]
	if !ok {
		return errors.New(fmt.Sprintf("Failed to find self URL of [%v]", existing))
	}

	if updates == nil {
		updates = map[string]string{}
	}

	schema, ok := rancherClient.Types[schemaType]
	if !ok {
		return errors.New("Unknown schema type [" + schemaType + "]")
	}

	if !contains(schema.ResourceMethods, "PUT") {
		return errors.New("Resource type [" + schemaType + "] is not updatable")
	}

	return rancherClient.doModify("PUT", selfUrl, updates, respObject)
}

func (rancherClient *RancherBaseClient) doById(schemaType string, id string, respObject interface{}) error {
	schema, ok := rancherClient.Types[schemaType]
	if !ok {
		return errors.New("Unknown schema type [" + schemaType + "]")
	}

	if !contains(schema.ResourceMethods, "GET") {
		return errors.New("Resource type [" + schemaType + "] can not be looked up by ID")
	}

	collectionUrl, ok := schema.Links[COLLECTION]
	if !ok {
		return errors.New("Failed to find collection URL for [" + schemaType + "]")
	}

	err := rancherClient.doGet(collectionUrl+"/"+id, nil, respObject)
	//TODO check for 404 and return nil, nil
	return err
}

func (rancherClient *RancherBaseClient) doResourceDelete(schemaType string, existing *Resource) error {
	schema, ok := rancherClient.Types[schemaType]
	if !ok {
		return errors.New("Unknown schema type [" + schemaType + "]")
	}

	if !contains(schema.ResourceMethods, "DELETE") {
		return errors.New("Resource type [" + schemaType + "] can not be deleted")
	}

	selfUrl, ok := existing.Links[SELF]
	if !ok {
		return errors.New(fmt.Sprintf("Failed to find self URL of [%v]", existing))
	}

	return rancherClient.doDelete(selfUrl)
}

func (rancherClient *RancherBaseClient) doEmptyAction(schemaType string, action string,
	existing *Resource, respObject interface{}) error {
	// TODO Actions with inputs currently not supported.

	if existing == nil {
		return errors.New("Existing object is nil")
	}

	actionUrl, ok := existing.Actions[action]
	if !ok {
		return errors.New(fmt.Sprintf("Action [%v] not available on [%v]", action, existing))
	}

	schema, ok := rancherClient.Types[schemaType]
	if !ok {
		return errors.New("Unknown schema type [" + schemaType + "]")
	}

	if schema.ResourceActions[action].Input != "" {
		return fmt.Errorf("Actions with inputs or outputs not yet support. Input: [%v] Output: [%v].",
			schema.ResourceActions[action].Input)
	}

	client := rancherClient.newHttpClient()
	req, err := http.NewRequest("POST", actionUrl, nil)
	if err != nil {
		return err
	}

	rancherClient.setupRequest(req)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Length", "0")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return newApiError(resp, actionUrl)
	}

	byteContent, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(byteContent, respObject)
}
