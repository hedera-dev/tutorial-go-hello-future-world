package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env")
	}

	accountIdStr := os.Getenv("ACCOUNT_ID")
	accountKeyStr := os.Getenv("ACCOUNT_PRIVATE_KEY")
	accountKeyStr = strings.TrimPrefix(accountKeyStr, "0x")

	// Configure client using environment variables
	accountId, err := hedera.AccountIDFromString(accountIdStr)
	if err != nil {
		log.Fatalf("Error parsing account ID: %v", err)
	}
	accountKey, err := hedera.PrivateKeyFromStringECDSA(accountKeyStr)
	if err != nil {
		log.Fatalf("Error parsing account private key: %v", err)
	}
	client := hedera.ClientForTestnet()
	client.SetOperator(accountId, accountKey)

	accountExplorerUrl := fmt.Sprintf(
		"https://hashscan.io/testnet/address/%s",
		accountId,
	)

	// Read EVM bytecode file for deployment
	evmBytecode, err := os.ReadFile("./my_contract_sol_MyContract.bin")
	if err != nil {
		log.Fatalf("Error reading bytecode file: %v", err)
	}

	// Upload bytecode to HFS, in preparation for deployment to HSCS
	fileCreateTransaction, err := hedera.NewFileCreateTransaction().
		SetKeys(accountKey).
		SetContents(evmBytecode).
		Execute(client)
	if err != nil {
		log.Fatalf("Error creating file: %v", err)
	}
	fileReceipt, err := fileCreateTransaction.
		SetValidateStatus(true).
		GetReceipt(client)
	if err != nil {
		log.Fatalf("Error getting file create receipt: %v", err)
	}
	bytecodeFileID := *fileReceipt.FileID

	// Deploy smart contract
	contractCreateTransaction, err := hedera.NewContractCreateTransaction().
		SetBytecodeFileID(bytecodeFileID).
		SetGas(100000).
		Execute(client)
	if err != nil {
		log.Fatalf("Error creating contract: %v", err)
	}
	contractReceipt, err := contractCreateTransaction.
		SetValidateStatus(true).
		GetReceipt(client)
	if err != nil {
		log.Fatalf("Error getting contract create receipt: %v", err)
	}
	myContractId := *contractReceipt.ContractID
	myContractExplorerUrl := fmt.Sprintf(
		"https://hashscan.io/testnet/address/%s",
		myContractId,
	)

	// Write data to smart contract
	// NOTE: Invoke a smart contract transaction
	// Step (3) in the accompanying tutorial
	// SetFunction("introduce", hedera.NewContractFunctionParameters().AddString(/* ... */)).
	contractExecuteTransaction, err := hedera.NewContractExecuteTransaction().
		SetContractID(myContractId).
		SetGas(100000).
		SetFunction("introduce", hedera.NewContractFunctionParameters().AddString("bguiz")).
		Execute(client)
	if err != nil {
		log.Fatalf("Error executing contract: %v", err)
	}
	executeReceipt, err := contractExecuteTransaction.
		SetValidateStatus(true).
		GetReceipt(client)
	if err != nil {
		log.Fatalf("Error getting contract execute receipt: %v", err)
	}
	myContractWriteTxId := executeReceipt.TransactionID
	myContractWriteTxExplorerUrl := fmt.Sprintf(
		"https://hashscan.io/testnet/transaction/%s",
		myContractWriteTxId,
	)

	// Read data from smart contract
	// NOTE: Invoke a smart contract query
	// Step (4) in the accompanying tutorial
	//	SetContractID(/* ... */).
	//	SetGas(100000).
	//	SetFunction(/* ... */).
	contractCallQuery, err := hedera.NewContractCallQuery().
		SetContractID(myContractId).
		SetGas(100000).
		SetFunction("greet", nil).
		Execute(client)
	if err != nil {
		log.Fatalf("Error calling contract: %v", err)
	}
	myContractQueryResult := contractCallQuery.GetString(0)

	client.Close()

	// output results
	fmt.Printf("accountId: %v\n", accountId)
	fmt.Printf("accountExplorerUrl: %v\n", accountExplorerUrl)
	fmt.Printf("myContractId: %v\n", myContractId)
	fmt.Printf("myContractExplorerUrl: %v\n", myContractExplorerUrl)
	fmt.Printf("myContractWriteTxId: %v\n", myContractWriteTxId)
	fmt.Printf("myContractWriteTxExplorerUrl: %v\n", myContractWriteTxExplorerUrl)
	fmt.Printf("myContractQueryResult: %v\n", myContractQueryResult)
}
