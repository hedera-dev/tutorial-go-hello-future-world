package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/imroc/req/v3"
	"github.com/joho/godotenv"
)

type TokenMNAPIResponse struct {
	Name        string `json:"name"`
	TotalSupply string `json:"total_supply"`
}

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

	fmt.Println("🟣 Creating new HTS token")
	tokenCreateTx, _ := hedera.NewTokenCreateTransaction().
		//Set the transaction memo
		SetTransactionMemo("Hello Future World token - xyz").
		// HTS `TokenType.FungibleCommon` behaves similarly to ERC20
		SetTokenType(hedera.TokenTypeFungibleCommon).
		// Configure token options: name, symbol, decimals, initial supply
		SetTokenName(`htsFt coin`).
		// Set the token symbol
		SetTokenSymbol("HTSFT").
		// Set the token decimals to 2
		SetDecimals(2).
		// Set the initial supply of the token to 1,000,000
		SetInitialSupply(1_000_000).
		// Configure token access permissions: treasury account, admin, freezing
		SetTreasuryAccountID(operatorId).
		// Set the freeze default value to false
		SetFreezeDefault(false).
		//Freeze the transaction and prepare for signing
		FreezeWith(client)

	// Get the transaction ID of the transaction. The SDK automatically generates and assigns a transaction ID when the transaction is created
	tokenCreateTxId := tokenCreateTx.GetTransactionID()
	fmt.Printf("The token create transaction ID: %s\n", tokenCreateTxId.String())

	// Sign the transaction with the private key of the treasury account (operator key)
	tokenCreateTxSigned := tokenCreateTx.Sign(operatorKey)

	// Submit the transaction to the Hedera Testnet
	tokenCreateTxSubmitted, _ := tokenCreateTxSigned.Execute(client)

	// Get the transaction receipt
	tokenCreateTxReceipt, _ := tokenCreateTxSubmitted.GetReceipt(client)

	// Get the token ID
	tokenId := tokenCreateTxReceipt.TokenID
	fmt.Printf("Token ID: %s\n", tokenId.String())

	client.Close()

	// Verify transactions using Hashscan
	// This is a manual step, the code below only outputs the URL to visit

	// View your token on HashScan
	fmt.Println("🟣 View the token on HashScan")
	tokenHashscanUrl :=
		fmt.Sprintf("https://hashscan.io/testnet/token/%s", tokenId.String())
	fmt.Printf("Token Hashscan URL: %s\n", tokenHashscanUrl)

	// Wait for 6s for record files (blocks) to propagate to mirror nodes
	time.Sleep(6 * time.Second)

	// Verify token using Mirror Node API
	fmt.Println("🟣 Get token data from the Hedera Mirror Node")
	tokenMirrorNodeApiUrl :=
		fmt.Sprintf("https://testnet.mirrornode.hedera.com/api/v1/tokens/%s", tokenId.String())
	fmt.Printf("The token Hedera Mirror Node API URL: %s\n", tokenMirrorNodeApiUrl)

	httpResp, err := req.R().Get(tokenMirrorNodeApiUrl)
	if err != nil {
		log.Fatalf("Failed to fetch token URL: %v", err)
	}
	var tokenResp TokenMNAPIResponse
	err = json.Unmarshal(httpResp.Bytes(), &tokenResp)
	if err != nil {
		log.Fatalf("Failed to parse JSON of response fetched from token URL: %v", err)
	}
	tokenName := tokenResp.Name
	fmt.Printf("The name of this token: %s\n", tokenName)
	tokenTotalSupply := tokenResp.TotalSupply
	fmt.Printf("The total supply of this token: %s\n", tokenTotalSupply)

	fmt.Println("🎉 Hello Future World - HTS Fungible Token - complete")
}
