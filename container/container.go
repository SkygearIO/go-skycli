package container

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"fmt"
)

const actionPartSeparator = ":"
const requestPartSeparator = "/"

// Container is a client-side view of remote Skygear functionality
type Container struct {
	APIKey      string
	Endpoint    string
	AccessToken string
}

// actionURL construct the corresponding URL to Skygear
func (c *Container) actionURL(action string) string {
	return c.Endpoint + "/" + strings.Replace(action, actionPartSeparator, requestPartSeparator, -1)
}

// createRequest add the necessary header to request
func (c *Container) createRequest(method, url, contentType string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	if c.APIKey != "" {
		req.Header.Set("X-Skygear-API-Key", c.APIKey)
	}
	if c.AccessToken != "" {
		req.Header.Set("X-Skygear-Access-Token", c.AccessToken)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	return req, nil
}

func (c *Container) fixRequestPayload(action string, payload map[string]interface{}) {
	if c.AccessToken != "" {
		payload["access_token"] = c.AccessToken
	}
	if action != "" {
		payload["action"] = action
	}
}

func getBytesResponse(req *http.Request) ([]byte, error) {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	jsonDataFromHTTP, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return jsonDataFromHTTP, nil
}

// MakeRequest sends request to Skygear
func (c *Container) MakeRequest(action string, request SkygearRequest) (response *SkygearResponse, err error) {
	url := c.actionURL(action)
	payload := request.MakePayload()
	c.fixRequestPayload(action, payload)

	jsonStr, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := c.createRequest("POST", url, "", bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}

	jsonDataFromHTTP, err := getBytesResponse(req)
	if err != nil {
		return nil, err
	}

	var jsonData map[string]interface{}
	err = json.Unmarshal(jsonDataFromHTTP, &jsonData)
	if err != nil {
		return
	}

	return &SkygearResponse{Payload: jsonData}, nil

}

func (c *Container) assetURL(filename string) string {
	expiredAt := time.Now().Add(time.Minute).UTC().Unix()
	url := c.Endpoint + "/files/" + filename + "?expiredAt=" + fmt.Sprintf("%d", expiredAt)
	return url
}

// PutAssetRequest sends asset PUT request to Skygear.
func (c *Container) PutAssetRequest(filename, contentType string, body io.Reader) (response *SkygearResponse, err error) {
	url := c.assetURL(filename)
	req, err := c.createRequest("PUT", url, contentType, body)
	if err != nil {
		return nil, err
	}

	jsonDataFromHTTP, err := getBytesResponse(req)
	if err != nil {
		return nil, err
	}

	var jsonData map[string]interface{}
	err = json.Unmarshal(jsonDataFromHTTP, &jsonData)
	if err != nil {
		return nil, err
	}

	return &SkygearResponse{Payload: jsonData}, nil
}

// GetAssetRequest sends GET request to Skygear and get the corresponding asset.
func (c *Container) GetAssetRequest(assetURL string) (response []byte, err error) {
	req, err := c.createRequest("GET", assetURL, "", nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Unexpected status code.")
	}

	dataFromHTTP, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return dataFromHTTP, nil
}

// PublicDatabaseID returns ID of the public database
func (c *Container) PublicDatabaseID() string {
	return "_public"
}

// PrivateDatabaseID returns ID of the current user's private database
func (c *Container) PrivateDatabaseID() string {
	return "_private"
}
