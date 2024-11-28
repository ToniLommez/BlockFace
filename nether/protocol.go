package nether

import (
	"fmt"
	"net"
	"sync"
)

var (
	i_am_leader = false

	clients = make(map[net.Conn]string)
	leaders = make(map[net.Conn]string)
	nodes   = make(map[net.Conn]string)

	clients_lock sync.Mutex
	leaders_lock sync.Mutex
	nodes_lock   sync.Mutex

	handlers = map[string]func(conn net.Conn){
		"LEADER?": handleLeaderRequisition,
		"PING":    handlePing,
		"PONG":    handlePong,
	}
)

func broadcast(message string) {
	for conn := range clients {
		sendMessage(message, conn)
	}

	for conn := range leaders {
		sendMessage(message, conn)
	}

	for conn := range nodes {
		sendMessage(message, conn)
	}
}

func dealWithRequisition(message string, conn net.Conn) {
	if handler, exists := handlers[message]; exists {
		handler(conn)
	} else {
		sendMessage("UNKNOWN_COMMAND", conn)
	}
}

func handleLeaderRequisition(conn net.Conn) {
	if i_am_leader {
		fmt.Printf("Respondendo que sou o lider\n")
		sendMessage("YES", conn)
		clientToNode(conn)
	} else {
		leader, _, _ := getAny(leaders)
		ipv6 := getIPv6(leader)
		fmt.Printf("Não sou o lider, devolvendo IP de um lider %v\n", ipv6)
		sendMessage(ipv6, conn)
	}
}

func handlePing(conn net.Conn) {
	sendMessage("PONG", conn)
}

func handlePong(conn net.Conn) {}

func StartAsLeader() error {
	i_am_leader = true
	fmt.Printf("Iniciando server como lider\n")
	return startServer()
}

func EnterToNetwork(ipv6 string) error {
	i_am_leader = false

	fmt.Printf("Conectando ao endereço\n")
	conn, err := connect(ipv6)
	if err != nil {
		return err
	}

	fmt.Printf("Perguntando se é o lider\n")
	sendMessage("LEADER?", conn)
	response, err := readMessage(conn)
	if err != nil {
		disconnectClient(conn)
		return err
	}

	if response == "YES" {
		fmt.Printf("É o lider!\n")
		clientToLeader(conn)
	} else {
		fmt.Printf("Não é o lider!\n")
		disconnectClient(conn)
		leader_ipv6 := response
		conn, err = connect(leader_ipv6)
		if err != nil {
			return err
		}
	}

	fmt.Printf("Iniciando server pessoal\n")
	go func() {
		if err := startServer(); err != nil {
			fmt.Println(err)
		}
	}()

	fmt.Printf("Entrando no modo de escuta\n")
	err = startChat(conn, leaders, removeLeader)

	return err
}

func PingAll() {
	broadcast("PING")
}
