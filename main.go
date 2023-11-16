package main

func main() {
	Chaincode-1, err := contractapi.NewChaincode(&MyContract{})
	if err != nil {
		fmt.Printf("Erro ao criar o chaincode: %s", err)
		return
	}

	if err := Chaincode-1.Start(); err != nil {
		fmt.Printf("Erro ao iniciar o chaincode: %s", err)
	}
}
