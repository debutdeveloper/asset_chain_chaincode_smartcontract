package user

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
	"github.com/shomali11/util/xhashes"
)

type User struct {
	FirstName        string    `json:"first_name"`
	LastName         string    `json:"last_name"`
	DocType          string    `json:"doc_type"`
	EmpID            string    `json:"emp_id"`
	Password         string    `json:"password"` //will be saved in hash
	CountryCode      string    `json:"country_code"`
	PhoneNumber      uint64    `json:"phone_number"`
	Email            string    `json:"email"`
	Role             uint8     `json:"role"`               //0=Normal User, 1=Admin User
	DeviceToken      string    `json:"device_token"`       //for push notifications
	Status           uint8     `json:"status"`             //1 = active, 0 = inactive
	IsRandomPassword bool      `json:"is_random_password"` // first time we will generate a random password
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	IsDeleted        bool      `json:"is_deleted"`
	DeletedAt        time.Time `json:"deleted_at"`
}

type AllUsersResponse struct {
	Users []UserData `json:"users"`
}

type UserResponse struct {
	ID                string `json:"user_id"`
	FirstName         string `json:"first_name"`
	LastName          string `json:"last_name"`
	Email             string `json:"email"`
	GeneratedPassword string `json:"generated_password"`
}

type Response struct {
	Status  int32       `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type LoginResponse struct {
	Status  int8        `json:"status"`
	Message string      `json:"message"`
	Id      string      `json:"id"`
	Data    interface{} `json:"data"`
}

type UserData struct {
	ID   string `json:"id"`
	User User   `json:"user"`
}

type UsersResponse struct {
	Users []UserData `json:"users"`
}
type ForgetPasswordResponse struct {
	NewPassword string `json:"new_password"`
}

/************************************************
************** STRUCTS FOR QUERY ****************
************************************************/
type QueryStruct struct {
	Selector interface{} `json:"selector"`
}

type LoginQuery struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Type     uint8  `json:"role"`
	Status   uint8  `json:"status"`
}

type EmailQuery struct {
	Email string `json:"email"`
}

type EmpIDQuery struct {
	EmployeeId string `json:"emp_id"`
}

type PhoneQuery struct {
	Phone uint64 `json:"phone_number"`
}

type IdQuery struct {
	ID string `json:"_id"`
}

//AddUser ... This function will add a new user to database
func AddUser(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 9 {
		return createErrorResponse("Invalid number of arguments. Expected 9", 422, nil)
	} else if len(args[4]) == 0 { //Phone number validation
		return createErrorResponse("Please enter a valid phone number", 422, nil)
	}

	var emailArr = []string{args[5]}
	var phoneArr = []string{args[4]}
	var empID = []string{args[2]}

	var emailValidateResponse = CheckEmail(stub, emailArr)
	var phoneValidateResponse = CheckPhone(stub, phoneArr)
	var empIDValidateResponse = CheckEmployeeID(stub, empID)

	if emailValidateResponse.Status == 500 {
		return emailValidateResponse
	} else if phoneValidateResponse.Status == 500 {
		return phoneValidateResponse
	} else if empIDValidateResponse.Status == 500 {
		return empIDValidateResponse
	}

	var randomPassword = args[8]
	var hashedPassword = xhashes.SHA256(randomPassword)
	phoneNumber, phoneConvError := strconv.ParseUint(args[4], 10, 64)

	timeNow := time.Now().UTC()
	id := args[7]

	if phoneConvError != nil {
		return shim.Error(phoneConvError.Error())
	}

	user := User{
		FirstName:        args[0],
		LastName:         args[1],
		EmpID:            args[2],
		DocType:          "user",
		Password:         hashedPassword,
		CountryCode:      args[3],
		PhoneNumber:      phoneNumber,
		Email:            args[5],
		Role:             0,
		DeviceToken:      args[6],
		Status:           1,
		IsRandomPassword: true,
		CreatedAt:        timeNow,
		UpdatedAt:        timeNow}

	userBytes, err := json.Marshal(user)

	if err != nil {
		return shim.Error(err.Error())
	}

	putStateError := stub.PutState(id, userBytes)

	fmt.Println("User Record => " + string(userBytes))
	fmt.Println("Hashed Password ==> " + hashedPassword)
	if putStateError != nil {
		return shim.Error(putStateError.Error())
	}

	userResponse := UserResponse{
		ID:                id,
		GeneratedPassword: randomPassword,
		Email:             args[5],
		FirstName:         args[0],
		LastName:          args[1]}

	response := Response{
		Status:  1,
		Message: "User created successfully",
		Data:    userResponse}

	responseBytes, respError := json.Marshal(response)

	if respError != nil {
		return shim.Error(respError.Error())
	}

	fmt.Println("Hashed Last Password ==> " + hashedPassword)
	return shim.Success(responseBytes)
}

//Login ... This function will login user
func Login(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 4 {
		return createErrorResponse("Incorrect number of arguments. Expected 4 arguments", 422, nil)
	}

	email := args[0]
	password := xhashes.SHA256(args[1])
	reqType, typeConvError := strconv.ParseUint(args[3], 10, 8)

	if typeConvError != nil {
		return shim.Error(typeConvError.Error())
	}

	loginQuery := LoginQuery{Email: email, Password: password, Type: uint8(reqType), Status: 1}
	query := QueryStruct{Selector: loginQuery}

	queryByte, err := json.Marshal(query)

	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("Query ==> " + string(queryByte))

	queryIterator, queryError := stub.GetQueryResult(string(queryByte))

	if queryError != nil {
		return shim.Error(queryError.Error())
	}

	if queryIterator.HasNext() {
		queryResult, _ := queryIterator.Next()

		var user User

		err := json.Unmarshal(queryResult.Value, &user)

		if err != nil {
			return shim.Error(err.Error())
		}

		if user.Status == 0 {
			return createErrorResponse("You have been disabled by admin.", 401, nil)
		}

		user.DeviceToken = args[2]

		userBytes, userConvError := json.Marshal(user)

		if userConvError != nil {
			return shim.Error(userConvError.Error())
		}

		userUpdateErr := stub.PutState(queryResult.Key, userBytes)

		if userUpdateErr != nil {
			return shim.Error(userUpdateErr.Error())
		}

		//return shim.Success(queryResult.Value)

		//send user data with id on successfull login
		loginResponse := LoginResponse{
			Status:  1,
			Message: "User login successful",
			Id:      queryResult.Key,
			Data:    user}

		responseBytes, respError := json.Marshal(loginResponse)

		if respError != nil {
			return createErrorResponse(respError.Error(), 500, nil)
		}

		return shim.Success(responseBytes)
	}
	return createErrorResponse("The email or password you have entered is wrong.", 401, nil)
}

//GetUser ... This function will fetch the details of a particular user
func GetUser(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		return createErrorResponse("Invalid number of arguments. Expected 1 argument (user id)", 422, nil)
	}

	selectorQuery, queryErr := stub.GetQueryResult("{\"selector\":{\"is_deleted\":false,\"doc_type\":\"user\",\"_id\":\"" + args[0] + "\"}}")

	if queryErr != nil {
		return createErrorResponse(queryErr.Error(), 500, nil)
	}

	if selectorQuery.HasNext() {
		queryResult, err := selectorQuery.Next()

		if err != nil {
			return createErrorResponse(err.Error(), 500, nil)
		}

		var user User
		unMarshalErr := json.Unmarshal(queryResult.Value, &user)

		if unMarshalErr != nil {
			return createErrorResponse(unMarshalErr.Error(), 500, nil)
		}

		user.Password = ""

		return shim.Success(createSuccessResponse("success", 200, user))
	}

	return createErrorResponse("User does not exist", 404, nil)
}

//ListAllUsers ... This function will list all the users
func ListAllUsers(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	var sort = "\"sort\":[{\"created_at\":\"desc\"}]"

	if len(args) == 2 {
		var sortType = "asc"

		if args[1] == "-1" {
			sortType = "desc"
		}

		var fieldType = ""

		if args[0] == "0" {
			fieldType = "first_name"
		} else if args[0] == "1" {
			fieldType = "last_name"
		} else if args[0] == "2" {
			fieldType = "email"
		}

		sort = "\"sort\":[{\"" + fieldType + "\":\"" + sortType + "\"}]"
	}
	selectorQuery, queryErr := stub.GetQueryResult("{\"selector\":{\"is_deleted\":false,\"doc_type\":\"user\"}," + sort + "}")

	if queryErr != nil {
		return shim.Error(queryErr.Error())
	}

	users := make([]UserData, 0, 1)

	for selectorQuery.HasNext() {
		queryResult, queryResultErr := selectorQuery.Next()

		if queryResultErr != nil {
			return shim.Error(queryResultErr.Error())
		}

		var user User

		unMarshalError := json.Unmarshal(queryResult.Value, &user)

		if unMarshalError != nil {
			return shim.Error(unMarshalError.Error())
		}

		user.Password = ""

		userData := UserData{ID: queryResult.Key, User: user}

		users = append(users, userData)
	}

	response := UsersResponse{Users: users}

	responseByte, err := json.Marshal(response)

	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(responseByte)

}

//CheckEmail ... This function will check whether user email exist in database or not
func CheckEmail(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		return shim.Error("Invalid number of arguments. Expected 1 argument.")
	} else if len(args[0]) == 0 {
		return shim.Error("Please enter a valid email.")
	}

	emailStruct := EmailQuery{Email: args[0]}
	selector := QueryStruct{Selector: emailStruct}

	selectorByte, err := json.Marshal(selector)

	if err != nil {
		return shim.Error(err.Error())
	}

	queryIterator, queryError := stub.GetQueryResult(string(selectorByte))

	if queryError != nil {
		return shim.Error(queryError.Error())
	}

	if queryIterator.HasNext() {
		return shim.Error("The email you have entered already exist.")
	}

	return shim.Success(createSuccessResponse("The email you have entered is a valid email.",
		1,
		nil))

}

//CheckEmployeeID ... This function will check whether employee id exist in database or not
func CheckEmployeeID(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		return shim.Error("Invalid number of arguments. Expected 1 argument.")
	} else if len(args[0]) == 0 {
		return shim.Error("Please enter a valid employee id.")
	}

	empStruct := EmpIDQuery{EmployeeId: args[0]}
	selector := QueryStruct{Selector: empStruct}

	selectorByte, err := json.Marshal(selector)

	if err != nil {
		return shim.Error(err.Error())
	}

	queryIterator, queryError := stub.GetQueryResult(string(selectorByte))

	if queryError != nil {
		return shim.Error(queryError.Error())
	}

	if queryIterator.HasNext() {

		return shim.Error("The employee id you have entered already exist.")
	}

	return shim.Success(createSuccessResponse("The employee id you have entered is valid",
		1,
		nil))
}

//ChangeDeleteStatus ... This function will be used for deleting a user
func ChangeDeleteStatus(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 2 {
		return createErrorResponse("Invalid number of arguments. Expected 2 arguments", 422, nil)

	}

	status, boolConvError := strconv.ParseBool(args[1])

	//handle bool conversion error
	if boolConvError != nil {
		return createErrorResponse(boolConvError.Error(), 500, nil)
	}

	idStruct := IdQuery{ID: args[0]}
	selectorStruct := QueryStruct{Selector: idStruct}

	queryByte, byteErr := json.Marshal(selectorStruct)

	//handle json marshal error
	if byteErr != nil {
		return createErrorResponse(byteErr.Error(), 500, nil)
	}

	queryIterator, queryErr := stub.GetQueryResult(string(queryByte))

	//handle query error
	if queryErr != nil {
		return createErrorResponse(queryErr.Error(), 500, nil)
	}

	if !queryIterator.HasNext() {
		return createErrorResponse("Please enter a valid user id", 422, nil)
	}
	queryResult, _ := queryIterator.Next()

	var user User

	err := json.Unmarshal(queryResult.Value, &user)

	//handle json unmarshal error
	if err != nil {
		return shim.Error(queryErr.Error())
	}

	user.IsDeleted = status

	if status {
		user.DeletedAt = time.Now().UTC()
	} else {
		user.DeletedAt = time.Time{}
	}

	//after setting the user delete status convert the struct into []byte to be updated
	//into couchdb
	userBytes, userByteConversionErr := json.Marshal(user)

	if userByteConversionErr != nil {
		return shim.Error(userByteConversionErr.Error())
	}

	updateStateErr := stub.PutState(args[0], userBytes)

	if updateStateErr != nil {
		return shim.Error(updateStateErr.Error())
	}

	return shim.Success(createSuccessResponse("The user status has been successfully deleted",
		1,
		nil))
}

//ChangeUserStatus ... This function will activate or inactivate user
func ChangeUserStatus(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 2 {
		return createErrorResponse("Invalid number of arguments. Expected 2 arguments", 422, nil)
	}

	shouldActive, boolConvError := strconv.ParseBool(args[1])

	//handle bool conversion error
	if boolConvError != nil {
		return createErrorResponse(boolConvError.Error(), 500, nil)
	}

	idStruct := IdQuery{ID: args[0]}
	selectorStruct := QueryStruct{Selector: idStruct}

	queryByte, byteErr := json.Marshal(selectorStruct)

	//handle json marshal error
	if byteErr != nil {
		return createErrorResponse(byteErr.Error(), 500, nil)
	}

	queryIterator, queryErr := stub.GetQueryResult(string(queryByte))

	//handle query error
	if queryErr != nil {
		return createErrorResponse(queryErr.Error(), 500, nil)
	}

	if !queryIterator.HasNext() {
		return createErrorResponse("Please enter a valid user id", 422, nil)
	}

	queryResult, _ := queryIterator.Next()

	var user User

	err := json.Unmarshal(queryResult.Value, &user)

	//handle json unmarshal error
	if err != nil {
		return createErrorResponse(queryErr.Error(), 500, nil)
	}

	user.UpdatedAt = time.Now().UTC()
	user.Status = 1

	if !shouldActive {
		user.Status = 0
	}

	//after setting the user delete status convert the struct into []byte to be updated
	//into couchdb
	userBytes, userByteConversionErr := json.Marshal(user)

	if userByteConversionErr != nil {
		return createErrorResponse(userByteConversionErr.Error(), 0, nil)
	}

	updateStateErr := stub.PutState(args[0], userBytes)

	if updateStateErr != nil {
		return createErrorResponse(updateStateErr.Error(), 0, nil)
	}

	return shim.Success(createSuccessResponse("The user active status has been changed successfully",
		1,
		nil))
}

//UpdateAdminProfile ... Update admin profile
func UpdateAdminProfile(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 3 {
		return createErrorResponse("Invalid number of arguments. Expected 3 argument.",
			0,
			nil)
	}

	queryIterator, queryError := stub.GetQueryResult("{\"selector\":{\"doc_type\":\"user\",\"_id\":\"" + args[0] + "\"}}")

	if queryError != nil {
		return createErrorResponse(queryError.Error(), 500, nil)
	}

	if queryIterator.HasNext() {
		result, resultError := queryIterator.Next()

		if resultError != nil {
			return createErrorResponse(resultError.Error(), 500, nil)
		}

		var user User
		unMarshalErr := json.Unmarshal(result.Value, &user)

		if unMarshalErr != nil {
			return createErrorResponse(unMarshalErr.Error(), 500, nil)
		}

		if user.Role != 1 {
			return createErrorResponse("This functionality is for admin only.", 500, nil)
		}

		user.FirstName = args[1]
		user.LastName = args[2]
		user.UpdatedAt = time.Now().UTC()

		byteData, marshalErr := json.Marshal(user)

		if marshalErr != nil {
			return createErrorResponse(marshalErr.Error(), 500, nil)
		}

		updateErr := stub.PutState(result.Key, byteData)

		if updateErr != nil {
			return createErrorResponse(updateErr.Error(), 500, nil)
		}

		return shim.Success([]byte("{\"message\":\"Profile has been updated successfully.\", \"status\":200,\"first_name\":\"" + user.FirstName + "\",\"last_name\":\"" + user.LastName + "\"}"))
	}

	return createErrorResponse("User not found", 404, nil)
}

//CheckPhone ... This function will check whether phone number exist in database or not
func CheckPhone(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {

		return createErrorResponse("Invalid number of arguments. Expected 1 argument.",
			0,
			nil)

	} else if len(args[0]) == 0 {

		return createErrorResponse("Please enter a valid phone number.",
			0,
			nil)
	}

	phoneInt, err := strconv.ParseUint(args[0], 10, 64)

	if err != nil {
		return createErrorResponse(err.Error(),
			0,
			nil)
	}

	phoneStruct := PhoneQuery{Phone: phoneInt}
	selector := QueryStruct{Selector: phoneStruct}

	selectorByte, err := json.Marshal(selector)

	if err != nil {
		return shim.Error(err.Error())
	}

	queryIterator, queryError := stub.GetQueryResult(string(selectorByte))

	if queryError != nil {
		return shim.Error(queryError.Error())
	}

	if queryIterator.HasNext() {

		return createErrorResponse("The phone number you have entered already exist.",
			0,
			nil)
	}

	return shim.Success(createSuccessResponse("The phone number you have entered is valid",
		1,
		nil))
}

//ChangePassword ... This function will change the user password
func ChangePassword(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Required 3 arguments.")
	}

	query, queryIteratorErr := stub.GetQueryResult("{\"selector\":{\"doc_type\":\"user\",\"_id\":\"" + args[0] + "\"}}")

	if queryIteratorErr != nil {
		return shim.Error(queryIteratorErr.Error())
	}

	if query.HasNext() {
		queryResult, queryErr := query.Next()

		if queryErr != nil {
			return shim.Error(queryErr.Error())
		}

		var user User
		unmarshalErr := json.Unmarshal(queryResult.Value, &user)

		if unmarshalErr != nil {
			return shim.Error(unmarshalErr.Error())
		}

		var oldPasswordHash = xhashes.SHA256(args[1])
		var newPasswordHash = xhashes.SHA256(args[2])

		if user.Password == oldPasswordHash {
			user.Password = newPasswordHash
			user.IsRandomPassword = false

			userBytes, userMarshalError := json.Marshal(user)

			if userMarshalError != nil {
				return shim.Error(userMarshalError.Error())
			}

			putStateErr := stub.PutState(args[0], userBytes)

			if putStateErr != nil {
				return shim.Error(putStateErr.Error())
			}

			return shim.Success(createSuccessResponse("Password has been changed successfully", 1, nil))

		} else {
			return shim.Error("Please enter correct password")
		}

	} else {
		return shim.Error("User does not exist.")
	}

}

//ForgetPassword ... function will be triggered on forget password
func ForgetPassword(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Required 2 arguments.")
	}
	resultsIterator, iteratorerror := stub.GetQueryResult("{\"selector\":{\"doc_type\":\"user\",\"email\":\"" + args[0] + "\"}}")
	if iteratorerror != nil {
		fmt.Println("Error occured ==> ", iteratorerror)
		return shim.Error(iteratorerror.Error())
	}
	if resultsIterator.HasNext() {
		queryResult, queryErr := resultsIterator.Next()
		if queryErr != nil {
			return shim.Error(queryErr.Error())
		}
		var user User
		unmarshalErr := json.Unmarshal(queryResult.Value, &user)

		if unmarshalErr != nil {
			return shim.Error(unmarshalErr.Error())
		}
		randomPassword := args[1]

		hashedPassword := xhashes.SHA256(randomPassword)
		user.Password = hashedPassword
		userBytes, userMarshalError := json.Marshal(user)

		if userMarshalError != nil {
			return shim.Error(userMarshalError.Error())
		}

		putStateErr := stub.PutState(queryResult.Key, userBytes)

		if putStateErr != nil {
			return shim.Error(putStateErr.Error())
		}
		var forgetpassord = ForgetPasswordResponse{NewPassword: randomPassword}

		return shim.Success(createSuccessResponse("Your new password has been sent to your email address.", 1, forgetpassord))
	} else {
		return shim.Error("No Such User")
	}
}

// createErrorResponse ... This function will create error response
func createErrorResponse(message string, status int32, data interface{}) peer.Response {
	response := Response{Message: message,
		Status: status,
		Data:   data}

	responseByte, err := json.Marshal(response)

	var peerResponse = peer.Response{Payload: responseByte, Status: status, Message: message}

	if err != nil {
		peerResponse.Payload = []byte("{\"message\":\"Something went wrong while parsing json\"}")
		peerResponse.Status = 500
	}

	return peerResponse
}

//createSuccessResponse... This function will create success response
func createSuccessResponse(message string, status int32, data interface{}) []byte {
	response := Response{Message: message,
		Status: status,
		Data:   data}

	responseByte, err := json.Marshal(response)

	if err != nil {
		return []byte(err.Error())
	}

	return responseByte
}
