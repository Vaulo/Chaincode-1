package main

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
