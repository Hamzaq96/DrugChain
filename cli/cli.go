package cli

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"

	"github.com/tensor-programming/golang-blockchain/blockchain"
	"github.com/tensor-programming/golang-blockchain/wallet"
)

type CommandLine struct {
}

func (cli *CommandLine) printUsage() {
	fmt.Println("Usage:")
	fmt.Println(" add -block BLOCK_DATA - add a block to the chain")
	fmt.Println(" createblockchain -address ADDRESS creates a blockchain")
	fmt.Println(" print - Prints the blocks in the chain")
	fmt.Println("getblock -hash HASH - get the block")
	fmt.Println("updateblock -hash HASH - update the block")
	fmt.Println(" addtransaction -from FROM -to TO -hash HASH - Add transaction to the block")
	fmt.Println(" createwallet - Creates a new Wallet")
	fmt.Println(" listaddresses - Lists the addresses in our wallet file")
}

func (cli *CommandLine) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		runtime.Goexit()
	}
}

func (cli *CommandLine) listAddresses() {
	wallets, _ := wallet.CreateWallets()
	addresses := wallets.GetAllAddresses()

	for _, address := range addresses {
		fmt.Println(address)
	}
}

func (cli *CommandLine) createWallet() {
	wallets, _ := wallet.CreateWallets()
	address := wallets.AddWallet()
	wallets.SaveFile()

	fmt.Printf("New address is: %s\n", address)
}

func (cli *CommandLine) addBlock(data string) {
	chain := blockchain.ContinueBlockChain("")
	defer chain.Database.Close()
	chain.AddBlock(data)
	fmt.Println("Added Block!")
}

func (cli *CommandLine) printChain() {
	chain := blockchain.ContinueBlockChain("")
	defer chain.Database.Close()
	iter := chain.Iterator()

	for {
		block := iter.Next()

		fmt.Printf("Prev. hash: %x\n", block.PrevHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Printf("Nonce: %x\n", block.Nonce)
		pow := blockchain.NewProof(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		for _, tx := range block.Transactions {
			fmt.Println(tx)
		}
		fmt.Println()

		if len(block.PrevHash) == 0 {
			break
		}
	}
}

func (cli *CommandLine) createBlockChain(address string) {
	if !wallet.ValidateAddress(address) {
		log.Panic("Address is not valid")
	}
	chain := blockchain.InitBlockChain(address)
	chain.Database.Close()
	fmt.Println("Finished!")
}

func (cli *CommandLine) getBlock(hash string) {
	chain := blockchain.ContinueBlockChain(hash)
	defer chain.Database.Close()

	block := chain.FindBlock(hash)
	fmt.Printf("Data: %s\n", block.Data)
	fmt.Printf("Hash: %x\n", block.Hash)
}

func (cli *CommandLine) updateBlock(hash string) {
	chain := blockchain.ContinueBlockChain(hash)
	defer chain.Database.Close()

	chain.UpdateBlock(hash)
	fmt.Println("Block updated")

	// fmt.Printf("Data: %s\n", block.Data)
	// fmt.Printf("Hash: %x\n", block.Hash)

}

func (cli *CommandLine) addTransaction(from, hash string) {
	if !wallet.ValidateAddress(from) {
		log.Panic("Address is not valid")
	}
	chain := blockchain.ContinueBlockChain(hash)
	defer chain.Database.Close()

	tx := blockchain.NewTransaction(from, hash, chain)
	chain.AddTransactions([]*blockchain.Transaction{tx}, hash)
	fmt.Println("Transaction Added")

}

func (cli *CommandLine) Run() {
	cli.validateArgs()

	addBlockCmd := flag.NewFlagSet("add", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("print", flag.ExitOnError)
	addBlockData := addBlockCmd.String("block", "", "Block data")
	getBlockCmd := flag.NewFlagSet("getblock", flag.ExitOnError)
	updateBlockCmd := flag.NewFlagSet("updateblock", flag.ExitOnError)
	addTransactionCmd := flag.NewFlagSet("addtransaction", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	listAddressesCmd := flag.NewFlagSet("listaddresses", flag.ExitOnError)

	getBlockHash := getBlockCmd.String("hash", "", "The hash to get block for")
	updateBlockHash := updateBlockCmd.String("hash", "", "The hash to get block for")
	sendFrom := addTransactionCmd.String("from", "", "Source wallet address")
	sendTo := addTransactionCmd.String("to", "", "Destination block hash")
	createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send genesis block reward to")

	switch os.Args[1] {
	case "add":
		err := addBlockCmd.Parse(os.Args[2:])
		blockchain.Handle(err)

	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}

	case "print":
		err := printChainCmd.Parse(os.Args[2:])
		blockchain.Handle(err)

	case "getblock":
		err := getBlockCmd.Parse(os.Args[2:])
		blockchain.Handle(err)

	case "updateblock":
		err := updateBlockCmd.Parse(os.Args[2:])
		blockchain.Handle(err)

	case "listaddresses":
		err := listAddressesCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}

	case "addtransaction":
		err := addTransactionCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}

	default:
		cli.printUsage()
		runtime.Goexit()
	}

	if addBlockCmd.Parsed() {
		if *addBlockData == "" {
			addBlockCmd.Usage()
			runtime.Goexit()
		}
		cli.addBlock(*addBlockData)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			runtime.Goexit()
		}
		cli.createBlockChain(*createBlockchainAddress)
	}

	if getBlockCmd.Parsed() {
		if *getBlockHash == "" {
			getBlockCmd.Usage()
			runtime.Goexit()
		}
		cli.getBlock(*getBlockHash)
	}

	if updateBlockCmd.Parsed() {
		if *updateBlockHash == "" {
			updateBlockCmd.Usage()
			runtime.Goexit()
		}
		cli.updateBlock(*updateBlockHash)
	}

	if createWalletCmd.Parsed() {
		cli.createWallet()
	}
	if listAddressesCmd.Parsed() {
		cli.listAddresses()
	}

	if addTransactionCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" {
			addTransactionCmd.Usage()
			runtime.Goexit()
		}

		cli.addTransaction(*sendFrom, *sendTo)
	}
}
