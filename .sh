#!/bin/bash

case $1 in

	"build")
		go build -o /build/main main.go
	;;

	"dev")
		APP_ENV=development go run main.go
	;;

  	*)
		go run main.go
    ;;

esac