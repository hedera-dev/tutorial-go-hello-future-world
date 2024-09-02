package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/hashgraph/hedera-sdk-go/v2"
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

	// NOTE: Create a topic
	// Step (1) in the accompanying tutorial
	//  const topicCreateTx = await
	//     /* ... */;
	//    .freezeWith(client);
	topicCreateTx, err := hedera.NewTopicCreateTransaction().
		FreezeWith(client)
	if err != nil {
		log.Fatalf("Error in TopicCreateTransaction: %v\n", err)
	}

	topicCreateTxSigned := topicCreateTx.Sign(accountKey)
	topicCreateTxResponse, err := topicCreateTxSigned.Execute(client)
	if err != nil {
		log.Fatalf("Error executing TopicCreateTransaction: %v\n", err)
	}

	topicCreateTxReceipt, err := topicCreateTxResponse.
		SetValidateStatus(true).
		GetReceipt(client)
	if err != nil {
		log.Fatalf("Error getting receipt for TopicCreateTransaction: %v\n", err)
	}

	topicCreateTxId := topicCreateTxReceipt.TransactionID
	topicId := topicCreateTxReceipt.TopicID
	topicExplorerUrl := fmt.Sprintf("https://hashscan.io/testnet/topic/%s", topicId)

	// NOTE: Publish message to topic
	// Step (2) in the accompanying tutorial
	//  const topicCreateTx = await
	//     /* ... */;
	//    .freezeWith(client);
	topicMsgSubmitTx, err := hedera.NewTopicMessageSubmitTransaction().
		SetTopicID(*topicId).
		SetMessage([]byte("Hello HCS - bguiz")).
		FreezeWith(client)
	if err != nil {
		log.Fatalf("Error in TopicMessageSubmitTransaction: %v\n", err)
	}

	topicMsgSubmitTxSigned := topicMsgSubmitTx.Sign(accountKey)
	topicMsgSubmitTxResponse, err := topicMsgSubmitTxSigned.Execute(client)
	if err != nil {
		log.Fatalf("Error executing TopicMessageSubmitTransaction: %v\n", err)
	}

	topicMsgSubmitTxId := topicMsgSubmitTxResponse.TransactionID
	topicMsgSubmitTxReceipt, err := topicMsgSubmitTxResponse.
		SetValidateStatus(true).
		GetReceipt(client)
	if err != nil {
		log.Fatalf("Error getting receipt for TopicMessageSubmitTransaction: %v\n", err)
	}

	topicMessageMirrorUrl := fmt.Sprintf(
		"https://testnet.mirrornode.hedera.com/api/v1/topics/%s/messages/%d",
		topicId,
		topicMsgSubmitTxReceipt.TopicSequenceNumber,
	)

	client.Close()

	// output results
	fmt.Printf("accountId: %v\n", accountId)
	fmt.Printf("topicId: %v\n", topicId)
	fmt.Printf("topicExplorerUrl: %v\n", topicExplorerUrl)
	fmt.Printf("topicCreateTxId: %v\n", topicCreateTxId)
	fmt.Printf("topicMsgSubmitTxId: %v\n", topicMsgSubmitTxId)
	fmt.Printf("topicMessageMirrorUrl: %v\n", topicMessageMirrorUrl)
}
