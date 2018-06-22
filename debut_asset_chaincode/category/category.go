package category

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/debut_asset_chaincode/response"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

//Category ... Structure definition for Category
type Category struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        string    `json:"doc_type"`
	Status      uint8     `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	IsDeleted   bool      `json:"is_deleted"`
	DeletedAt   time.Time `json:"deleted_at"`
}

//CategoryResponse ... structure definition for category response
type CategoryResponse struct {
	Categories []Category `json:"categories"`
}

type QueryStruct struct {
	Selector interface{} `json:"selector"`
}
type CategoryUniqueQuery struct {
	Name      string `json:"name"`
	IsDeleted bool   `json:"is_deleted"`
	Type      string `json:"doc_type"`
}

type CategoryByIdResponse struct {
	ID      string   `json:"id"`
	Request Category `json:"request"`
}

type Response struct {
	Status  uint16               `json:"status"`
	Message string               `json:"message"`
	Data    CategoryByIdResponse `json:"data"`
}

//CreateCategory ... this function is used to create category
func CreateCategory(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 4 {
		return shim.Error(response.CreateErrorResponse("Incorrect number of arguments. Expecting 4",
			0,
			nil))
	}
	var nameArr = []string{args[0]}
	var nameValidateResponse = CheckName(stub, nameArr)
	if nameValidateResponse.Status == 500 {
		return nameValidateResponse
	}
	var uniqueID = args[3]
	status, _ := strconv.ParseUint(args[2], 10, 64)

	var category = Category{Name: args[0],
		Description: args[1],
		Type:        "category",
		Status:      uint8(status),
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC()}

	catAsBytes, _ := json.Marshal(category)
	err := stub.PutState(uniqueID, catAsBytes)

	if err != nil {
		shim.Error(response.CreateErrorResponse(err.Error(),
			0,
			nil))
	}

	return shim.Success(response.CreateSuccessResponse("Category created successfully",
		1,
		nil))
}

//CheckEmail ... This function will check whether user email exist in database or not
func CheckName(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		return shim.Error("Invalid number of arguments. Expected 1 argument.")
	} else if len(args[0]) == 0 {
		return shim.Error("Please enter a valid email.")
	}

	resultsIterator, err := stub.GetQueryResult("{\"selector\":{\"is_deleted\":false,\"doc_type\":\"category\",\"name\":\"" + args[0] + "\"}}")
	if err != nil {
		return shim.Error(err.Error())
	}

	if resultsIterator.HasNext() {
		return shim.Error("The name you have entered already exist.")
	}

	return shim.Success(response.CreateSuccessResponse("Category Updated successfully",
		1,
		nil))

}

//ListAllCategory ... this function will return the list of all category
func ListAllCategory(stub shim.ChaincodeStubInterface) peer.Response {

	resultsIterator, err := stub.GetQueryResult("{\"selector\":{\"is_deleted\":false,\"doc_type\":\"category\",\"status\":1}}")
	if err != nil {
		fmt.Println("Error occured ==> ", err)
		return shim.Error(err.Error())
	}

	var buffer bytes.Buffer
	buffer.WriteString("{\"category\":[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]}")

	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(buffer.Bytes())
}

//GetcategoryByID ... This function will return a particular category
func GetCategoryByID(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments.Expected 1 argument")
	}

	resultsIterator, err := stub.GetQueryResult("{\"selector\":{\"doc_type\":\"category\",\"_id\":\"" + args[0] + "\"}}")
	if err != nil {
		return shim.Error(err.Error())
	}

	if resultsIterator.HasNext() {
		resultQuery, resultError := resultsIterator.Next()

		if resultError != nil {
			return shim.Error(resultError.Error())
		}

		var req Category

		unMarshalError := json.Unmarshal(resultQuery.Value, &req)

		if unMarshalError != nil {
			return shim.Error(unMarshalError.Error())
		}

		categoryResponse := CategoryByIdResponse{ID: resultQuery.Key, Request: req}
		response := Response{Status: 1, Message: "success", Data: categoryResponse}

		responseByte, resConvErr := json.Marshal(response)

		if resConvErr != nil {
			return shim.Error(resConvErr.Error())
		}

		return shim.Success(responseByte)

	}
	return shim.Error("Could not find any Category.")

}

//this function will update the category

func UpdateCategory(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 4 {
		return shim.Error("Invalid number of arguments. Expected 4 arguments")
	}
	var unique_id = args[3]
	status, _ := strconv.ParseUint(args[2], 10, 64)

	var category = Category{
		Name:        args[0],
		Description: args[1],
		Type:        "category",
		Status:      uint8(status),
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	catAsBytes, _ := json.Marshal(category)
	err := APIstub.PutState(unique_id, catAsBytes)
	if err != nil {
		shim.Error(response.CreateErrorResponse(err.Error(),
			0,
			nil))
	}

	return shim.Success(response.CreateSuccessResponse("Category Updated successfully",
		1,
		nil))

}

//this function will delete the category
func DeleteCategory(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error("Invalid number of arguments. Expected 1 arguments")
	}
	catAsBytes, _ := APIstub.GetState(args[0])
	cat := Category{}
	json.Unmarshal(catAsBytes, &cat)
	cat.IsDeleted = true
	catAsBytes, _ = json.Marshal(cat)
	err := APIstub.PutState(args[0], catAsBytes)
	if err != nil {
		shim.Error(response.CreateErrorResponse(err.Error(),
			0,
			nil))
	}

	return shim.Success(response.CreateSuccessResponse("Category Updated successfully",
		1,
		nil))
}

//this function will change the status of categoray
func ChangeStatus(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 2 {
		return shim.Error("Invalid number of arguments. Expected 2 arguments")
	}
	catAsBytes, _ := APIstub.GetState(args[0])
	cat := Category{}
	json.Unmarshal(catAsBytes, &cat)
	status, _ := strconv.ParseUint(args[1], 10, 64)
	cat.Status = uint8(status)
	catAsBytes, _ = json.Marshal(cat)
	err := APIstub.PutState(args[0], catAsBytes)
	if err != nil {
		shim.Error(response.CreateErrorResponse(err.Error(),
			0,
			nil))
	}

	return shim.Success(response.CreateSuccessResponse("Category Updated successfully",
		1,
		nil))
}
