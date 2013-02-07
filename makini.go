package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
    "flag"
)

import "github.com/kylelemons/go-gypsy/yaml"

const ADN_API_BASE = "https://alpha-api.app.net/stream/0/"

var (
    file = flag.String("config", "config.yaml", "YAML config file")
)

var user_id string
var adn_app_access_token string
var adn_user_access_token string

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

func apiCall(method string, endpoint string, contentType string, body io.Reader, params map[string]string, useAppToken bool) (result *APIResponse, err error) {
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

	var token string

	if useAppToken {
		token = adn_app_access_token
	} else {
		token = adn_user_access_token
	}

	req.Header.Add("Authorization", "BEARER "+token)

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

func apiGet(endpoint string, params map[string]string, useAppToken bool) (result *APIResponse, err error) {
	return apiCall("GET", endpoint, "", nil, params, useAppToken)
}

func apiPost(endpoint string, params map[string]string, postBody map[string]string, useAppToken bool) (result *APIResponse, err error) {
	values := make(url.Values)

	for key, val := range postBody {
		values.Set(key, val)
	}

	return apiCall("POST", endpoint, "application/x-www-form-urlencoded", strings.NewReader(values.Encode()), params, useAppToken)
}

func apiPostJson(endpoint string, params map[string]string, postBody interface{}, useAppToken bool) (result *APIResponse, err error) {
	body, err := json.Marshal(postBody)
	if err != nil {
		return nil, err
	}

	return apiCall("POST", endpoint, "application/json", bytes.NewBuffer(body), params, useAppToken)
}

func consumeStream(url string, msgChan chan []byte) {
	for {
		res, err := http.Get(url)

		if err == nil {
			buf := bufio.NewReader(res.Body)

			for {
				line, err := buf.ReadBytes('\n')

				if err == nil {
					if len(line) > 0 {
						msgChan <- line
					}
				} else {
					log.Print("Error while reading: ", err)
					break
				}
			}
		} else {
			log.Print("Error while connecting: ", err)
		}

		time.Sleep(time.Second)
	}
}

func unmarshalStream(in chan []byte, out chan *APIResponse) {
	for {
		m := &APIResponse{}
		err := json.Unmarshal(<-in, &m)

		if err != nil {
			log.Print("Error decoding: ", err)
		} else {
			out <- m
		}
	}
}

func reply(channelID string, text string) {
	endpoint := fmt.Sprintf("channels/%s/messages", channelID)

	body := map[string]string{
		"text": text,
	}

	if _, err := apiPostJson(endpoint, nil, body, false); err != nil {
		log.Print("Error replying in channel ", channelID, ": ", err)
	}
}

func logStream(in chan *APIResponse) {
	for {
		obj := <-in

		if obj.Meta["type"] == "message" && obj.Meta["channel_type"] == "net.app.core.pm" {
			if data, ok := obj.Data.(map[string]interface{}); ok {
				if user, ok := data["user"].(map[string]interface{}); ok {
					if user["id"] != user_id {
						log.Print("Got message: ", data["text"], " from ", user["username"])
						msg := fmt.Sprintf("Hi, @%s! What's up?", user["username"])
						reply(data["channel_id"].(string), msg)
					}
				}
			}
		}
	}
}

func getStreamEndpoint() string {
	params := map[string]string{
		"key": "makini",
	}

	obj, err := apiGet("streams", params, true)
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

func getUserID() string {
	obj, err := apiGet("token", nil, false)
	if err != nil {
		log.Fatal("Error getting token: ", err)
	}

	token := obj.Data.(map[string]interface{})
	user := token["user"].(map[string]interface{})
	user_id := user["id"].(string)

	return user_id
}

func main() {
    flag.Parse()

    config, err := yaml.ReadFile(*file)
    if err != nil {
        log.Fatalf("Error loading config (%q): %s", *file, err)
    }

    adn_user_access_token, _ = config.Get("tokens.user")
    adn_app_access_token, _ = config.Get("tokens.app")

	url := getStreamEndpoint()
	user_id = getUserID()

	bytes := make(chan []byte)
	go consumeStream(url, bytes)

	messages := make(chan *APIResponse)
	go unmarshalStream(bytes, messages)

	logStream(messages)
}
