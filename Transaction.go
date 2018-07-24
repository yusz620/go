package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/satori/go.uuid"
	//log "log"
	"strconv"
	time "time"
)

const (
	dateBase = "2006-01-02 15:04:05" //当前时间的字符串，2006-01-02 15:04:05据说是golang的诞生时间，固定写法
)

type Transaction struct {
}
type TransactionItem struct {
	AccountId      string `json:"AccountId"`
	AccountName    string `json:"AccountName"`
	AccountBalance int    `json:"AccountBalance"`
	TradDate       string `json:"TradDate"`
}

func (t *Transaction) Invoke(APIstub shim.ChaincodeStubInterface) pb.Response {
	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()

	// Route to the appropriate handler function to interact with the ledger appropriately
	if function == "Init" {
		return t.Init(APIstub)
	} else if function == "Create" {
		return t.create(APIstub, args)
	} else if function == "queryTransactionItems" {
		return t.queryTransactionItems(APIstub)
	} else if function == "queryByAccountName" {
		return t.queryByAccountName(APIstub, args)
	}

	return shim.Error("Invalid Smart Contract function name.")
}
func (t *Transaction) Init(stub shim.ChaincodeStubInterface) pb.Response {

	return shim.Success(nil)
}

func (t *Transaction) create(APIstub shim.ChaincodeStubInterface, args []string) pb.Response {
	//fmt.Println("Start to init" + strconv.Atoi(time.Now().Unix()))
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	uuid, err := uuid.NewV4()
	uuids := uuid.String()
	tradDate := time.Now().Format(dateBase)
	AccountName := args[0]
	//AccountBalance := strconv.Atoi(args[1])
	AccountBalance, err := strconv.Atoi(args[1])
	if err != nil {
		return shim.Error("Expecting integer value for asset holding")
	}
	//var transactionItem = TransactionItem{AccountId: uuid, AccountName: AccountName, AccountBalance: AccountBalance}
	var transactionItem = TransactionItem{AccountId: uuids, AccountName: AccountName, AccountBalance: AccountBalance, TradDate: tradDate}

	transactionItemAsBytes, _ := json.Marshal(transactionItem)
	APIstub.PutState(args[0], transactionItemAsBytes)

	return shim.Success(nil)
}

func (t *Transaction) queryTransactionItems(APIstub shim.ChaincodeStubInterface) pb.Response {

	startKey := "T1"
	endKey := "T9999999"

	resultsIterator, err := APIstub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

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
		buffer.WriteString("{\"transaction\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"transactionItem\":[")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("]")
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- queryAllTransactionItems:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

func (t *Transaction) queryByAccountName(APIstub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	itemAsBytes, _ := APIstub.GetState(args[0])
	item := TransactionItem{}

	json.Unmarshal(itemAsBytes, &item)

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[{\"AccountName\":")
	buffer.WriteString("\"")
	buffer.WriteString(args[0])
	buffer.WriteString("\",\"TransactionItem\":[")
	buffer.WriteString(string(itemAsBytes))
	buffer.WriteString("]}]")

	return shim.Success(buffer.Bytes())
}

func main() {

	// Create a new Smart Contract
	err := shim.Start(new(Transaction))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
