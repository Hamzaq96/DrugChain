package main

import (
	"os"

	"github.com/tensor-programming/golang-blockchain/cli"
)

func main() {
	defer os.Exit(0)
	cmd := cli.CommandLine{}
	cmd.Run()
}

// fun main() {
// 	err := godotenv.Load()
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	chain := blockchain.ContinueBlockChain("3000")
// 	defer chain.Database.Close()
// 	go CloseDB(chain)

// 	// bcServer = make(chan []Block)

// 	// // create genesis block
// 	// t := time.Now()
// 	// genesisBlock := Block{0, t.String(), 0, "", ""}
// 	// spew.Dump(genesisBlock)
// 	// Blockchain = append(Blockchain, genesisBlock)

// 	// start TCP and serve TCP server
// 	server, err := net.Listen("tcp", ":"+os.Getenv("ADDR"))
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer server.Close()

// 	for {
// 		conn, err := server.Accept()
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		go handleConn(conn)
// 	}
// }

// func handleConn(conn net.Conn) {

// 	defer conn.Close()

// 	io.WriteString(conn, "Enter a drug: ")

// 	scanner := bufio.NewScanner(conn)

// 	// take in BPM from stdin and add it to blockchain after conducting necessary validation
// 	go func() {
// 		for scanner.Scan() {
// 			drug := scanner.Text()
// 			// if err != nil {
// 			// 	log.Printf("%v not a number: %v", scanner.Text(), err)
// 			// 	continue
// 			// }

// 			chain := blockchain.ContinueBlockChain("3000")
// 			defer chain.Database.Close()
// 			chain.MineBlock(drugname)
// 			fmt.Println("Added Block!")

// 			io.WriteString(conn, "\nEnter a new BPM:")
// 		}
// 	}()

// }
