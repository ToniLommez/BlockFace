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

	handlers map[string]func(conn net.Conn, parts []string)

	under_election    = false
	number_of_leaders = 0
	election_zeroes   = 0
	election_message  = ""

	become_leader_after_election = true

	new_leaders      = make([]string, 0)
	new_leaders_lock sync.Mutex
)

func initHandlers() {
	handlers = map[string]func(conn net.Conn, parts []string){
		"LEADER?":      handleLeaderRequisition,
		"PING":         handlePing,
		"PONG":         func(conn net.Conn, parts []string) {},
		"ELECTION":     handleElection,
		"NEW_ELECTION": handleElectionPreparing,
		"ELECTED":      handleElected,
		"WIN_ADVICE":   handleWinAdvice,
		"WIN":          handleWin,
		"WIN_ACCEPTED": handleWinAccepted,
		"WIN_REJECTED": handleWinRejected,
	}
}

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

	go startChat(conn, leaders, removeLeader)

	return nil
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

func broadcastLeaders(message string) {
	for conn := range leaders {
		sendMessage(message, conn)
	}
}

func broadcastNodes(message string) {
	for conn := range nodes {
		sendMessage(message, conn)
	}
}

func PingAll() {
	broadcast("PING")
}

func StartElection(numberOfLeaders int) error {
	if !i_am_leader {
		return fmt.Errorf("only leaders can start an election")
	}

	zeroes := 16
	message := randomString(30, 40)
	requisition := fmt.Sprintf("NEW_ELECTION %d %d %s", numberOfLeaders, zeroes, message)

	fmt.Printf("Iniciando preparacao para eleicao!\nnumero de lideres: %2d, zeros: %2d, message[0:10]: %s\n", numberOfLeaders, zeroes, string(election_message[0:10]))
	broadcastLeaders(requisition)

	return nil
}

func handleElectionPreparing(conn net.Conn, parts []string) {
	under_election = true
	number_of_leaders, _ = strconv.Atoi(parts[1])
	election_zeroes, _ = strconv.Atoi(parts[2])
	election_message = parts[3]

	requisition := fmt.Sprintf("NEW_ELECTION %d %s", election_zeroes, election_message)

	broadcastNodes(requisition)
}

func handleElection(conn net.Conn, parts []string) {
	election_zeroes, _ = strconv.Atoi(parts[1])
	election_message := []byte(parts[2])

	fmt.Printf("Processo de eleição iniciado, zeros: %8d, election_message[0:10]%s\n", election_zeroes, string(election_message[0:10]))
	fmt.Printf("Iniciando proof of work\n")
	nonce, found := proof_of_work(election_zeroes, election_message)
	if found {
		fmt.Printf("Eu fui o ganhador! nonce: %16d, hash: %s\n", nonce, getHash(election_message, nonce))
		message := fmt.Sprintf("WIN %d", nonce)
		leader, _, _ := getAny(leaders)
		sendMessage(message, leader)
	}
}

func handleWinAdvice(conn net.Conn, parts []string) {
	if !i_am_leader {
		fmt.Println("only leaders can start -handle a win advice-")
		return
	}

	new_leaders_lock.Lock()

	leader_ipv6 := parts[1]
	new_leaders = append(new_leaders, leader_ipv6)

	if len(new_leaders) == number_of_leaders {
		message := "ELECTED"
		for _, leader := range new_leaders {
			message = fmt.Sprintf("%s %s", message, leader)
		}
		broadcastNodes(message)
	}
	new_leaders_lock.Unlock()
}

func handleWin(conn net.Conn, parts []string) {
	if !i_am_leader {
		fmt.Println("only leaders can start -handle a WIN-")
		return
	}

	if !under_election {
		fmt.Println("cannot handle win while not on election")
		sendMessage("WIN_REJECTED", conn)
		return
	}

	nonce, _ := strconv.Atoi(parts[1])

	valid := validateProof([]byte(election_message), nonce, election_zeroes)
	if valid {
		message := fmt.Sprintf("WIN_ADVICE %s", conn.RemoteAddr())
		broadcastLeaders(message)
		sendMessage("WIN_ACCEPTED", conn)
	}
}

func handleWinAccepted(conn net.Conn, parts []string) {
	become_leader_after_election = true
}

func handleWinRejected(conn net.Conn, parts []string) {
	become_leader_after_election = false
}

func handleElected(conn net.Conn, parts []string) {
	if cancelFunc != nil {
		fmt.Println("Cancelando mineração, outro líder já foi eleito.")
		cancelFunc()
	}

	under_election = false
	number_of_leaders = 0
	election_zeroes = 0
	election_message = ""
	new_leaders = make([]string, 0)

	for leader_conn := range leaders {
		disconnectLeader(leader_conn)
	}

	if become_leader_after_election {
		i_am_leader = true
	} else {
		i_am_leader = false
	}

	new_leaders := parts[1:]
	if len(new_leaders) == 0 {
		fmt.Printf("NENHUM LIDER ENCONTRADO! PANICO")
	}
	if i_am_leader {
		new_leader, _ := chooseRandom(new_leaders)
		new_leader_conn, err := connect(new_leader)
		clientToLeader(new_leader_conn)
		if err == nil {
			go startChat(new_leader_conn, leaders, removeLeader)
		}
	} else {
		for _, leader := range new_leaders {
			new_leader_conn, err := connect(leader)
			clientToLeader(new_leader_conn)
			if err == nil {
				go startChat(new_leader_conn, leaders, removeLeader)
			}
		}
	}
}
