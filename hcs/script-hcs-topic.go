package main

import (
	"encoding/json"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/imroc/req/v3"
	"github.com/joho/godotenv"
)

type TopicMessagesMNAPIResponse struct {
	Messages []struct {
		SequenceNumber int64 `json:"sequence_number"`
		Message string `json:"message"`
	} `json:"messages"`
}

func main() {
	fmt.Println("🏁 Hello Future World - HCS Topic - start")

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

	// Create a Hedera Consensus Service (HCS) topic
	fmt.Println("🟣 Creating new HCS topic")
	topicCreateTx, _ := hedera.NewTopicCreateTransaction().
	  SetTopicMemo("Hello Future World topic - xyz").
	  // Freeze the transaction to prepare for signing
	  FreezeWith(client);

	// Get the transaction ID of the transaction.
	// The SDK automatically generates and assigns a transaction ID when the transaction is created
	topicCreateTxId := topicCreateTx.GetTransactionID();
	fmt.Printf("The topic create transaction ID: %s\n", topicCreateTxId.String())

	// Sign the transaction with the private key of the treasury account (operator key)
	topicCreateTxSigned := topicCreateTx.Sign(operatorKey)

	// Submit the transaction to the Hedera Testnet
	topicCreateTxSubmitted, _ := topicCreateTxSigned.Execute(client)

	// Get the transaction receipt
	topicCreateTxReceipt, _ := topicCreateTxSubmitted.GetReceipt(client)

	// Get the token ID
	topicId := topicCreateTxReceipt.TopicID
	fmt.Printf("Topic ID: %s\n", topicId.String())

	// Publish a message to the Hedera Consensus Service (HCS) topic
	fmt.Println("🟣 Publish message to HCS topic")
	topicMsgSubmitTx, _ := hedera.NewTopicMessageSubmitTransaction().
	  //Set the transaction memo with the hello future world ID
	  SetTransactionMemo("Hello Future World topic message - xyz").
	  SetTopicID(*topicId).
	  // Set the topic message contents
	  SetMessage([]byte("Hello HCS!")).
	  // Freeze the transaction to prepare for signing
	  FreezeWith(client);

	// Get the transaction ID of the transaction.
	// The SDK automatically generates and assigns a transaction ID when the transaction is created
	topicMsgSubmitTxId := topicMsgSubmitTx.GetTransactionID();
	fmt.Printf("The topic message submit transaction ID: %s\n", topicMsgSubmitTxId.String())

	// Sign the transaction with the private key of the treasury account (operator key)
	topicMsgSubmitTxSigned := topicMsgSubmitTx.Sign(operatorKey)

	// Submit the transaction to the Hedera Testnet
	topicMsgSubmitTxSubmitted, _ := topicMsgSubmitTxSigned.Execute(client)

	// Get the transaction receipt
	topicMsgSubmitTxReceipt, _ := topicMsgSubmitTxSubmitted.GetReceipt(client)

	// Get the topic message sequence number
	topicMsgSeqNum := topicMsgSubmitTxReceipt.TopicSequenceNumber
	fmt.Printf("Topic Message Sequence Number: %v\n", topicMsgSeqNum)

	client.Close()

	// Verify transaction using Hashscan
	// This is a manual step, the code below only outputs the URL to visit

	// View your topic on HashScan
	fmt.Println("🟣 View the topic on HashScan")
	topicHashscanUrl :=
		fmt.Sprintf("https://hashscan.io/testnet/topic/%s", topicId.String())
	fmt.Printf("Topic Hashscan URL: %s\n", topicHashscanUrl)

	// Wait for 6s for record files (blocks) to propagate to mirror nodes
	time.Sleep(6 * time.Second)

	// Verify topic using Mirror Node API
	fmt.Println("🟣 Get topic data from the Hedera Mirror Node")
	topicMirrorNodeApiUrl :=
		fmt.Sprintf("https://testnet.mirrornode.hedera.com/api/v1/topics/%s/messages?encoding=base64&limit=5&order=asc&sequencenumber=1", topicId.String())
	fmt.Printf("The token Hedera Mirror Node API URL: %s\n", topicMirrorNodeApiUrl)

	httpResp, err := req.R().Get(topicMirrorNodeApiUrl)
	if err != nil {
		log.Fatalf("Failed to fetch token URL: %v", err)
	}
	var topicResp TopicMessagesMNAPIResponse
	err = json.Unmarshal(httpResp.Bytes(), &topicResp)
	if err != nil {
		log.Fatalf("Failed to parse JSON of response fetched from topic message URL: %v", err)
	}

	topicMessages := topicResp.Messages
	fmt.Println("Messages retrieved from this topic:")
	for _, entry := range topicMessages {
		seqNum := entry.SequenceNumber
		decodedMsg, _ := base64.StdEncoding.DecodeString(entry.Message)
		fmt.Printf("#%v: %s\n", seqNum, decodedMsg)
	}

	fmt.Println("🎉 Hello Future World - HCS Topic - complete")
}
