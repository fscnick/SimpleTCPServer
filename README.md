# Simple TCP Server

# Build
`gb build`

# Run 
`./bin/server -h` to check default config and provide config.

`./bin/server` run server.


# Test
`nc -kl 18080 -c 'echo -e "HTTP/1.1 200 OK\r\n$(date)\r\n\r\n"'` Simple http server for mocking external API.