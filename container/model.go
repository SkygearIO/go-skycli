package container

// SkygearRequest encapsulates payload for making Skygear requests
type SkygearRequest interface {
	// MakePayload creates map structure of the payload for sending
	// to remove server
	MakePayload() map[string]interface{}
}

// GenericRequest implements payload for a generic request
type GenericRequest struct {
	Payload map[string]interface{}
}

// MakePayload creates request payload for a generic request
func (r *GenericRequest) MakePayload() map[string]interface{} {
	return r.Payload
}

// SkygearResponse encapsulates payload received from Skygear
type SkygearResponse struct {
	Payload map[string]interface{}
}

// IsError returns if response is an error
func (r *SkygearResponse) IsError() bool {
	_, ok := r.Payload["error"]
	return ok
}

// Error returns error in the response if any
func (r *SkygearResponse) Error() *SkygearError {
	data, ok := r.Payload["error"].(map[string]interface{})
	if !ok {
		return nil
	}
	skygearError := MakeError(data)
	return &skygearError
}

// SkygearError encapsulates data of an Skygear response
type SkygearError struct {
	ID      string
	Message string
	Code    int
	Type    string
}

// MakeError creates an SkygearError
func MakeError(data map[string]interface{}) SkygearError {
	err := SkygearError{}
	err.ID, _ = data["_id"].(string)
	err.Message, _ = data["message"].(string)
	if err.Message == "" {
		err.Message = "Unknown Error"
	}
	err.Code, _ = data["code"].(int)
	err.Type, _ = data["type"].(string)
	return err
}

// IsError checks whether the map is containing data for an error
func IsError(data map[string]interface{}) bool {
	return data["_type"] == "error"
}

func (e SkygearError) Error() string {
	return e.Message
}
