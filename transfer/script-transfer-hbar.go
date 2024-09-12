package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/imroc/req/v3"
	"github.com/joho/godotenv"
)

type TransferMNAPIResponse struct {
	Account string `json:"account"`
	Amount  int64  `json:"amount"`
}
type TransferTransactionMNAPIResponse struct {
	Transactions []struct {
		Transfers []TransferMNAPIResponse `json:"transfers"`
	} `json:"transactions"`
}

func main() {
	fmt.Println("🏁 Hello Future World - Transfer Hbar - start")

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

	fmt.Println("🟣 Creating, signing, and submitting the transfer transaction")

	recipientAccount1, _ := hedera.AccountIDFromString("0.0.200")
	recipientAccount2, _ := hedera.AccountIDFromString("0.0.201")
	transferTx, _ := hedera.NewTransferTransaction().
		SetTransactionMemo(fmt.Sprintf("Hello Future World transfer - xyz")).
		// Debit  3 HBAR from the operator account (sender)
		AddHbarTransfer(operatorId, hedera.HbarFrom(-3, hedera.HbarUnits.Hbar)).
		// Credit 1 HBAR to account 0.0.200 (1st recipient)
		AddHbarTransfer(recipientAccount1, hedera.HbarFrom(1, hedera.HbarUnits.Hbar)).
		// Credit 2 HBAR to account 0.0.201 (2nd recipient)
		AddHbarTransfer(recipientAccount2, hedera.HbarFrom(2, hedera.HbarUnits.Hbar)).
		// Freeze the transaction to prepare for signing
		FreezeWith(client)

	// Get the transaction ID for the transfer transaction
	transferTxId := transferTx.GetTransactionID()
	fmt.Printf("The transfer transaction ID: %s\n", transferTxId.String())

	// Sign the transaction with the account that is being debited (operator account) and the transaction fee payer account (operator account)
	// Since the account that is being debited and the account that is paying for the transaction are the same, only one account's signature is required
	transferTxSigned := transferTx.Sign(operatorKey)

	// Submit the transaction to the Hedera Testnet
	transferTxSubmitted, err := transferTxSigned.Execute(client)
	if err != nil {
		log.Fatalf("Error executing TransferTransaction: %v\n", err)
	}

	// Get the transfer transaction receipt
	transferTxReceipt, err := transferTxSubmitted.GetReceipt(client)
	if err != nil {
		log.Fatalf("Error getting receipt for TransferTransaction: %v\n", err)
	}
	transactionStatus := transferTxReceipt.Status
	fmt.Printf("The transfer transaction status is: %s\n", transactionStatus.String())

	// NOTE: Query HBAR balance using AccountBalanceQuery
	newAccountBalance, _ := hedera.NewAccountBalanceQuery().
		SetAccountID(operatorId).
		Execute(client)
	newHbarBalance := newAccountBalance.Hbars
	fmt.Printf("The new account balance after the transfer: %s\n", newHbarBalance.String())

	client.Close()

	// View the transaction in HashScan
	fmt.Println("🟣 View the transfer transaction transaction in HashScan")
	transferTxVerifyHashscanUrl :=
		fmt.Sprintf("https://hashscan.io/testnet/transaction/%s", transferTxId.String())
	fmt.Printf("Copy and paste this URL in your browser: %s\n", transferTxVerifyHashscanUrl)

	fmt.Println("🟣 Get transfer transaction data from the Hedera Mirror Node")

	// Wait for 6s for record files (blocks) to propagate to mirror nodes
	time.Sleep(6 * time.Second)

	// The transfer transaction mirror node API request
	transferTxIdMirrorNodeFormat, _ :=
		ConvertTransactionIDForMirrorNodeAPI(transferTxId)
	transferTxVerifyMirrorNodeApiUrl :=
		fmt.Sprintf("https://testnet.mirrornode.hedera.com/api/v1/transactions/%s?nonce=0", transferTxIdMirrorNodeFormat)
	fmt.Printf("The transfer transaction Hedera Mirror Node API URL: %s\n", transferTxVerifyMirrorNodeApiUrl)

	httpResp, err := req.R().Get(transferTxVerifyMirrorNodeApiUrl)
	if err != nil {
		log.Fatalf("Failed to fetch account balance URL: %v", err)
	}
	var transferTxResp TransferTransactionMNAPIResponse
	err = json.Unmarshal(httpResp.Bytes(), &transferTxResp)
	if err != nil {
		log.Fatalf("Failed to parse JSON of response fetched from account balance URL: %v", err)
	}
	transferJsonAccountTransfers := transferTxResp.Transactions[0].Transfers

	// Discard all entries whose values are less than 0.1 hbars
	// as these are network fees
	var filteredTransfers []TransferMNAPIResponse
	for _, entry := range transferJsonAccountTransfers {
		if math.Abs(float64(entry.Amount)) >= 100_000_00 {
			filteredTransfers = append(filteredTransfers, entry)
		}
	}
	sort.Slice(filteredTransfers, func(i, j int) bool {
		return filteredTransfers[i].Amount < filteredTransfers[j].Amount
	})
	var transferJsonAccountTransfersFinalAmounts []map[string]string
	for _, entry := range filteredTransfers {
		transferJsonAccountTransfersFinalAmounts = append(transferJsonAccountTransfersFinalAmounts, map[string]string{
			"AccountID": entry.Account,
			"Amount":    hedera.HbarFromTinybar(entry.Amount).ToString(hedera.HbarUnits.Hbar),
		})
	}

	fmt.Printf("%-15s %-15s\n", "AccountID", "Amount")
	fmt.Println(strings.Repeat("-", 30))
	for _, entry := range transferJsonAccountTransfersFinalAmounts {
		fmt.Printf("%-15s %-15s\n", entry["AccountID"], entry["Amount"])
	}

	fmt.Println("🎉 Hello Future World - Transfer Hbar - complete")
}

func ConvertTransactionIDForMirrorNodeAPI(txID hedera.TransactionID) (string, error) {
	// The transaction ID has to be converted to the correct format to pass in the mirror node query (0.0.x@x.x to 0.0.x-x-x)
	parts := strings.SplitN(txID.String(), "@", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid transaction ID format: %s", txID)
	}

	convertedPart := strings.ReplaceAll(parts[1], ".", "-")

	return fmt.Sprintf("%s-%s", parts[0], convertedPart), nil
}
