# Simple TCP Server


## Build
`gb build`

## Run 
`./bin/server -h` to check default config and provide config.

`./bin/server` run server.


## Test
`nc 127.0.0.1 8080` Connect to server. Note the input contains '\n' at the end of string. Use `printf 'quit'|nc 127.0.0.1 8080` instead.

`nc -kl 18080 -c 'echo -e "HTTP/1.1 200 OK\r\n$(date)\r\n\r\n"'` Simple http server for mocking external API.

`./bin/mock_external`

## TODO

* mock client

* escape quit

* 
