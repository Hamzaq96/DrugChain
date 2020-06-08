package blockchain

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"
)

type Block struct {
	Timestamp    int64
	Hash         []byte
	Data         []byte
	PrevHash     []byte
	Transactions []*Transaction
	Nonce        int
	Height       int
	Version      int
}

func CreateBlock(data string, prevHash []byte, height int) *Block {
	block := &Block{time.Now().Unix(), []byte{}, []byte(data), prevHash, []*Transaction{}, 0, height, 0}
	pow := NewProof(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

func UpdateBlock(timestamp int64, hash []byte, data []byte, prevHash []byte, txs []*Transaction, nonce int, height, ver int) *Block {
	block := &Block{timestamp, hash, data, prevHash, txs, nonce, height, ver}

	return block
}

func Genesis(coinbase *Transaction) *Block {
	return CreateBlock("Genesis", []byte{}, 0)
}

func (b *Block) Serialize() []byte {
	var res bytes.Buffer
	encoder := gob.NewEncoder(&res)

	err := encoder.Encode(b)

	Handle(err)

	return res.Bytes()
}

func Deserialize(data []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(data))

	err := decoder.Decode(&block)

	Handle(err)

	return &block
}

func Handle(err error) {
	if err != nil {
		log.Panic(err)
	}
}
