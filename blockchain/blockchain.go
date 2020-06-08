package blockchain

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/dgraph-io/badger"
)

const (
	dbPath      = "./tmp/blocks_%s"
	genesisData = "First Transaction from Genesis"
)

type BlockChain struct {
	LastHash []byte
	Database *badger.DB
}

func DBexists(path string) bool {
	if _, err := os.Stat(path + "/MANIFEST"); os.IsNotExist(err) {
		return false
	}

	return true
}

func ContinueBlockChain(nodeId string) *BlockChain {
	path := fmt.Sprintf(dbPath, nodeId)
	if DBexists(path) == false {
		fmt.Println("No existing blockchain found, create one!")
		runtime.Goexit()
	}

	var lastHash []byte

	opts := badger.DefaultOptions
	opts.Dir = path
	opts.ValueDir = path

	db, err := openDB(path, opts)
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

func InitBlockChain(address, nodeId string) *BlockChain {
	path := fmt.Sprintf(dbPath, nodeId)
	if DBexists(path) {
		fmt.Println("Blockchain already exists")
		runtime.Goexit()
	}
	var lastHash []byte
	opts := badger.DefaultOptions
	opts.Dir = path
	opts.ValueDir = path

	db, err := openDB(path, opts)
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

func (chain *BlockChain) AddBlock(block *Block) {
	err := chain.Database.Update(func(txn *badger.Txn) error {
		// if _, err := txn.Get(block.Hash); err == nil {
		// 	return nil
		// }

		blockData := block.Serialize()
		err := txn.Set(block.Hash, blockData)
		Handle(err)

		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, _ := item.Value()

		item, err = txn.Get(lastHash)
		Handle(err)
		lastBlockData, _ := item.Value()

		lastBlock := Deserialize(lastBlockData)

		if block.Height > lastBlock.Height {
			err = txn.Set([]byte("lh"), block.Hash)
			Handle(err)
			chain.LastHash = block.Hash
		}

		return nil
	})
	Handle(err)
}

func (chain *BlockChain) GetBlock(blockHash []byte) (Block, error) {
	var block Block

	err := chain.Database.View(func(txn *badger.Txn) error {
		if item, err := txn.Get(blockHash); err != nil {
			return errors.New("Block is not found")
		} else {
			blockData, _ := item.Value()

			block = *Deserialize(blockData)
		}
		return nil
	})
	if err != nil {
		return block, err
	}

	return block, nil
}

func (chain *BlockChain) GetBlockHashes() [][]byte {
	var blocks [][]byte

	iter := chain.Iterator()

	for {
		block := iter.Next()

		blocks = append(blocks, block.Hash)

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return blocks
}

func (chain *BlockChain) GetBestHeight() int {
	var lastBlock Block

	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, _ := item.Value()

		item, err = txn.Get(lastHash)
		Handle(err)
		lastBlockData, _ := item.Value()

		lastBlock = *Deserialize(lastBlockData)

		return nil
	})
	Handle(err)

	return lastBlock.Height
}

func (chain *BlockChain) GetBestVersion() int {
	var lastBlock Block

	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, _ := item.Value()

		item, err = txn.Get(lastHash)
		Handle(err)
		lastBlockData, _ := item.Value()

		lastBlock = *Deserialize(lastBlockData)

		return nil
	})
	Handle(err)

	return lastBlock.Version
}

func (chain *BlockChain) MineBlock(data string) *Block {
	var lastHash []byte
	var lastHeight int

	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, err = item.Value()

		item, err = txn.Get(lastHash)
		Handle(err)
		lastBlockData, _ := item.Value()

		lastBlock := Deserialize(lastBlockData)

		lastHeight = lastBlock.Height

		return err
	})
	Handle(err)

	newBlock := CreateBlock(data, lastHash, lastHeight+1)

	err = chain.Database.Update(func(txn *badger.Txn) error {
		err := txn.Set(newBlock.Hash, newBlock.Serialize())
		Handle(err)
		err = txn.Set([]byte("lh"), newBlock.Hash)

		chain.LastHash = newBlock.Hash

		return err
	})
	Handle(err)

	return newBlock
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

func (chain *BlockChain) AddTransactions(transactions []*Transaction, hash string) *Block {
	var block *Block

	iter := chain.Iterator()

	for {
		block = iter.Next()
		bhash := hex.EncodeToString(block.Hash)

		if bhash == hash {
			var lastHash []byte
			var lastBlock *Block

			err := chain.Database.View(func(txn *badger.Txn) error {
				item, err := txn.Get([]byte("lh"))
				Handle(err)
				lastHash, err = item.Value()
				item, err = txn.Get(lastHash)
				Handle(err)
				lastBlockData, _ := item.Value()

				lastBlock = Deserialize(lastBlockData)

				return err
			})
			Handle(err)
			trans := append(block.Transactions, transactions[0])
			version := block.Version + 1

			newBlock := UpdateBlock(block.Timestamp, block.Hash, block.Data, block.PrevHash, trans, block.Nonce, block.Height, version)

			lhash := hex.EncodeToString(lastHash)
			if lhash != bhash {
				newlastBlock := UpdateBlock(lastBlock.Timestamp, lastBlock.Hash, lastBlock.Data, lastBlock.PrevHash, lastBlock.Transactions, lastBlock.Nonce, lastBlock.Height, lastBlock.Version+1)
				err = chain.Database.Update(func(txn *badger.Txn) error {
					err := txn.Delete(lastBlock.Hash)
					err = txn.Delete(block.Hash)
					Handle(err)
					err = txn.Set(newlastBlock.Hash, newlastBlock.Serialize())
					err = txn.Set([]byte("lh"), newlastBlock.Hash)
					err = txn.Set(newBlock.Hash, newBlock.Serialize())

					return err
				})
				Handle(err)
			}

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
			return newBlock
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

<<<<<<< HEAD
func (chain *BlockChain) UpdateBlock(data, hash string) *Block {
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

			newBlock := UpdateBlock(block.Timestamp, block.Hash, []byte(data), block.PrevHash, block.Transactions, block.Nonce, block.Height, block.Version)

			// newBlock := UpdateBlock(block.Hash, "Data changed too", block.PrevHash, block.Nonce)

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

			return newBlock
		}
	}

	// return nil
}

=======
>>>>>>> 5b9bd97b039a2d9261bad8a0738b116c80b9dca1
func retry(dir string, originalOpts badger.Options) (*badger.DB, error) {
	lockPath := filepath.Join(dir, "LOCK")
	if err := os.Remove(lockPath); err != nil {
		return nil, fmt.Errorf(`removing "LOCK": %s`, err)
	}
	retryOpts := originalOpts
	retryOpts.Truncate = true
	db, err := badger.Open(retryOpts)
	return db, err
}

func openDB(dir string, opts badger.Options) (*badger.DB, error) {
	if db, err := badger.Open(opts); err != nil {
		if strings.Contains(err.Error(), "LOCK") {
			if db, err := retry(dir, opts); err == nil {
				log.Println("database unlocked, value log truncated")
				return db, nil
			}
			log.Println("could not unlock database:", err)
		}
		return nil, err
	} else {
		return db, nil
	}
}
<<<<<<< HEAD
=======

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
>>>>>>> 5b9bd97b039a2d9261bad8a0738b116c80b9dca1
