package network

import (
	"bytes"
	"encoding/gob"
<<<<<<< HEAD
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/tensor-programming/golang-blockchain/blockchain"
	"github.com/tensor-programming/golang-blockchain/wallet"
	"gopkg.in/vrecan/death.v3"
)

type Message1 struct {
	Data string
}

type Message2 struct {
	From string
	To   string
}

type Message3 struct {
	Hash string
	Data string
}

type Message4 struct {
	Hash string
}

func makeMuxRouter(nodeId string) http.Handler {
	muxRouter := mux.NewRouter()

	muxRouter.HandleFunc("/wallet", handleGetWallets).Methods("GET")
	muxRouter.HandleFunc("/addblock", handleAddBlock).Methods("POST")
	muxRouter.HandleFunc("/addwallet", handleAddWallet).Methods("POST")
	muxRouter.HandleFunc("/transaction", handleAddTransaction).Methods("POST")
	muxRouter.HandleFunc("/updateblock", handleUpdateBlock).Methods("POST")
	muxRouter.HandleFunc("/getblock", handleGetBlock).Methods("POST")
	return muxRouter
}

func StartServer(nodeID, minerAddress string) error {
	mux := makeMuxRouter(nodeID)
	httpAddr := nodeID
	log.Println("Listening on ", nodeID)
	s := &http.Server{
		Addr:           ":" + httpAddr,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if err := s.ListenAndServe(); err != nil {
		return err
	}

	return nil
}

func handleGetBlockchain(w http.ResponseWriter, r *http.Request) {

	bytes, err := json.MarshalIndent(blockchain.ContinueBlockChain("3000"), "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	io.WriteString(w, string(bytes))
}

func handleGetWallets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET,PUT,POST,DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Set-Cookie", "HttpOnly;Secure;SameSite=Strict")

	nodeID := os.Getenv("NODE_ID")

	wallets, _ := wallet.CreateWallets(nodeID)

	addresses := wallets.GetAllAddresses()

	for _, address := range addresses {
		// fmt.Println(address)
		io.WriteString(w, address)
	}
}

func handleAddWallet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET,PUT,POST,DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Set-Cookie", "HttpOnly;Secure;SameSite=Strict")

	nodeID := os.Getenv("NODE_ID")

	wallets, _ := wallet.CreateWallets(nodeID)

	address := wallets.AddWallet()
	wallets.SaveFile(nodeID)

	respondWithJSON(w, r, http.StatusCreated, address)
}

func handleAddBlock(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET,PUT,POST,DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Set-Cookie", "HttpOnly;Secure;SameSite=Strict")

	var m Message1
	nodeID := os.Getenv("NODE_ID")
	chain := blockchain.ContinueBlockChain(nodeID)

	fmt.Println("The data is: ", r.Body)
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m); err != nil {
		respondWithJSON(w, r, http.StatusBadRequest, r.Body)
		return
	}
	defer r.Body.Close()
	defer chain.Database.Close()

	// fmt.Println("The data is: ", m.Data)

	// newBlock, err := generateBlock(Blockchain[len(Blockchain)-1], m.BPM)
	newBlock := chain.MineBlock(m.Data)
	// if err != nil {
	// 	respondWithJSON(w, r, http.StatusInternalServerError, m)
	// 	return
	// }
	// if isBlockValid(newBlock, Blockchain[len(Blockchain)-1]) {
	// 	newBlockchain := append(Blockchain, newBlock)
	// 	replaceChain(newBlockchain)
	// 	spew.Dump(Blockchain)
	// }

	respondWithJSON(w, r, http.StatusCreated, newBlock)

}

func handleAddTransaction(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET,PUT,POST,DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Set-Cookie", "HttpOnly;Secure;SameSite=Strict")

	var m Message2
	nodeID := os.Getenv("NODE_ID")
	chain := blockchain.ContinueBlockChain(nodeID)

	fmt.Println("The data is: ", r.Body)
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m); err != nil {
		respondWithJSON(w, r, http.StatusBadRequest, r.Body)
		return
	}
	defer r.Body.Close()
	defer chain.Database.Close()

	// fmt.Println("The data is: ", m.Data)

	// newBlock, err := generateBlock(Blockchain[len(Blockchain)-1], m.BPM)
	tx := blockchain.NewTransaction(m.From, "3000", m.To, chain)
	chain.AddTransactions([]*blockchain.Transaction{tx}, m.To)
	// if err != nil {
	// 	respondWithJSON(w, r, http.StatusInternalServerError, m)
	// 	return
	// }
	// if isBlockValid(newBlock, Blockchain[len(Blockchain)-1]) {
	// 	newBlockchain := append(Blockchain, newBlock)
	// 	replaceChain(newBlockchain)
	// 	spew.Dump(Blockchain)
	// }

	respondWithJSON(w, r, http.StatusCreated, tx)

}

func handleUpdateBlock(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET,PUT,POST,DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Set-Cookie", "HttpOnly;Secure;SameSite=Strict")

	var m Message3
	nodeID := os.Getenv("NODE_ID")
	chain := blockchain.ContinueBlockChain(nodeID)

	fmt.Println("The data is: ", r.Body)
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m); err != nil {
		respondWithJSON(w, r, http.StatusBadRequest, r.Body)
		return
	}
	defer r.Body.Close()
	defer chain.Database.Close()

	newBlock := chain.UpdateBlock(m.Data, m.Hash)
	// fmt.Println("Block updated")

	respondWithJSON(w, r, http.StatusCreated, newBlock)

}

func handleGetBlock(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET,PUT,POST,DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Set-Cookie", "HttpOnly;Secure;SameSite=Strict")

	var m Message4
	nodeID := os.Getenv("NODE_ID")
	chain := blockchain.ContinueBlockChain(nodeID)

	fmt.Println("The data is: ", r.Body)
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m); err != nil {
		respondWithJSON(w, r, http.StatusBadRequest, r.Body)
		return
	}
	defer r.Body.Close()
	defer chain.Database.Close()

	block := chain.FindBlock(m.Hash)
	// fmt.Println("Block updated")

	respondWithJSON(w, r, http.StatusCreated, string(block.Data))

}

func respondWithJSON(w http.ResponseWriter, r *http.Request, code int, payload interface{}) {
	response, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("HTTP 500: Internal Server Error"))
		return
	}
	w.WriteHeader(code)
	w.Write(response)
}

// func NodeIsKnown(addr string) bool {
// 	for _, node := range KnownNodes {
// 		if node == addr {
// 			return true
// 		}
// 	}

// 	return false
// }
=======
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"runtime"
	"syscall"

	"github.com/tensor-programming/golang-blockchain/blockchain"
	"gopkg.in/vrecan/death.v3"
)

const (
	protocol      = "tcp"
	version       = 1
	commandLength = 12
)

var (
	nodeAddress     string
	mineAddress     string
	KnownNodes      = []string{"localhost:3000"}
	blocksInTransit = [][]byte{}
	memoryPool      = make(map[string]blockchain.Transaction)
)

type Addr struct {
	AddrList []string
}

type Block struct {
	AddrFrom string
	Block    []byte
}

type GetBlocks struct {
	AddrFrom string
}

type GetData struct {
	AddrFrom string
	Type     string
	ID       []byte
}

type Inv struct {
	AddrFrom string
	Type     string
	Items    [][]byte
}

type Tx struct {
	AddrFrom    string
	Transaction []byte
}

//Helps in syncing the blockchain across all nodes.
type Version struct {
	Version     int
	BestHeight  int
	BestVersion int
	AddrFrom    string
}

func CmdToBytes(cmd string) []byte {
	var bytes [commandLength]byte

	for i, c := range cmd {
		bytes[i] = byte(c)
	}

	return bytes[:]
}

func BytesToCmd(bytes []byte) string {
	var cmd []byte

	for _, b := range bytes {
		if b != 0x0 {
			cmd = append(cmd, b)
		}
	}

	return fmt.Sprintf("%s", cmd)
}

func ExtractCmd(request []byte) []byte {
	return request[:commandLength]
}

//Makes sure that all the blockchains in each node are synced with one another.
func RequestBlocks() {
	for _, node := range KnownNodes {
		SendGetBlocks(node)
	}
}

func SendAddr(address string) {
	nodes := Addr{KnownNodes}
	nodes.AddrList = append(nodes.AddrList, nodeAddress)
	payload := GobEncode(nodes)
	request := append(CmdToBytes("addr"), payload...)

	SendData(address, request)
}

func SendBlock(addr string, b *blockchain.Block) {
	data := Block{nodeAddress, b.Serialize()}
	payload := GobEncode(data)
	request := append(CmdToBytes("block"), payload...)

	SendData(addr, request)
}

func SendInv(address, kind string, items [][]byte) {
	inventory := Inv{nodeAddress, kind, items}
	payload := GobEncode(inventory)
	request := append(CmdToBytes("inv"), payload...)

	SendData(address, request)
}

func SendGetBlocks(address string) {
	payload := GobEncode(GetBlocks{nodeAddress})
	request := append(CmdToBytes("getblocks"), payload...)

	SendData(address, request)
}

func SendGetData(address, kind string, id []byte) {
	payload := GobEncode(GetData{nodeAddress, kind, id})
	request := append(CmdToBytes("getdata"), payload...)

	SendData(address, request)
}

func SendTx(addr string, tnx *blockchain.Transaction) {
	data := Tx{nodeAddress, tnx.Serialize()}
	payload := GobEncode(data)
	request := append(CmdToBytes("tx"), payload...)

	SendData(addr, request)
}

func SendVersion(addr string, chain *blockchain.BlockChain) {
	bestHeight := chain.GetBestHeight()
	bestVersion := chain.GetBestVersion()
	payload := GobEncode(Version{version, bestHeight, bestVersion, nodeAddress})

	request := append(CmdToBytes("version"), payload...)

	SendData(addr, request)
}

func SendData(addr string, data []byte) {
	conn, err := net.Dial(protocol, addr)

	if err != nil {
		fmt.Printf("%s is not available\n", addr)
		var updatedNodes []string

		for _, node := range KnownNodes {
			if node != addr {
				updatedNodes = append(updatedNodes, node)
			}
		}

		KnownNodes = updatedNodes

		return
	}

	defer conn.Close()

	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		log.Panic(err)
	}
}

func HandleAddr(request []byte) {
	var buff bytes.Buffer
	var payload Addr

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)

	}

	KnownNodes = append(KnownNodes, payload.AddrList...)
	fmt.Printf("there are %d known nodes\n", len(KnownNodes))
	RequestBlocks()
}

func HandleBlock(request []byte, chain *blockchain.BlockChain) {
	var buff bytes.Buffer
	var payload Block

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	blockData := payload.Block
	block := blockchain.Deserialize(blockData)

	fmt.Println("Recevied a new block!")
	chain.AddBlock(block)

	fmt.Printf("Added block %x\n", block.Hash)

	if len(blocksInTransit) > 0 {
		blockHash := blocksInTransit[0]
		SendGetData(payload.AddrFrom, "block", blockHash)

		blocksInTransit = blocksInTransit[1:]
	}
}

func HandleInv(request []byte, chain *blockchain.BlockChain) {
	var buff bytes.Buffer
	var payload Inv

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("Recevied inventory with %d %s\n", len(payload.Items), payload.Type)

	if payload.Type == "block" {
		blocksInTransit = payload.Items

		blockHash := payload.Items[0]
		SendGetData(payload.AddrFrom, "block", blockHash)

		newInTransit := [][]byte{}
		for _, b := range blocksInTransit {
			if bytes.Compare(b, blockHash) != 0 {
				newInTransit = append(newInTransit, b)
			}
		}
		blocksInTransit = newInTransit
	}

	if payload.Type == "tx" {
		txID := payload.Items[0]

		if memoryPool[hex.EncodeToString(txID)].ID == nil {
			SendGetData(payload.AddrFrom, "tx", txID)
		}
	}
}

func HandleGetBlocks(request []byte, chain *blockchain.BlockChain) {
	var buff bytes.Buffer
	var payload GetBlocks

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	blocks := chain.GetBlockHashes()
	SendInv(payload.AddrFrom, "block", blocks)
}

//Look into it while integrating this module.
func HandleGetData(request []byte, chain *blockchain.BlockChain) {
	var buff bytes.Buffer
	var payload GetData

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	if payload.Type == "block" {
		block, err := chain.GetBlock([]byte(payload.ID))
		if err != nil {
			return
		}

		SendBlock(payload.AddrFrom, &block)
	}

	//Need to make changes here.
	if payload.Type == "tx" {
		txID := hex.EncodeToString(payload.ID)
		tx := memoryPool[txID]

		SendTx(payload.AddrFrom, &tx)
	}
}

func HandleVersion(request []byte, chain *blockchain.BlockChain) {
	var buff bytes.Buffer
	var payload Version

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	bestHeight := chain.GetBestHeight()
	bestVersion := chain.GetBestVersion()

	otherHeight := payload.BestHeight
	otherVersion := payload.BestVersion

	if bestHeight < otherHeight || bestVersion < otherVersion {
		SendGetBlocks(payload.AddrFrom)
	} else if bestHeight > otherHeight || bestVersion > otherVersion {
		SendVersion(payload.AddrFrom, chain)
	}

	if !NodeIsKnown(payload.AddrFrom) {
		KnownNodes = append(KnownNodes, payload.AddrFrom)
	}
}

//Look into the following 2 functions (Major changes might be needed.)
func HandleTx(request []byte, chain *blockchain.BlockChain) {
	var buff bytes.Buffer
	var payload Tx

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	txData := payload.Transaction
	tx := blockchain.DeserializeTransaction(txData)
	memoryPool[hex.EncodeToString(tx.ID)] = tx

	fmt.Printf("%s, %d", nodeAddress, len(memoryPool))

	if nodeAddress == KnownNodes[0] {
		for _, node := range KnownNodes {
			if node != nodeAddress && node != payload.AddrFrom {
				SendInv(node, "tx", [][]byte{tx.ID})
			}
		}
	} else {
		if len(memoryPool) >= 2 && len(mineAddress) > 0 {
			// MineTx(chain)
		}
	}
}

func MineTx(transactions []*blockchain.Transaction, hash string) {
	// var txs []*blockchain.Transaction

	// // for id := range memoryPool {
	// // 	fmt.Printf("tx: %s\n", memoryPool[id].ID)
	// // 	tx := memoryPool[id]
	// // 	if chain.VerifyTransaction(&tx) {
	// // 		txs = append(txs, &tx)
	// // 	}
	// // }

	// // if len(txs) == 0 {
	// // 	fmt.Println("All Transactions are invalid")
	// // 	return
	// // }

	// // cbTx := blockchain.CoinbaseTx(mineAddress, "")
	// // txs = append(txs, cbTx)

	// newBlock := chain.AddTransactions(txs, hash)

	// // UTXOSet := blockchain.UTXOSet{chain}
	// // UTXOSet.Reindex()

	// fmt.Println("Block Updated")

	// // for _, tx := range txs {
	// // 	txID := hex.EncodeToString(tx.ID)
	// // 	delete(memoryPool, txID)
	// // }

	// for _, node := range KnownNodes {
	// 	if node != nodeAddress {
	// 		SendInv(node, "block", [][]byte{newBlock.Hash})
	// 	}
	// }

	// // if len(memoryPool) > 0 {
	// // 	MineTx(chain)
	// // }
}

func HandleConnection(conn net.Conn, chain *blockchain.BlockChain) {
	req, err := ioutil.ReadAll(conn)
	defer conn.Close()

	if err != nil {
		log.Panic(err)
	}
	command := BytesToCmd(req[:commandLength])
	fmt.Printf("Received %s command\n", command)

	switch command {
	case "addr":
		HandleAddr(req)
	case "block":
		HandleBlock(req, chain)
	case "inv":
		HandleInv(req, chain)
	case "getblocks":
		HandleGetBlocks(req, chain)
	case "getdata":
		HandleGetData(req, chain)
	case "tx":
		HandleTx(req, chain)
	case "version":
		HandleVersion(req, chain)
	default:
		fmt.Println("Unknown command")
	}

}

func StartServer(nodeID, minerAddress string) {
	nodeAddress = fmt.Sprintf("localhost:%s", nodeID)
	mineAddress = minerAddress
	ln, err := net.Listen(protocol, nodeAddress)
	if err != nil {
		log.Panic(err)
	}
	defer ln.Close()

	chain := blockchain.ContinueBlockChain(nodeID)
	defer chain.Database.Close()
	go CloseDB(chain)

	if nodeAddress != KnownNodes[0] {
		SendVersion(KnownNodes[0], chain)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Panic(err)
		}
		go HandleConnection(conn, chain)

	}
}

func NodeIsKnown(addr string) bool {
	for _, node := range KnownNodes {
		if node == addr {
			return true
		}
	}

	return false
}
>>>>>>> 5b9bd97b039a2d9261bad8a0738b116c80b9dca1

func GobEncode(data interface{}) []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(data)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

func CloseDB(chain *blockchain.BlockChain) {
	d := death.NewDeath(syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	d.WaitForDeathWithFunc(func() {
		defer os.Exit(1)
		defer runtime.Goexit()
		chain.Database.Close()
	})
}
