package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/imroc/req/v3"
	"github.com/joho/godotenv"
	"github.com/shopspring/decimal"
	bip32 "github.com/tyler-smith/go-bip32"
	bip39 "github.com/tyler-smith/go-bip39"
)

type AccountBalanceResponse struct {
	Balances []struct {
		Account string `json:"account"`
		Balance int64  `json:"balance"`
	} `json:"balances"`
}

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	seedPhrase := os.Getenv("SEED_PHRASE")
	if seedPhrase == "" {
		log.Fatal("Please set required keys in .env file.")
	}

	// Create an ECSDA secp256k1 private key based on a BIP-39 seed phrase,
	// plus the default BIP-32/BIP-44 HD Wallet derivation path used by Metamask.

	// BIP-39 Generate seed from seed phrase
	seed := bip39.NewSeed(seedPhrase, "")

	// BIP-32 Derive master key
	masterKey, err := bip32.NewMasterKey(seed)
	if err != nil {
		log.Fatalf("Failed to create master key: %v", err)
	}

	// BIP-44 Derive private key of single account from BIP-32 master key + BIP-44 HD path
	// Using m/44'/60'/0'/0/0 for compatibility with Metamask hardcoded default
	purpose, _ := masterKey.NewChildKey(bip32.FirstHardenedChild + 44)
	coinType, _ := purpose.NewChildKey(bip32.FirstHardenedChild + 60)
	account, _ := coinType.NewChildKey(bip32.FirstHardenedChild + 0)
	change, _ := account.NewChildKey(0)
	childKey, _ := change.NewChildKey(0)

	privateKeyEcdsa, err := crypto.ToECDSA(childKey.Key)
	if err != nil {
		log.Fatalf("Failed to convert to ECDSA: %v", err)
	}
	privateKeyBytes := crypto.FromECDSA(privateKeyEcdsa)

	privateKey, err := hedera.PrivateKeyFromBytesECDSA(privateKeyBytes)
	if err != nil {
		log.Fatalf("Failed to create private key from bytes: ", err)
	}
	privateKeyHex := fmt.Sprintf("0x%s", privateKey.StringRaw())
	evmAddress := fmt.Sprintf("0x%s", privateKey.PublicKey().ToEvmAddress())
	accountExplorerUrl := fmt.Sprintf("https://hashscan.io/testnet/account/%s", evmAddress)
	accountBalanceFetchApiUrl := fmt.Sprintf(
		"https://testnet.mirrornode.hedera.com/api/v1/balances?account.id=%s&limit=1&order=asc", evmAddress)

	var accountId string
	var accountBalanceTinybar int64
	var accountBalanceHbar string
	httpResp, err := req.R().Get(accountBalanceFetchApiUrl)
	if err != nil {
		fmt.Printf("Failed to fetch account balance URL: %v", err)
	} else {
		var accountBalanceResp AccountBalanceResponse
		err = json.Unmarshal(httpResp.Bytes(), &accountBalanceResp)
		if err != nil {
			log.Printf("Failed to parse JSON of response fetched from account balance URL: %v", err)
		} else {
			if len(accountBalanceResp.Balances) > 0 {
				item := accountBalanceResp.Balances[0]
				accountId = item.Account
				accountBalanceTinybar = item.Balance
				accountBalanceHbar = decimal.
					NewFromInt(accountBalanceTinybar).
					Div(decimal.NewFromInt(10_000_000)).
					StringFixed(8)
			}
		}
	}

	fmt.Printf("privateKeyHex: %s\n", privateKeyHex)
	fmt.Printf("evmAddress: %s\n", evmAddress)
	fmt.Printf("accountExplorerUrl: %s\n", accountExplorerUrl)
	fmt.Printf("accountBalanceFetchApiUrl: %s\n", accountBalanceFetchApiUrl)
	fmt.Printf("accountId: %s\n", accountId)
	fmt.Printf("accountBalanceHbar: %s\n", accountBalanceHbar)
}
