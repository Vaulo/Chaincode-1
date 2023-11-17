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
func (mc *MyContract) QueryBanco(ctx contractapi.TransactionContextInterface) (*[]byte, error) {
	// Conexão com o MySQL
	db, err := sql.Open("mysql", "root:movepass@tcp(localhost:3306)/moveuff")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Query SQL
	queryToday := `
	SELECT 
		departure.id AS Departure_Datetime, 
		trips.totalDistance_km, 
		trips.id AS TripID, 
		arrival.id AS Arrival_Datetime
	FROM trip_x_parkingslot_departures AS departure
	JOIN trips ON departure.Trips_id = trips.id
	JOIN trip_x_parkingslot_arrivals AS arrival ON arrival.Trips_id = trips.id
	WHERE DATE(departure.id) = CURDATE() AND DATE(arrival.id) = CURDATE()
`

	// Executar a query
	rows, err := db.Query(queryToday)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// Variável para armazenar o dicionário JSON
	resultDict := make(map[int]map[string]interface{})

	// Variável para armazenar a soma de totalDistance_km
	totalDistanceSum := 0.0

	for rows.Next() {
		var departureDatetime string // Modificado para usar string
		var totalDistanceKm float64
		var tripID int
		var arrivalDatetime string // Modificado para usar string

		// Ler os valores do resultado da query
		err := rows.Scan(&departureDatetime, &totalDistanceKm, &tripID, &arrivalDatetime)
		if err != nil {
			log.Fatal(err)
		}

		// Criar um mapa com os dados da linha atual
		rowData := map[string]interface{}{
			"Departure_Datetime": departureDatetime,
			"totalDistance_km":   totalDistanceKm,
			"TripID":             tripID,
			"Arrival_Datetime":   arrivalDatetime,
		}

		// Adicionar os dados ao objeto do dicionário JSON
		resultDict[tripID] = rowData

		// Somar o valor de totalDistance_km
		totalDistanceSum += totalDistanceKm
	}

	// Verificar erros na iteração sobre as linhas
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	// Imprimir o dicionário JSON
	jsonResult, err := json.MarshalIndent(resultDict, "", "    ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(jsonResult))

	// Imprimir a soma de totalDistance_km
	fmt.Printf("Soma de totalDistance_km: %.2f\n", totalDistanceSum)

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
