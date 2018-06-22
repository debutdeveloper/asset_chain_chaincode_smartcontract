package rq

import (
	"bytes"
	"encoding/json"
	"strconv"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
	"github.com/vjeantet/jodaTime"
)

//Request ... Structure definition for Request
type Request struct {
	Description string    `json:"description"` //args[0]
	User        User      `json:"user"`
	Asset       Asset     `json:"asset"`
	StartTime   time.Time `json:"start_timing"` //args[1]
	EndTime     time.Time `json:"end_timing"`   //args[2]
	Type        string    `json:"doc_type"`
	Priority    uint8     `json:"priority"` //0=low, 1=medium, 2=high //args[3]
	Status      uint8     `json:"status"`   //0=waiting, 1=approved(handover_pending), 2=expired, 3=rejected, 4=completed, 5=cancelled, 6=handover_done, 7=handover_cancelled //args[4]
	IsDeleted   bool      `json:"is_deleted"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

//User ... User Struct
type User struct {
	ID        string `json:"id"`         //args[5]
	FirstName string `json:"first_name"` //args[6]
	LastName  string `json:"last_name"`  //args[7]
}

//Asset ... Asset Struct
type Asset struct {
	ID          string `json:"id"`          //args[8]
	Name        string `json:"name"`        //args[9]
	Description string `json:"description"` //args[10]
}

type Response struct {
	Status  uint16          `json:"status"`
	Message string          `json:"message"`
	Data    RequestResponse `json:"data"`
}

type RequestResponse struct {
	ID      string  `json:"id"`
	Request Request `json:"request"`
}

//CreateRequest ... This function will create a new request
func CreateRequest(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 12 {
		return shim.Error("Incorrect number of arguments. Expecting 12")
	}

	var id = args[11]
	status, statusConvErr := strconv.ParseUint(args[4], 10, 8)
	startTiming, startTimeConvErr := jodaTime.Parse("dd/MM/yyyy HH:mm:ss", args[1])
	endTiming, endTimeConvError := jodaTime.Parse("dd/MM/yyyy HH:mm:ss", args[2])
	priority, priorityConvError := strconv.ParseUint(args[3], 10, 8)

	//handle conversion error here
	if statusConvErr != nil {
		return shim.Error(statusConvErr.Error())
	} else if startTimeConvErr != nil {
		return shim.Error(startTimeConvErr.Error())
	} else if endTimeConvError != nil {
		return shim.Error(endTimeConvError.Error())
	} else if priorityConvError != nil {
		return shim.Error(priorityConvError.Error())
	} else if !startTiming.After(time.Now().UTC()) {
		return shim.Error("Please select valid time slot")
	} else if !endTiming.After(startTiming) || startTiming == endTiming {
		return shim.Error("Please select valid time range")
	}

	user := User{ID: args[5], FirstName: args[6], LastName: args[7]}
	asset := Asset{ID: args[8], Name: args[9], Description: args[10]}

	var request = Request{
		Description: args[0],
		User:        user,
		Asset:       asset,
		StartTime:   startTiming,
		EndTime:     endTiming,
		Type:        "request",
		Priority:    uint8(priority),
		Status:      uint8(status),
		IsDeleted:   false,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC()}

	requestByte, marshalErr := json.Marshal(request)

	if marshalErr != nil {
		return shim.Error(marshalErr.Error())
	}

	saveErr := stub.PutState(id, requestByte)

	if saveErr != nil {
		return shim.Error(saveErr.Error())
	}

	return shim.Success([]byte("{\"status\":1,\"message\":\"Request has been submitted successfully\"}"))
}

//ListAllRequest ... this function will return the list of all Request
func ListAllRequest(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	var sort = "\"sort\":[{\"created_at\":\"desc\"}]"

	if len(args) == 2 {
		var sortType = "asc"

		if args[1] == "-1" {
			sortType = "desc"
		}

		var fieldType = ""

		if args[0] == "0" {
			fieldType = "user.first_name"
		} else if args[0] == "1" {
			fieldType = "user.id"
		}

		sort = "\"sort\":[{\"" + fieldType + "\":\"" + sortType + "\"}]"
	}

	resultsIterator, err := stub.GetQueryResult("{\"selector\":{\"doc_type\":\"request\"}," + sort + "}")
	if err != nil {
		return shim.Error(err.Error())
	}

	var buffer bytes.Buffer
	buffer.WriteString("{\"requests\":[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		if err != nil {
			return shim.Error(err.Error())
		}

		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"id\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"value\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]}")

	return shim.Success(buffer.Bytes())
}

//GetRequestsForAsset ... This function will fetch all requests for a particular asset
func GetRequestsForAsset(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1 argument (Asset id)")
	}

	resultsIterator, err := stub.GetQueryResult("{\"selector\":{\"doc_type\":\"request\",\"asset\":{\"id\":\"" + args[0] + "\"}}}")

	if err != nil {
		return shim.Error(err.Error())
	}

	assetIterator, assetErr := stub.GetQueryResult("{\"selector\":{\"doc_type\":\"assets\",\"_id\":\"" + args[0] + "\"}}")

	if assetErr != nil {
		return shim.Error(assetErr.Error())
	}

	if assetIterator.HasNext() {
		asset, nextErr := assetIterator.Next()

		if nextErr != nil {
			return shim.Error(nextErr.Error())
		}

		var buffer bytes.Buffer
		buffer.WriteString("{\"asset\":" + string(asset.Value) + ",\"requests\":[")

		bArrayMemberAlreadyWritten := false
		for resultsIterator.HasNext() {
			queryResponse, err := resultsIterator.Next()
			if err != nil {
				return shim.Error(err.Error())
			}

			if err != nil {
				return shim.Error(err.Error())
			}

			// Add a comma before array members, suppress it for the first array member
			if bArrayMemberAlreadyWritten == true {
				buffer.WriteString(",")
			}

			buffer.WriteString("{\"id\":")
			buffer.WriteString("\"")
			buffer.WriteString(queryResponse.Key)
			buffer.WriteString("\"")

			buffer.WriteString(", \"value\":")
			// Record is a JSON object, so we write as-is
			buffer.WriteString(string(queryResponse.Value))
			buffer.WriteString("}")
			bArrayMemberAlreadyWritten = true
		}
		buffer.WriteString("]}")

		return shim.Success(buffer.Bytes())
	}

	return shim.Error("Asset not found")

}

//GetRequestByID ... This function will return a particular request
func GetRequestByID(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments.Expected 1 argument")
	}

	resultsIterator, err := stub.GetQueryResult("{\"selector\":{\"doc_type\":\"request\",\"_id\":\"" + args[0] + "\"}}")
	if err != nil {
		return shim.Error(err.Error())
	}

	if resultsIterator.HasNext() {
		resultQuery, resultError := resultsIterator.Next()

		if resultError != nil {
			return shim.Error(resultError.Error())
		}

		var req Request

		unMarshalError := json.Unmarshal(resultQuery.Value, &req)

		if unMarshalError != nil {
			return shim.Error(unMarshalError.Error())
		}

		requestResponse := RequestResponse{ID: resultQuery.Key, Request: req}
		response := Response{Status: 1, Message: "success", Data: requestResponse}

		responseByte, resConvErr := json.Marshal(response)

		if resConvErr != nil {
			return shim.Error(resConvErr.Error())
		}

		return shim.Success(responseByte)

	}
	return shim.Error("Could not find any request.")

}

//ChangeRequestStatus ... This function will change the status of request
func ChangeRequestStatus(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 2 {
		return shim.Error("Invalid number of arguments. Expected 2 arguments.")
	}

	status, statusParseErr := strconv.ParseUint(args[1], 10, 8)

	if statusParseErr != nil {
		return shim.Error(statusParseErr.Error())
	}

	if status < 0 || status > 7 {
		return shim.Error("Please enter a valid request status")
	}

	resultsIterator, iteratorErr := stub.GetQueryResult("{\"selector\":{\"doc_type\":\"request\",\"_id\":\"" + args[0] + "\"}}")

	if iteratorErr != nil {
		return shim.Error(iteratorErr.Error())
	}

	if resultsIterator.HasNext() {

		result, resultErr := resultsIterator.Next()

		if resultErr != nil {
			return shim.Error(resultErr.Error())
		}

		var request Request

		reqUnmarshalErr := json.Unmarshal(result.Value, &request)

		if reqUnmarshalErr != nil {
			return shim.Error(reqUnmarshalErr.Error())
		}

		request.UpdatedAt = time.Now().UTC()
		request.Status = uint8(status)

		reqBytes, reqMarshalErr := json.Marshal(request)

		if reqMarshalErr != nil {
			return shim.Error(reqMarshalErr.Error())
		}

		updateErr := stub.PutState(args[0], reqBytes)

		if updateErr != nil {
			return shim.Error(updateErr.Error())
		}

		return shim.Success([]byte("{\"status\":1,\"message\":\"Request has been successfully updated.\"}"))

	}

	return shim.Error("Request does not exist. Please make sure you are using correct id.")
}

//UpdateTimeSlot ... This function will update the timeslot of a request
func UpdateTimeSlot(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 3 {
		return shim.Error("Invalid number of arguments. Expected 3 arguments.")
	}

	var arg = []string{args[0]}
	requestByID := GetRequestByID(stub, arg)

	if requestByID.Status == 500 {
		return shim.Error("Could not find the request. Please make sure you are using correct request id.")
	}

	startTimeSlot, startTimeErr := jodaTime.Parse("dd/MM/yyyy HH:mm:ss", args[1])
	endTimeSlot, endTimeErr := jodaTime.Parse("dd/MM/yyyy HH:mm:ss", args[2])

	if startTimeErr != nil {
		return shim.Error(startTimeErr.Error())
	} else if endTimeErr != nil {
		return shim.Error(endTimeErr.Error())
	} else if !endTimeSlot.After(startTimeSlot) {
		return shim.Error("Invalid timeslot entered.")
	}

	if !startTimeSlot.After(time.Now().UTC()) {
		return shim.Error("Invalid start time.")
	}

	var responsePayload Response
	unMarshalErr := json.Unmarshal(requestByID.Payload, &responsePayload)

	if unMarshalErr != nil {
		return shim.Error(unMarshalErr.Error())
	}

	request := responsePayload.Data.Request

	request.StartTime = startTimeSlot
	request.EndTime = endTimeSlot
	request.UpdatedAt = time.Now().UTC()

	requestByte, marshalErr := json.Marshal(request)

	if marshalErr != nil {
		return shim.Error(marshalErr.Error())
	}

	updateErr := stub.PutState(arg[0], requestByte)

	if updateErr != nil {
		return shim.Error(updateErr.Error())
	}

	return shim.Success([]byte("{\"status\":1,\"message\":\"Timeslot has been updated\"}"))
}

//GetNextImmediateRequest ... This function will return the next immediate request for particular asset
func GetNextImmediateRequest(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expected 1 argument.(asset_id)")
	}

	now := time.Now()

	startTime := jodaTime.Format("yyyy-MM-ddTHH:mm:ssZ", now)
	endTime := jodaTime.Format("yyyy-MM-ddTHH:mm:ssZ", now.Add(6*time.Hour))

	queryIterator, queryErr := stub.GetQueryResult("{\"selector\":{\"doc_type\":\"request\",\"asset\":{\"id\":\"" + args[0] + "\"},\"$and\":[{\"start_timing\":{\"$gte\":\"" + startTime + "\"}},{\"start_timing\":{\"$lte\":\"" + endTime + "\"}}],\"status\":1},\"sort\":[{\"start_timing\":\"asc\"}]}")

	if queryErr != nil {
		return shim.Error(queryErr.Error())
	}

	var buffer bytes.Buffer
	buffer.WriteString("{\"request\":{")

	if queryIterator.HasNext() {
		queryResponse, err := queryIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		buffer.WriteString("\"id\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"value\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
	}

	buffer.WriteString("}}")

	return shim.Success(buffer.Bytes())

}

//GetAssetsByUserID ... This function will return the list of assets that were related with the provided user id
func GetAssetsByUserID(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expected 1 argument.")
	}

	queryIterator, queryErr := stub.GetQueryResult("{\"selector\":{\"doc_type\":\"request\",\"user\":{\"id\":\"" + args[0] + "\"},\"status\":6}}")

	if queryErr != nil {
		return shim.Error(queryErr.Error())
	}

	var buffer bytes.Buffer
	buffer.WriteString("{\"requests\":[")

	bArrayMemberAlreadyWritten := false
	for queryIterator.HasNext() {
		queryResponse, err := queryIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"id\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"value\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]}")

	return shim.Success(buffer.Bytes())
}

//SortRequests ... This function will sort requests based on input field types and sort type
func SortRequests(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expected 2 argument.(field_type, sort_type)")
	}

	var sortType = "asc"

	if args[1] == "-1" {
		sortType = "desc"
	}

	var fieldType = ""

	if args[0] == "0" {
		fieldType = "user.first_name"
	} else if args[0] == "1" {
		fieldType = "user.id"
	}

	resultsIterator, err := stub.GetQueryResult("{\"selector\":{\"doc_type\":\"request\"},\"sort\":[{\"" + fieldType + "\":\"" + sortType + "\"}]}")
	if err != nil {
		return shim.Error(err.Error())
	}

	var buffer bytes.Buffer
	buffer.WriteString("{\"requests\":[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		if err != nil {
			return shim.Error(err.Error())
		}

		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"id\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"value\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]}")

	return shim.Success(buffer.Bytes())
}
