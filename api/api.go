package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	TokenURLBase      string
	TokenHostOverride string
	APIURLBase        string
	APIHostOverride   string
	ClientID          string
	ClientSecret      string
)

type APIClient struct {
	AccessToken string
}

type APIResponse struct {
	Meta map[string]interface{}
	Data interface{}
}

type AuthResponse struct {
	AccessToken string `json:"access_token"`
	Error       string
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

type User struct {
	APIClient
	LastFetch time.Time
	APIObject map[string]interface{}
	Flags struct {
		SentHelp bool
		SentIntro bool
	}
}

func (user *User) UserID() string {
	return user.APIObject["id"].(string)
}

func (user *User) Username() string {
	return user.APIObject["username"].(string)
}

func (user *User) IsExpired() bool {
	return time.Since(user.LastFetch) > 10*time.Minute
}

var userCache = make(map[string]*User)

func GetUser(apiObject map[string]interface{}, scopes []string) (user *User, err error) {
	return GetUserByID(apiObject["id"].(string), scopes, apiObject)
}

func GetUserByID(userID string, scopes []string, apiObject map[string]interface{}) (user *User, err error) {
	user, ok := userCache[userID]
	if ok && !user.IsExpired() {
		return user, nil
	}

	if !ok {
		user = &User{}
	}

	// go get a token and create/update user object

	tokenParams := map[string]string{
		"grant_type": "xyx_mxml_internal_implicit_token",
		"user_id":    userID,
	}

	if len(scopes) > 0 {
		tokenParams["scope"] = strings.Join(scopes, ",")
	}

	userClient, err := GetToken(tokenParams)

	if err != nil {
		return nil, err
	}

	user.AccessToken = userClient.AccessToken

	if apiObject != nil {
		user.APIObject = apiObject
		user.LastFetch = time.Now()
	} else {
		user.Refresh()
	}

	userCache[user.UserID()] = user

	return user, nil
}

func GetToken(params map[string]string) (result *APIClient, err error) {
	endpoint := TokenURLBase + "/oauth/access_token"

	values := make(url.Values)

	if params != nil {
		for key, val := range params {
			values.Set(key, val)
		}
	}

	req, err := http.NewRequest("POST", endpoint, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(ClientID, ClientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if TokenHostOverride != "" {
		req.Host = TokenHostOverride
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	rbody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	m := &AuthResponse{}

	err = json.Unmarshal(rbody, &m)
	if err != nil {
		return nil, err
	}

	if m.Error != "" {
		return nil, errors.New(m.Error)
	}

	return &APIClient{AccessToken: m.AccessToken}, nil
}

func (client *APIClient) apiCall(method string, endpoint string, contentType string, body io.Reader, params map[string]string) (result *APIResponse, err error) {
	endpoint = APIURLBase + endpoint

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

	if APIHostOverride != "" {
		req.Host = APIHostOverride
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

func (user *User) Refresh() (err error) {
	obj, err := user.Get("/stream/0/token", nil)
	if err != nil {
		return err
	}

	user.APIObject = obj.Data.(map[string]interface{})["user"].(map[string]interface{})
	user.LastFetch = time.Now()

	return nil
}

func (client *APIClient) Reply(channelID string, contents map[string]interface{}) {
	endpoint := fmt.Sprintf("/stream/0/channels/%s/messages", channelID)

	if _, err := client.PostJSON(endpoint, nil, contents); err != nil {
		log.Print("Error replying in channel ", channelID, ": ", err)
	}
}

func (client *APIClient) GetStreamEndpoint(key string) string {
	params := map[string]string{
		"key": key,
	}

	obj, err := client.Get("/stream/0/streams", params)
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

func (user *User) GetInvite(email string) (string, int, error) {
	params := map[string]string{
		// empty post body :(
		"foo": "bar",
	}

	if email != "" {
		params["email"] = email
	}

	obj, err := user.Post("/stream/0/users/invite", nil, params)
	if err != nil {
		return "", 0, err
	}

	inviteURL := obj.Data.(map[string]interface{})["url"].(string)
	remainingCount := int(obj.Data.(map[string]interface{})["remaining_count"].(float64))

	return inviteURL, remainingCount, nil
}

func (user *User) GetInviteCount() (int, error) {
	obj, err := user.Get("/stream/0/users/invite/count", nil)
	if err != nil {
		return 0, err
	}

	remainingCount := int(obj.Data.(map[string]interface{})["remaining_count"].(float64))

	return remainingCount, nil
}
