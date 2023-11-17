#!/bin/bash

# Baixar binários do Hyperledger Fabric
curl -sSL https://bit.ly/2ysbOFE | bash -s

# Baixar imagens do Hyperledger Fabric
curl -sSL https://bit.ly/2ysbOFE | bash -s -- -s

# Adicionar binários ao PATH (certifique-se de ajustar o caminho conforme necessário)
export PATH=$PATH:$(pwd)/fabric-samples/bin

# Instalar o SDK Go para o Hyperledger Fabric
go get -u github.com/hyperledger/fabric-sdk-go

# Instalar o pacote Chaincode do Hyperledger Fabric para Go
go get -u github.com/hyperledger/fabric-contract-api-go/contractapi

# Mensagem de confirmação
echo "Pacotes do Hyperledger Fabric para Go foram instalados com sucesso!"
# Mensagem de confirmação
echo "Hyperledger Fabric binários e imagens foram instalados com sucesso!"
