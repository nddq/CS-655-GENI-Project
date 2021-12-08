#!/bin/sh
sudo apt update  
sudo apt install golang
go get -u github.com/gorilla/mux
go run main.go