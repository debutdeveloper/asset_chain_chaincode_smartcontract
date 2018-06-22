package assets

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/debut_asset_chaincode/response"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
	"github.com/vjeantet/jodaTime"
)

type Assets struct {
	Name           string      `json:"name"`
	AssetNumber    uint32      `json:"asset_number"`
	SerialNumber   string      `json:"serial_number"`
	CategoryID     interface{} `json:"category"`
	Color          string      `json:"color"`
	Type           string      `json:"doc_type"`
	Description    string      `json:"description"`
	Quantity       uint8       `json:"quantity"`
	PurchaseDate   time.Time   `json:"purchase_date"`
	WarrantyUpto   time.Time   `json:"warranty_upto"`
	PurchaseAmount float64     `json:"purchasing_amount"`
	Status         uint8       `json:"status"`
	AssignedTo     User        `json:"assigned_to"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
	IsDeleted      bool        `json:"is_deleted"`
	DeletedAt      time.Time   `json:"deleted_at"`
}

//CategoryID ... structure for category
type CategoryID struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AssetByIdResponse struct {
	ID      string `json:"id"`
	Request Assets `json:"asset"`
}

type Response struct {
	Status  int8              `json:"status"`
	Message string            `json:"message"`
	Data    AssetByIdResponse `json:"data"`
}

//CreateAsset ... this function is used to create assests
func CreateAsset(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 13 {
		return shim.Error("Incorrect number of arguments. Expecting 13")
	}
	resultsIterator, err1 := stub.GetQueryResult("{\"selector\":{\"doc_type\":\"category\",\"_id\":\"" + args[10] + "\"}}")
	if err1 != nil {
		return shim.Error(err1.Error())
	}
	if !resultsIterator.HasNext() {
		return shim.Error("cateogary id is not correct.")
	}
	var nameArr = []string{args[0]}
	var nameValidateResponse = CheckName(stub, nameArr)
	if nameValidateResponse.Status == 500 {
		return nameValidateResponse
	}
	var uniqueID = args[12]
	status, _ := strconv.ParseUint(args[2], 10, 8)
	assetnumber, _ := strconv.ParseUint(args[3], 10, 32)
	quantity, _ := strconv.ParseUint(args[6], 10, 8)
	purchasedate, _ := jodaTime.Parse("dd/MM/yyyy", args[7])
	warrantyupto, _ := jodaTime.Parse("dd/MM/yyyy", args[8])
	purchaseamount, _ := strconv.ParseFloat(args[9], 64)
	categoryid := CategoryID{ID: args[10], Name: args[11]}

	var admin = User{ID: "1wdcmjpjhrc6bv4", Name: "Admin"}

	var assets = Assets{CategoryID: categoryid,
		Name:           args[0],
		Description:    args[1],
		Type:           "assets",
		Status:         uint8(status),
		AssignedTo:     admin,
		AssetNumber:    uint32(assetnumber),
		SerialNumber:   args[4],
		Color:          args[5],
		Quantity:       uint8(quantity),
		PurchaseDate:   purchasedate,
		WarrantyUpto:   warrantyupto,
		PurchaseAmount: purchaseamount,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC()}
	assetAsBytes, _ := json.Marshal(assets)

	err := stub.PutState(uniqueID, assetAsBytes)

	if err != nil {
		return shim.Error(response.CreateErrorResponse(err.Error(),
			0,
			nil))
	}

	return shim.Success(response.CreateSuccessResponse("Asset created successfully",
		1,
		nil))
}

//GetAssetByID ... This function will return a particular asset
func GetAssetByID(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments.Expected 1 argument")
	}

	resultsIterator, err := stub.GetQueryResult("{\"selector\":{\"doc_type\":\"assets\",\"_id\":\"" + args[0] + "\"}}")
	if err != nil {
		return shim.Error(err.Error())
	}

	if resultsIterator.HasNext() {
		resultQuery, resultError := resultsIterator.Next()

		if resultError != nil {
			return shim.Error(resultError.Error())
		}

		var asset Assets

		unMarshalError := json.Unmarshal(resultQuery.Value, &asset)

		if unMarshalError != nil {
			return shim.Error(unMarshalError.Error())
		}

		assetResponse := AssetByIdResponse{ID: resultQuery.Key, Request: asset}
		response := Response{Status: 1, Message: "success", Data: assetResponse}

		responseByte, resConvErr := json.Marshal(response)

		if resConvErr != nil {
			return shim.Error(resConvErr.Error())
		}

		return shim.Success(responseByte)

	}
	return shim.Error("Could not find any assets.")

}

func CheckName(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		return shim.Error("Invalid number of arguments. Expected 1 argument.")
	} else if len(args[0]) == 0 {
		return shim.Error("Please enter a valid email.")
	}

	resultsIterator, err := stub.GetQueryResult("{\"selector\":{\"is_deleted\":false,\"doc_type\":\"assets\",\"name\":\"" + args[0] + "\"}}")
	if err != nil {
		return shim.Error(err.Error())
	}

	if resultsIterator.HasNext() {
		return shim.Error("The asset name you have entered already exist.")
	}

	return shim.Success(response.CreateSuccessResponse("Category Updated successfully",
		1,
		nil))

}

//ChangeAssetStatus ... This function will change the status of asset
func ChangeAssetStatus(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 2 {
		return shim.Error("Invalid number of arguments. Expected 2 argument. (asset_id & status)")
	}

	resultsIterator, err1 := APIstub.GetQueryResult("{\"selector\":{\"doc_type\":\"category\",\"_id\":\"" + args[10] + "\"}}")

	if err1 != nil {
		return shim.Error(err1.Error())
	}

	if !resultsIterator.HasNext() {
		return shim.Error("cateogary id is not correct.")
	}

	assetByte, nextErr := resultsIterator.Next()

	if nextErr != nil {
		return shim.Error(nextErr.Error())
	}

	status, intParseErr := strconv.ParseInt(args[1], 10, 8)

	if intParseErr != nil {
		return shim.Error(intParseErr.Error())
	}

	var asset Assets
	unMarshalErr := json.Unmarshal(assetByte.Value, &asset)

	if unMarshalErr != nil {
		return shim.Error(unMarshalErr.Error())
	}

	asset.Status = uint8(status)

	marshalledAsset, marshalErr := json.Marshal(asset)

	if marshalErr != nil {
		return shim.Error(marshalErr.Error())
	}

	putError := APIstub.PutState(args[0], marshalledAsset)

	if putError != nil {
		return shim.Error(putError.Error())
	}

	return shim.Success([]byte("{\"status\":1, \"message\":\"Status updated successfully\"}"))
}

//UpdateAsset ... this function will update the Asset
func UpdateAsset(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 13 {
		return shim.Error("Incorrect number of arguments. Expecting 13")
	}
	resultsIterator, err1 := APIstub.GetQueryResult("{\"selector\":{\"doc_type\":\"category\",\"_id\":\"" + args[10] + "\"}}")
	if err1 != nil {
		return shim.Error(err1.Error())
	}
	if !resultsIterator.HasNext() {
		return shim.Error("cateogary id is not correct.")
	}

	var unique_id = args[12]
	status, _ := strconv.ParseUint(args[2], 10, 8)
	assetnumber, _ := strconv.ParseUint(args[3], 10, 32)
	quantity, _ := strconv.ParseUint(args[6], 10, 8)
	purchasedate, _ := jodaTime.Parse("dd/MM/yyyy", args[7])
	warrantyupto, _ := jodaTime.Parse("dd/MM/yyyy", args[8])
	purchaseamount, _ := strconv.ParseFloat(args[9], 64)
	categoryid := CategoryID{ID: args[10], Name: args[11]}
	var assets = Assets{CategoryID: categoryid,
		Name:           args[0],
		Description:    args[1],
		Type:           "assets",
		Status:         uint8(status),
		AssetNumber:    uint32(assetnumber),
		SerialNumber:   args[4],
		Color:          args[5],
		Quantity:       uint8(quantity),
		PurchaseDate:   purchasedate,
		WarrantyUpto:   warrantyupto,
		PurchaseAmount: purchaseamount,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC()}
	assetAsBytes, _ := json.Marshal(assets)

	err := APIstub.PutState(unique_id, assetAsBytes)
	if err != nil {
		return shim.Error(response.CreateErrorResponse(err.Error(),
			0,
			nil))
	}

	return shim.Success(response.CreateSuccessResponse("Asset updated successfully",
		1,
		nil))
}

//ListAllAssets ... this function will return the list of all Assets
func ListAllAssets(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	var sort = "\"sort\":[{\"created_at\":\"desc\"}]"

	if len(args) == 2 {
		var sortType = "asc"

		if args[1] == "-1" {
			sortType = "desc"
		}

		var fieldType = ""

		if args[0] == "0" {
			fieldType = "name"
		}

		sort = "\"sort\":[{\"" + fieldType + "\":\"" + sortType + "\"}]"
	}
	resultsIterator, err := stub.GetQueryResult("{\"selector\":{\"is_deleted\":false,\"doc_type\":\"assets\"}," + sort + "}")

	if err != nil {
		fmt.Println("Error occured ==> ", err)
		return shim.Error(err.Error())
	}

	var buffer bytes.Buffer
	buffer.WriteString("{\"assets\":[")

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

//DeleteAsset ... this function will delete the assets
func DeleteAsset(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error("Invalid number of arguments. Expected 1 arguments")
	}

	assetAsBytes, _ := APIstub.GetState(args[0])
	asset := Assets{}
	json.Unmarshal(assetAsBytes, &asset)
	asset.IsDeleted = true
	assetAsBytes, _ = json.Marshal(asset)
	err := APIstub.PutState(args[0], assetAsBytes)
	if err != nil {
		shim.Error(response.CreateErrorResponse(err.Error(),
			0,
			nil))
	}

	return shim.Success(response.CreateSuccessResponse("Asset deleted successfully",
		1,
		nil))
}

//AssignAsset ... This function will update asset assigned status
func AssignAsset(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 4 {
		return shim.Error("Invalid number of arguments. Expected 4 arguments (asset_id, user_id, first_name, last_name)")
	}

	resultsIterator, err := stub.GetQueryResult("{\"selector\":{\"is_deleted\":false,\"doc_type\":\"assets\",\"_id\":\"" + args[0] + "\"}}")

	if err != nil {
		return shim.Error(err.Error())
	}

	if resultsIterator.HasNext() {
		result, resultErr := resultsIterator.Next()

		if resultErr != nil {
			return shim.Error(resultErr.Error())
		}

		var asset Assets

		unMarshalErr := json.Unmarshal(result.Value, &asset)

		if unMarshalErr != nil {
			return shim.Error(unMarshalErr.Error())
		}

		var user = User{ID: args[1], Name: args[2] + " " + args[3]}

		asset.AssignedTo = user

		marshalledAsset, marshalErr := json.Marshal(asset)

		if marshalErr != nil {
			return shim.Error(marshalErr.Error())
		}

		putStateErr := stub.PutState(result.Key, marshalledAsset)

		if putStateErr != nil {
			return shim.Error(putStateErr.Error())
		}

		return shim.Success([]byte("{\"message\":\"Assigned status changed successfully\", \"status\":\"1\"}"))
	}

	return shim.Error("Could not find any asset with input asset id.")

}
