package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

const ADN_API_BASE = "https://alpha-api.app.net/stream/0/"

type APIClient struct {
	AccessToken string
}

type APIResponse struct {
	Meta map[string]interface{}
	Data interface{}
}

type APIError struct {
	msg string
}

func (r *APIResponse) IsError() bool {
	_, ok := r.Meta["error_message"]

	return ok
}

func (r *APIResponse) GetError() *APIError {
	return &APIError{msg: r.Meta["error_message"].(string)}
}

func (e APIError) Error() string {
	return e.msg
}

func (client *APIClient) apiCall(method string, endpoint string, contentType string, body io.Reader, params map[string]string) (result *APIResponse, err error) {
	endpoint = ADN_API_BASE + endpoint

	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	values := make(url.Values)

	if params != nil {
		for key, val := range params {
			values.Set(key, val)
		}
	}

	u.RawQuery = values.Encode()

	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "BEARER "+client.AccessToken)

	if method == "POST" {
		req.Header.Set("Content-Type", contentType)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	rbody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	m := &APIResponse{}

	err = json.Unmarshal(rbody, &m)
	if err != nil {
		log.Print("asdf 3", string(rbody))
		return nil, err
	}

	if m.IsError() {
		return nil, error(m.GetError())
	}

	return m, nil
}

func (client *APIClient) Get(endpoint string, params map[string]string) (result *APIResponse, err error) {
	return client.apiCall("GET", endpoint, "", nil, params)
}

func (client *APIClient) Post(endpoint string, params map[string]string, postBody map[string]string) (result *APIResponse, err error) {
	values := make(url.Values)

	for key, val := range postBody {
		values.Set(key, val)
	}

	return client.apiCall("POST", endpoint, "application/x-www-form-urlencoded", strings.NewReader(values.Encode()), params)
}

func (client *APIClient) PostJSON(endpoint string, params map[string]string, postBody interface{}) (result *APIResponse, err error) {
	body, err := json.Marshal(postBody)
	if err != nil {
		return nil, err
	}

	return client.apiCall("POST", endpoint, "application/json", bytes.NewBuffer(body), params)
}

func (client *APIClient) GetUserID() string {
	obj, err := client.Get("token", nil)
	if err != nil {
		log.Fatal("Error getting token: ", err)
	}

	token := obj.Data.(map[string]interface{})
	user := token["user"].(map[string]interface{})
	userID := user["id"].(string)

	return userID
}

func (client *APIClient) Reply(channelID string, text string) {
	endpoint := fmt.Sprintf("channels/%s/messages", channelID)

	body := map[string]string{
		"text": text,
	}

	if _, err := client.PostJSON(endpoint, nil, body); err != nil {
		log.Print("Error replying in channel ", channelID, ": ", err)
	}
}

func (client *APIClient) GetStreamEndpoint(key string) string {
	params := map[string]string{
		"key": key,
	}

	obj, err := client.Get("streams", params)
	if err != nil {
		log.Fatal("Error getting stream endpoint: ", err)
	}

	streams := obj.Data.([]interface{})

	for _, entry := range streams {
		m := entry.(map[string]interface{})
		if endpoint, ok := m["endpoint"]; ok {
			if endpoint_s, ok := endpoint.(string); ok {
				return endpoint_s
			}
		}
	}

	return ""
}
