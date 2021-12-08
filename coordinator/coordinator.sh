#!/bin/sh
sudo apt update  
sudo apt install golang
go get -u github.com/gorilla/mux
cd CS-655-GENI-Project/coordinator
go run main.go