package nether

import (
	"fmt"
	"net"
	"strconv"
	"strings"
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

	handlers = map[string]func(conn net.Conn, parts []string){
		"LEADER?":  handleLeaderRequisition,
		"PING":     handlePing,
		"PONG":     func(conn net.Conn, parts []string) {},
		"ELECTION": handleElection,
	}
)

func dealWithRequisition(message string, conn net.Conn) {
	parts := strings.Fields(message)
	command := parts[0]

	if handler, exists := handlers[command]; exists {
		handler(conn, parts)
	} else {
		sendMessage("UNKNOWN_COMMAND", conn)
	}
}

func handleLeaderRequisition(conn net.Conn, parts []string) {
	if i_am_leader {
		sendMessage("YES", conn)
	} else {
		leader, _, _ := getAny(leaders)
		ipv6 := getIPv6(leader)
		sendMessage(ipv6, conn)
	}
}

func handlePing(conn net.Conn, parts []string) {
	sendMessage("PONG", conn)
}

func StartAsLeader() error {
	i_am_leader = true
	return startServer()
}

func EnterToNetwork(ipv6 string) error {
	i_am_leader = false

	conn, err := connect(ipv6)
	if err != nil {
		return err
	}

	sendMessage("LEADER?", conn)
	response, err := readMessage(conn)
	if err != nil {
		disconnectClient(conn)
		return err
	}

	if response == "YES" {
		clientToLeader(conn)
	} else {
		disconnectClient(conn)
		leader_ipv6 := response
		conn, err = connect(leader_ipv6)
		if err != nil {
			return err
		}
	}

	go func() {
		if err := startServer(); err != nil {
			fmt.Println(err)
		}
	}()

	err = startChat(conn, leaders, removeLeader)

	return err
}

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

func PingAll() {
	broadcast("PING")
}

func StartElection() error {
	if !i_am_leader {
		return fmt.Errorf("only liders can start an election")
	}

	zeroes := 16
	hash := randomString(30, 40)
	message := fmt.Sprintf("ELECTION %d %s", zeroes, hash)

	broadcast(message)

	// todo: leader participate

	return nil
}

func handleElection(conn net.Conn, parts []string) {
	zeroes, _ := strconv.Atoi(parts[1])
	hash := []byte(parts[2])

	nonce, found := proof_of_work(zeroes, hash)
	if found {
		message := fmt.Sprintf("WIN %d", nonce)
		leader, _, _ := getAny(leaders)
		sendMessage(message, leader)
	}
}

func receiveElectionMessage(message string) {
	if message == "ELECTED" {
		if cancelFunc != nil {
			fmt.Println("Cancelando mineração, outro líder já foi eleito.")
			cancelFunc()
		}
	}
}
