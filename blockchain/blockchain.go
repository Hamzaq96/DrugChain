package blockchain

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/dgraph-io/badger"
)

const (
	dbPath      = "./tmp/blocks"
	dbFile      = "./tmp/blocks/MANIFEST" //Used to verify if the blockchain exist or not.
	genesisData = "First transaction from Genesis."
)

type BlockChain struct {
	LastHash []byte
	Database *badger.DB
}

func DBexists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}

func ContinueBlockChain(address string) *BlockChain {
	if DBexists() == false {
		fmt.Println("No existing blockchain found, create one!")
		runtime.Goexit()
	}

	var lastHash []byte

	opts := badger.DefaultOptions
	opts.Dir = dbPath
	opts.ValueDir = dbPath

	db, err := badger.Open(opts)
	Handle(err)

	err = db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, err = item.Value()

		return err
	})
	Handle(err)

	chain := BlockChain{lastHash, db}

	return &chain
}

func InitBlockChain(address string) *BlockChain {
	var lastHash []byte

	if DBexists() {
		fmt.Println("Blockchain already exists")
		runtime.Goexit()
	}

	opts := badger.DefaultOptions
	opts.Dir = dbPath
	opts.ValueDir = dbPath

	db, err := badger.Open(opts)
	Handle(err)

	err = db.Update(func(txn *badger.Txn) error {
		cbtx := CoinbaseTx(address, genesisData)
		genesis := Genesis(cbtx)
		fmt.Println("Genesis created")
		err = txn.Set(genesis.Hash, genesis.Serialize())
		Handle(err)
		err = txn.Set([]byte("lh"), genesis.Hash)

		lastHash = genesis.Hash

		return err

	})

	Handle(err)

	blockchain := BlockChain{lastHash, db}
	return &blockchain
}

func (chain *BlockChain) AddBlock(data string) {
	var lastHash []byte

	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, err = item.Value()

		return err
	})
	Handle(err)

	newBlock := CreateBlock(data, lastHash)

	err = chain.Database.Update(func(txn *badger.Txn) error {
		err := txn.Set(newBlock.Hash, newBlock.Serialize())
		Handle(err)
		err = txn.Set([]byte("lh"), newBlock.Hash)

		chain.LastHash = newBlock.Hash

		return err
	})
	Handle(err)
}

func (chain *BlockChain) FindBlock(hash string) *Block {
	var block *Block

	iter := chain.Iterator()

	for {
		block = iter.Next()
		bhash := hex.EncodeToString(block.Hash)

		if bhash == hash {
			return block
		}

	}

	return nil
}

func (chain *BlockChain) AddTransactions(transactions []*Transaction, hash string) {
	var block *Block

	iter := chain.Iterator()

	for {
		block = iter.Next()
		bhash := hex.EncodeToString(block.Hash)

		if bhash == hash {
			var lastHash []byte

			err := chain.Database.View(func(txn *badger.Txn) error {
				item, err := txn.Get([]byte("lh"))
				Handle(err)
				lastHash, err = item.Value()

				return err
			})
			Handle(err)
			trans := append(block.Transactions, transactions[0])

			newBlock := UpdateBlock(block.Hash, block.Data, block.PrevHash, trans, block.Nonce)

			lhash := hex.EncodeToString(lastHash)
			if lhash == bhash {
				err = chain.Database.Update(func(txn *badger.Txn) error {
					err := txn.Delete(block.Hash)
					Handle(err)
					err = txn.Set(newBlock.Hash, newBlock.Serialize())
					err = txn.Set([]byte("lh"), newBlock.Hash)
					return err
				})
				Handle(err)
			}

			err = chain.Database.Update(func(txn *badger.Txn) error {
				err := txn.Delete(block.Hash)
				Handle(err)
				err = txn.Set(newBlock.Hash, newBlock.Serialize())
				return err
			})
			Handle(err)

		}
	}

	// return nil
}

func (bc *BlockChain) FindTransaction(ID []byte) (Transaction, error) {
	iter := bc.Iterator()

	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return Transaction{}, errors.New("Transaction does not exist")
}

func (chain *BlockChain) UpdateBlock(hash string) {
	// var block *Block

	// iter := chain.Iterator()

	// for {
	// 	block = iter.Next()
	// 	bhash := hex.EncodeToString(block.Hash)

	// 	if bhash == hash {
	// 		var lastHash []byte

	// 		err := chain.Database.View(func(txn *badger.Txn) error {
	// 			item, err := txn.Get([]byte("lh"))
	// 			Handle(err)
	// 			lastHash, err = item.Value()

	// 			return err
	// 		})
	// 		Handle(err)

	// 		// newBlock := UpdateBlock(block.Hash, "Data changed too", block.PrevHash, block.Nonce)

	// 		lhash := hex.EncodeToString(lastHash)
	// 		if lhash == bhash {
	// 			err = chain.Database.Update(func(txn *badger.Txn) error {
	// 				err := txn.Delete(block.Hash)
	// 				Handle(err)
	// 				err = txn.Set(newBlock.Hash, newBlock.Serialize())
	// 				err = txn.Set([]byte("lh"), newBlock.Hash)
	// 				return err
	// 			})
	// 			Handle(err)
	// 		}

	// 		err = chain.Database.Update(func(txn *badger.Txn) error {
	// 			err := txn.Delete(block.Hash)
	// 			Handle(err)
	// 			err = txn.Set(newBlock.Hash, newBlock.Serialize())
	// 			return err
	// 		})
	// 		Handle(err)

	// 	}
	// }

	// return nil
}
