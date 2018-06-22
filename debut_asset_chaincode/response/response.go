package response

import "encoding/json"

//Response ... The basic response structure
type Response struct {
	Status  uint16      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

//CreateErrorResponse ... This function will create error response
func CreateErrorResponse(message string, status uint16, data interface{}) string {
	// response := Response{Message: message,
	// 	Status: status,
	// 	Data:   data}

	// responseByte, err := json.Marshal(response)

	// if err != nil {
	// 	return "Something went wrong while parsing json"
	// }

	// return string(responseByte)

	return message
}

//CreateSuccessResponse ... This function will create success response
func CreateSuccessResponse(message string, status uint16, data interface{}) []byte {
	response := Response{Message: message,
		Status: status,
		Data:   data}

	responseByte, err := json.Marshal(response)

	if err != nil {
		return []byte("Something went wrong while parsing json")
	}

	return responseByte
}
