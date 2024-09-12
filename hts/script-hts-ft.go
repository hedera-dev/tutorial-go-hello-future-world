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
	fmt.Println("🏁 Hello Future World - HTS Fungible Token - start")

	// Load environment variables from .env file
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Initialize the operator account
	operatorIdStr := os.Getenv("OPERATOR_ACCOUNT_ID")
	operatorKeyStr := os.Getenv("OPERATOR_ACCOUNT_PRIVATE_KEY")
	if operatorIdStr == "" || operatorKeyStr == "" {
		log.Fatal("Must set OPERATOR_ACCOUNT_ID, OPERATOR_ACCOUNT_PRIVATE_KEY")
	}

	operatorId, _ := hedera.AccountIDFromString(operatorIdStr)
	// Necessary because Go SDK v2.37.0 does not handle the `0x` prefix automatically
	// Ref: https://github.com/hashgraph/hedera-sdk-go/issues/1057
	operatorKeyStr = strings.TrimPrefix(operatorKeyStr, "0x")
	operatorKey, _ := hedera.PrivateKeyFromStringECDSA(operatorKeyStr)
	fmt.Printf("Using account: %s\n", operatorId)
	fmt.Printf("Using operatorKey: %s\n", operatorKeyStr)

	// The client operator ID and key is the account that will be automatically set to pay for the transaction fees for each transaction
	client := hedera.ClientForTestnet()
	client.SetOperator(operatorId, operatorKey)

	// Set the default maximum transaction fee (in HBAR)
	client.SetDefaultMaxTransactionFee(hedera.HbarFrom(100, hedera.HbarUnits.Hbar))
	// Set the default maximum payment for queries (in HBAR)
	client.SetDefaultMaxQueryPayment(hedera.HbarFrom(50, hedera.HbarUnits.Hbar))

	// TODO ... SDK methods

	client.Close()

	// TODO ... Hashscan/ Mirror node APIs

	fmt.Println("🎉 Hello Future World - HTS Fungible Token - complete")
}
