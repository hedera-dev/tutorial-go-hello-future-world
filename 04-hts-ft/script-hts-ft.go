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

type AccountTokenBalanceResponse struct {
	Tokens []struct {
		Balance int64 `json:"balance"`
	} `json:"tokens"`
}

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

	// Create the token
	tokenCreateTx, err := hedera.NewTokenCreateTransaction().
		SetTokenType(hedera.TokenTypeFungibleCommon).
		SetTokenName("bguiz coin").
		SetTokenSymbol("BGZ").
		SetDecimals(2).
		SetInitialSupply(1_000_000).
		SetTreasuryAccountID(accountId).
		SetAdminKey(accountKey).
		SetFreezeDefault(false).
		FreezeWith(client)
	if err != nil {
		log.Fatalf("Error in TokenCreateTransaction: %v\n", err)
	}

	tokenCreateTxSigned := tokenCreateTx.Sign(accountKey)
	tokenCreateTxSubmitted, err := tokenCreateTxSigned.Execute(client)
	if err != nil {
		log.Fatalf("Error executing TokenCreateTransaction: %v\n", err)
	}

	tokenCreateTxReceipt, err := tokenCreateTxSubmitted.
		SetValidateStatus(true).
		GetReceipt(client)
	if err != nil {
		log.Fatalf("Error getting receipt for TokenCreateTransaction: %v\n", err)
	}

	tokenId := tokenCreateTxReceipt.TokenID
	tokenExplorerUrl := fmt.Sprintf("https://hashscan.io/testnet/token/%s", tokenId)

	client.Close()

	// Query token balance of account (mirror node)
	// need to wait 3 seconds for the record files to be ingested by the mirror nodes
	time.Sleep(3 * time.Second)

	// NOTE: Mirror Node API to query specified token balance
	// Step (2) in the accompanying tutorial
	// const accountBalanceFetchApiUrl =
	//     /* ... */;
	accountBalanceFetchApiUrl := fmt.Sprintf(
		"https://testnet.mirrornode.hedera.com/api/v1/accounts/%s/tokens?token.id=%s&limit=1&order=desc",
		accountId,
		tokenId,
	)
	httpResp, err := req.R().Get(accountBalanceFetchApiUrl)
	if err != nil {
		log.Fatalf("Failed to fetch account balance URL: %v", err)
	}
	var accountTokenBalanceResp AccountTokenBalanceResponse
	err = json.Unmarshal(httpResp.Bytes(), &accountTokenBalanceResp)
	if err != nil {
		log.Fatalf("Failed to parse JSON of response fetched from account balance URL: %v", err)
	}
	accountBalanceToken := accountTokenBalanceResp.Tokens[0].Balance

	// output results
	fmt.Printf("accountId: %v\n", accountId)
	fmt.Printf("tokenId: %v\n", tokenId)
	fmt.Printf("tokenExplorerUrl: %v\n", tokenExplorerUrl)
	fmt.Printf("accountTokenBalance: %v\n", accountBalanceToken)
	fmt.Printf("accountBalanceFetchApiUrl: %v\n", accountBalanceFetchApiUrl)
}
