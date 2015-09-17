package container

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
)

const actionPartSeparator = ":"
const requestPartSeparator = "/"

// Container is a client-side view of remote Ourd functionality
type Container struct {
	APIKey      string
	Endpoint    string
	AccessToken string
}

func (c *Container) actionURL(action string) string {
	return c.Endpoint + "/" + strings.Replace(action, actionPartSeparator, requestPartSeparator, -1)
}

func (c *Container) createRequest(action string, payload map[string]interface{}) *http.Request {

	url := c.actionURL(action)
	//fmt.Printf("making request for: %v\n", url)
	if c.AccessToken != "" {
		payload["access_token"] = c.AccessToken
	}
	if action != "" {
		payload["action"] = action
	}
	var jsonStr, _ = json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("X-Ourd-API-Key", c.APIKey)
	if c.AccessToken != "" {
		req.Header.Set("X-Ourd-Access-Token", c.AccessToken)
	}
	return req
}

// MakeRequest sends request to Ourd
func (c *Container) MakeRequest(action string, request OurdRequest) (response *OurdResponse, err error) {

	req := c.createRequest(action, request.MakePayload())
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	jsonDataFromHTTP, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return
	}

	var jsonData map[string]interface{}
	err = json.Unmarshal([]byte(jsonDataFromHTTP), &jsonData)

	if err != nil {
		return
	}

	return &OurdResponse{Payload: jsonData}, nil

}

// PrivateDatabase returns ID of the public database
func (c *Container) PublicDatabaseID() string {
	return "_public"
}

// PrivateDatabase returns ID of the current user's private database
func (c *Container) PrivateDatabaseID() string {
	return "_private"
}
