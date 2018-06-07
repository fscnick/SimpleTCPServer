
# Simple TCP Server

This is a simple forward server. Transfer the received text to external server wrapped with HTTP . It also provides API for querying server status.

  

## Build

Enviroment: go version go1.9.4 linux/amd64

  
Build command: `gb build`

  

## Run

* `./bin/server -h` 
Check default config and provide config.

* `./bin/server` 
  Run server. Serve a TCP server receiving text from client and transfer the text wrapped with HTTP to external API server. Also, serve HTTP on another port for investigating the current status.

## Test

Choose the way preferred.

* `nc 127.0.0.1 8080` 
Connect to server. 
Note: the input contains '\n' at the end of string. Use `printf 'quit'|nc 127.0.0.1 8080` instead.


* `./bin/mock_external` 
  Mock external API. 

  

## TODO

* ~~mock external~~

* mock client

* escape quit