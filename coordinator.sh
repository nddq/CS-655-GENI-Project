#!/bin/sh

sudo apt update  
sudo apt install golang
go get -u github.com/gorilla/mux
git clone https://github.com/nddq/CS-655-GENI-Project.git
cd CS-655-GENI-Project/coordinator
go run main.go