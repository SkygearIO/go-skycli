package container

// OurdRequest encapsulates payload for making Ourd requests
type OurdRequest interface {
	// MakePayload creates map structure of the payload for sending
	// to remove server
	MakePayload() map[string]interface{}
}

// GenericRequest implements paylaod for a generic request
type GenericRequest struct {
	Payload map[string]interface{}
}

// MakePayload creates request payload for a generic request
func (r *GenericRequest) MakePayload() map[string]interface{} {
	return r.Payload
}

// OurdResponse encapsulates payload received from Ourd
type OurdResponse struct {
	Payload map[string]interface{}
}

// IsError returns if response is an error
func (r *OurdResponse) IsError() bool {
	_, ok := r.Payload["error"]
	return ok
}

// Error returns error in the response if any
func (r *OurdResponse) Error() *OurdError {
	data, ok := r.Payload["error"].(map[string]interface{})
	if !ok {
		return nil
	}
	ourdError := MakeError(data)
	return &ourdError
}

// OurdError encapsulates data of an Ourd response
type OurdError struct {
	ID      string
	Message string
	Code    int
	Type    string
}

// MakeError creates an OurdError
func MakeError(data map[string]interface{}) OurdError {
	err := OurdError{}
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
