package network

import (
	"bytes"
	"encoding/gob"
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
