## Client

The client connects to the server at the `/cotacao` endpoint, retrieves the latest bid, and saves it to `cotacao.txt`.

To run the client:

```sh
go run client.go
```

## Server

The server starts a web server on port 8080 and listens for incoming connections. When a request is received at the `/cotacao` endpoint, it fetches data from an external API, saves the bid to a database, and sends it to the client.

To run the server:

```sh
go run server.go
```
