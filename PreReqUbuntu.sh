#!/bin/bash

echo -e "Instalando Artefatos de Compilação\n"

# Atualiza os repositórios
sudo apt update

# Instala as ferramentas de compilação
sudo apt install -y build-essential

echo -e "\nInstalando GoLang\n"

# Baixa e instala o Go
sudo rm -rf /opt/go
sudo curl -fsSL https://golang.org/dl/go1.15.12.linux-amd64.tar.gz | sudo tar -C /opt -xz

# Configura as variáveis de ambiente do Go
mkdir -p $HOME/go
echo "export GOPATH=$HOME/go" >> ~/.bashrc
echo "export GOROOT=/opt/go" >> ~/.bashrc

echo -e "\nInstalando NodeJs\n"

# Adiciona o repositório do Node.js e instala
curl -sL https://deb.nodesource.com/setup_10.x -o nodesource_setup.sh
sudo bash nodesource_setup.sh
sudo apt install -y nodejs

echo -e "\nInstalando Docker\n"

# Instala o Docker
curl -fsSL https://get.docker.com -o get-docker.sh && sudo sh get-docker.sh

# Adiciona o usuário ao grupo docker
sudo usermod -aG docker $USER

# Reinicia o serviço Docker
sudo systemctl restart docker.service

echo -e "\nInstalando Docker-Compose\n"

# Instala o Docker-Compose
sudo curl -L "https://github.com/docker/compose/releases/download/1.29.1/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Limpa os arquivos de instalação
rm -f go1.15.12.linux-amd64.tar.gz nodesource_setup.sh get-docker.sh

# Volta para o diretório home
cd $HOME

echo -e "\nPersonalizando variáveis de ambiente\n"

# Atualiza as variáveis de ambiente
source ~/.bashrc

echo -e "\nAmbiente configurado\n"

