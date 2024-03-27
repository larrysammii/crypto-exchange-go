package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func transferCrypto(client *ethclient.Client, fromPrivateKey *ecdsa.PrivateKey, to common.Address, amount *big.Int) error {
	context := context.Background()
	publicKey := fromPrivateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	nonce, err := client.PendingNonceAt(context, fromAddress)
	if err != nil {
		return err

	}

	gasLimit := uint64(21000)
	gasPrice, err := client.SuggestGasPrice(context)
	if err != nil {
		log.Fatal(err)
	}

	transaction := types.NewTransaction(nonce, to, amount, gasLimit, gasPrice, nil)

	chainID := big.NewInt(1337)
	signedTx, err := types.SignTx(transaction, types.NewEIP155Signer(chainID), fromPrivateKey)
	if err != nil {
		return err
	}

	return client.SendTransaction(context, signedTx)

}
