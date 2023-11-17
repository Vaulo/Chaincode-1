// main.go 
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"crypto/sha256"
	"encoding/hex"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// TripData struct para representar os dados de uma viagem
type TripData struct {
	DepartureDatetime string  `json:"Departure_Datetime"`
	TotalDistanceKm   float64 `json:"totalDistance_km"`
	TripID            int     `json:"TripID"`
	ArrivalDatetime   string  `json:"Arrival_Datetime"`
}

// Transaction representa uma transação no livro-razão
type Transaction struct {
	Timestamp time.Time `json:"timestamp"`
	Data      string    `json:"data"`
}

// Block representa um bloco contendo várias transações
type Block struct {
	Transactions []Transaction `json:"transactions"`
}

// Blockchain representa uma sequência de blocos
type Blockchain struct {
	Blocks []Block `json:"blocks"`
}

// Bloco atual
var currentBlock *Block

// Último carimbo de data/hora de transação
var lastTransactionTimestamp time.Time

// Máximo de transações por bloco
const maxTransactionsPerBlock = 10

// Limite de tempo do bloco (10 minutos)
const blockTimeLimit = 10 * time.Minute

// MyContract define o chaincode para consulta de dados do MySQL e transações
type MyContract struct {
	contractapi.Contract
}

// QueryBanco function to query data from MySQL and add transactions to the ledger
func (mc *MyContract) QueryBanco(ctx contractapi.TransactionContextInterface) (*string, error) {
	// ... seu código existente ...

	return &jsonResult, nil
}

// AdicionarTransacao adiciona uma transação ao bloco atual
func (mc *MyContract) AdicionarTransacao(ctx contractapi.TransactionContextInterface, data string) error {
	if currentBlock == nil {
		currentBlock = &Block{
			Transactions: []Transaction{},
		}
	}

	transaction := Transaction{
		Timestamp: time.Now(),
		Data:      data,
	}

	currentBlock.Transactions = append(currentBlock.Transactions, transaction)
	lastTransactionTimestamp = transaction.Timestamp

	// Verificar se o número máximo de transações por bloco foi atingido
	if len(currentBlock.Transactions) >= maxTransactionsPerBlock {
		// Fechar o bloco e adicionar ao ledger
		err := mc.FecharBloco(ctx)
		if err != nil {
			return fmt.Errorf("Erro ao fechar o bloco: %v", err)
		}
	}

	return nil
}

// FecharBloco fecha o bloco atual se o limite de tempo ou número máximo de transações for atingido
func (mc *MyContract) FecharBloco(ctx contractapi.TransactionContextInterface) error {
	// Verificar se há transações no bloco atual
	if currentBlock == nil || len(currentBlock.Transactions) == 0 {
		return nil // Nenhum bloco a fechar
	}

	// Verificar se o tempo desde a última transação ultrapassou o limite
	if time.Since(lastTransactionTimestamp) >= blockTimeLimit {
		// Criar um novo bloco
		currentBlock = &Block{
			Transactions: []Transaction{},
		}
		return nil
	}

	// Obter a blockchain do estado
	blockchainJSON, err := ctx.GetStub().GetState("blockchain")
	if err != nil {
		return fmt.Errorf("Erro ao obter blockchain do estado: %v", err)
	}

	var blockchain Blockchain
	if blockchainJSON != nil {
		err = json.Unmarshal(blockchainJSON, &blockchain)
		if err != nil {
			return fmt.Errorf("Erro ao deserializar blockchain do JSON: %v", err)
		}
	}

	// Adicionar o bloco ao blockchain
	blockchain.Blocks = append(blockchain.Blocks, *currentBlock)

	// Serializar o blockchain para JSON
	blockchainJSON, err = json.Marshal(blockchain)
	if err != nil {
		return fmt.Errorf("Erro ao serializar blockchain para JSON: %v", err)
	}

	// Calcular o hash do bloco usando SHA-256
	hash := calcularHash(blockchainJSON)

	// Imprimir o hash do bloco
	fmt.Printf("Hash do Bloco: %s\n", hash)

	// Adicionar o blockchain ao estado
	err = ctx.GetStub().PutState("blockchain", blockchainJSON)
	if err != nil {
		return fmt.Errorf("Erro ao adicionar blockchain ao estado: %v", err)
	}

	currentBlock = &Block{
		Transactions: []Transaction{},
	}

	return nil
}

// Função auxiliar para calcular o hash usando SHA-256
func calcularHash(data []byte) string {
	hasher := sha256.New()
	hasher.Write(data)
	return hex.EncodeToString(hasher.Sum(nil))
}

func main() {
	Chaincodemove, err := contractapi.NewChaincode(&MyContract{})
	if err != nil {
		fmt.Printf("Erro ao criar o chaincode: %s", err)
		return
	}

	if err := Chaincodemove.Start(); err != nil {
		fmt.Printf("Erro ao iniciar o chaincode: %s", err)
	}
}
