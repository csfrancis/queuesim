package main

var numClients = 100

func main() {
	clients := make([]*Client, numClients)
	for i := 0; i < numClients; i++ {
		clients[i] = NewClient()
	}

	reactor := &Reactor{
		Clients: clients,
		Backend: NewRandomBackend(10),
	}

	reactor.Run()
}
