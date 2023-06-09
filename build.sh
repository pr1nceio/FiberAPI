#!/bin/bash

RED='\033[0;31m'
BLUE='\033[0;34m'
GREEN='\033[0;32m'
NC='\033[0m'
GRAY='\033[0;37m'


echo -e "${GREEN}Resolving deps...${GRAY}"
go mod tidy
go install github.com/swaggo/swag/cmd/swag@latest
echo -e "${GREEN}Generating Swagger docs...${GRAY}"
swag init -g ./cmd/main.go --pd
echo -e "${GREEN}Building...${GRAY}"
CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath
R=fiber_api
echo -e "${GREEN}Pushing to $R...${GRAY}"
sudo docker build -t cr.yandex/crpr24jcqm2dno6qlm3b/$R . && docker push cr.yandex/crpr24jcqm2dno6qlm3b/$R
echo -e "${GREEN}Done.${NC}"