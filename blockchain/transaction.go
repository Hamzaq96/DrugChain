package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"strings"

	"github.com/tensor-programming/golang-blockchain/wallet"
)

type Transaction struct {
	ID        []byte
	BlockHash string
	Signature []byte
	// Inputs []TxInput
	// Outputs []TxOutput
}

// type TxOutput struct {
// 	PubKey string
// }

// type TxInput struct {
// 	ID []byte
// 	Sig string
// }

func (tx *Transaction) Hash() []byte {
	var hash [32]byte

	txCopy := *tx
	txCopy.ID = []byte{}

	hash = sha256.Sum256(txCopy.Serialize())

	return hash[:]
}

func (tx Transaction) Serialize() []byte {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	return encoded.Bytes()
}

func DeserializeTransaction(data []byte) Transaction {
	var transaction Transaction

	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&transaction)
	Handle(err)
	return transaction
}

func (tx *Transaction) SetID() {
	var encoded bytes.Buffer
	var hash [32]byte

	encode := gob.NewEncoder(&encoded)
	err := encode.Encode(tx)
	Handle(err)

	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}

func CoinbaseTx(from string, data string) *Transaction {
	// if data == "" {
	// 	data = fmt.Sprintf("Welcome to the first locally developed blockchain", from)
	// }

	// tx := Transaction{nil, from, from}
	// tx.SetID()

	// return &tx
	return nil
}

func NewTransaction(from string, blockHash string, chain *BlockChain) *Transaction {
	block := chain.FindBlock(blockHash)
	bhash := hex.EncodeToString(block.Hash)

	wallets, err := wallet.CreateWallets()
	Handle(err)
	w := wallets.GetWallet(from)
	// pubKeyHash := wallet.PublicKeyHash(w.PublicKey)

	tx := Transaction{nil, bhash, w.PublicKey}
	// tx.ID = tx.Hash()
	tx.SetID()

	return &tx
}

// func (tx *Transaction) Sign()

func (tx Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Transaction %x:", tx.ID))
	// lines = append(lines, fmt.Sprintf("    Block %x:", tx.BlockHash))
	lines = append(lines, fmt.Sprintf("    Signature %x:", tx.Signature))

	// for i, input := range tx.Inputs {
	// 	lines = append(lines, fmt.Sprintf("     Input %d:", i))
	// 	lines = append(lines, fmt.Sprintf("       TXID:     %x", input.ID))
	// 	lines = append(lines, fmt.Sprintf("       Out:       %d", input.Out))
	// 	lines = append(lines, fmt.Sprintf("       Signature: %x", input.Signature))
	// 	lines = append(lines, fmt.Sprintf("       PubKey:    %x", input.PubKey))
	// }

	// for i, output := range tx.Outputs {
	// 	lines = append(lines, fmt.Sprintf("     Output %d:", i))
	// 	lines = append(lines, fmt.Sprintf("       Value:  %d", output.Value))
	// 	lines = append(lines, fmt.Sprintf("       Script: %x", output.PubKeyHash))
	// }

	return strings.Join(lines, "\n")
}
