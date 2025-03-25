package nether

import (
	"encoding/base64"
	"fmt"
	"net"
	"os"
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

	become_leader_after_election = false

	new_leaders      = make([]string, 0)
	new_leaders_lock sync.Mutex
)

func initHandlers() {
	handlers = map[string]func(conn net.Conn, parts []string){
		"LEADER?":         handleLeaderRequisition,
		"PING":            handlePing,
		"PONG":            func(conn net.Conn, parts []string) {},
		"ELECTION":        handleElection,
		"NEW_ELECTION":    handleElectionPreparing,
		"ELECTED":         handleElected,
		"WIN_ADVICE":      handleWinAdvice,
		"WIN":             handleWin,
		"WIN_ACCEPTED":    handleWinAccepted,
		"WIN_REJECTED":    handleWinRejected,
		"GET_BLOCKCHAIN":  handleGetBlockchain,
		"BLOCKCHAIN_DATA": handleBlockchainData,
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
	fmt.Printf("Pedido de identificação de lider recebido\n")
	if i_am_leader {
		fmt.Printf("Respondendo que sou lider\n")
		sendMessage("YES", conn)
	} else {
		leader, _, _ := getAny(leaders)
		ipv6 := getIPv6(leader)
		fmt.Printf("Respondendo que não sou lider, ip do lider sendo enviado: %s\n", ipv6)
		sendMessage(ipv6, conn)
	}
}

func handlePing(conn net.Conn, parts []string) {
	fmt.Printf("PING recebido, respondendo com PONG\n")
	sendMessage("PONG", conn)
}

func StartAsLeader() error {
	i_am_leader = true
	fmt.Printf("Iniciando e se auto intitulando lider da nova rede\n")

	return startServer()
}

func EnterToNetwork(ipv6 string) error {
	i_am_leader = false

	fmt.Printf("Tentando se conectar a %s\n", ipv6)
	conn, err := connect(ipv6)
	if err != nil {
		fmt.Printf("Erro encontrado: %v\n", err)
		return err
	}

	fmt.Printf("Conexão realizada, pergutando se é lider\n")
	sendMessage("LEADER?", conn)
	response, err := readMessage(conn)
	if err != nil {
		fmt.Printf("erro %v\n", err)
		disconnectClient(conn)
		return err
	}

	if response == "YES" {
		fmt.Printf("É lider, salvando\n")
		clientToLeader(conn)
	} else {
		disconnectClient(conn)
		leader_ipv6 := response
		fmt.Printf("Não é lider, novo endereço de lider recebido como resposta: %v\n", leader_ipv6)
		conn, err = connect(leader_ipv6)
		if err != nil {
			return err
		}
		clientToLeader(conn)
		fmt.Printf("Conectado ao novo lider\n")
	}

	go func() {
		fmt.Printf("Inicializando o proprio server para receber entradas\n")
		if err := startServer(); err != nil {
			fmt.Printf("%s\n", err)
		}
	}()

	fmt.Printf("Abrindo e mantendo a conexão com o novo lider\n")
	go startChat(conn, removeLeader)

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

func StartElection(numberOfLeaders int, numberOfZeroes int) error {
	if !i_am_leader {
		return fmt.Errorf("only leaders can start an election")
	}

	message := randomString(30, 40)
	fmt.Printf("Iniciando preparacao para eleicao!\nnumero de lideres: %2d, zeros: %2d, message[0:10]: %s\n", numberOfLeaders, numberOfZeroes, string(message[0:10]))

	requisition := fmt.Sprintf("NEW_ELECTION %d %d %s", numberOfLeaders, numberOfZeroes, message)
	broadcastLeaders(requisition)

	return nil
}

func handleElectionPreparing(conn net.Conn, parts []string) {
	under_election = true
	number_of_leaders, _ = strconv.Atoi(parts[1])
	election_zeroes, _ = strconv.Atoi(parts[2])
	election_message = parts[3]

	fmt.Printf("Liders se preparando para a eleicao e avisando os nodes\n")
	requisition := fmt.Sprintf("ELECTION %d %s", election_zeroes, election_message)

	broadcastNodes(requisition)
	broadcastLeaders(requisition)
}

func handleElection(conn net.Conn, parts []string) {
	election_zeroes, _ = strconv.Atoi(parts[1])
	election_message := []byte(parts[2])

	fmt.Printf("Processo de eleição iniciado, zeros: %8d, election_message[0:10] %s\n", election_zeroes, string(election_message[0:10]))
	fmt.Printf("Iniciando proof of work\n")
	nonce, found := proof_of_work(election_zeroes, election_message)
	if found {
		fmt.Printf("Eu fui o ganhador! nonce: %16s, hash: %s\n", nonce, getHash(election_message, nonce))
		message := fmt.Sprintf("WIN %s", nonce)
		leader, _, _ := getAny(leaders)
		sendMessage(message, leader)
	}
}

func handleWinAdvice(conn net.Conn, parts []string) {
	if !i_am_leader {
		fmt.Printf("only leaders can start -handle a win advice-\n")
		return
	}

	new_leaders_lock.Lock()

	leader_ipv6 := parts[1]
	new_leaders = append(new_leaders, leader_ipv6)
	fmt.Printf("Novo lider adicionado a lista: %s\n", leader_ipv6)

	if len(new_leaders) == number_of_leaders {
		fmt.Printf("Eleicao finalizada, avisando os nos dos lideres encontrados\n")
		message := "ELECTED"
		for _, leader := range new_leaders {
			message = fmt.Sprintf("%s %s", message, leader)
		}
		broadcast(message)
	}

	new_leaders_lock.Unlock()
}

func handleWin(conn net.Conn, parts []string) {
	if !i_am_leader {
		fmt.Printf("only leaders can start -handle a WIN-\n")
		return
	}

	if !under_election {
		fmt.Printf("cannot handle win while not on election\n")
		sendMessage("WIN_REJECTED", conn)
		return
	}

	nonce := parts[1]

	valid := validateProof([]byte(election_message), nonce, election_zeroes)
	if valid {
		fmt.Printf("Win valido encontrado, avisando lideres, e entregando o ACCEPT\n")
		remoteAddr := conn.RemoteAddr().String()
		host, _, _ := net.SplitHostPort(remoteAddr)
		message := fmt.Sprintf("WIN_ADVICE %s", host)
		broadcastLeaders(message)
		sendMessage("WIN_ACCEPTED", conn)
	}
}

func handleWinAccepted(conn net.Conn, parts []string) {
	fmt.Printf("Win aceito!\n")
	become_leader_after_election = true
}

func handleWinRejected(conn net.Conn, parts []string) {
	fmt.Printf("Win rejeitado!\n")
	become_leader_after_election = false
}

func handleElected(conn net.Conn, parts []string) {
	fmt.Printf("Cancelando mineração, outro líder já foi eleito.\n")
	STOP_PROCESSING = true
	if cancelFunc != nil {
		cancelFunc()
	}

	under_election = false
	number_of_leaders = 0
	election_zeroes = 0
	election_message = ""
	new_leaders = make([]string, 0)

	fmt.Printf("Desconectando de outros lideres.\n")
	for leader_conn := range leaders {
		disconnectLeader(leader_conn)
	}

	if become_leader_after_election {
		fmt.Printf("Vou me tornar lider\n")
		i_am_leader = true
	} else {
		fmt.Printf("Nao vou me tornar lider\n")
		i_am_leader = false
	}

	new_leaders := parts[1:]
	if len(new_leaders) == 0 {
		fmt.Printf("NENHUM LIDER ENCONTRADO! PANICO\n")
	}
	if i_am_leader {
		fmt.Printf("Eu sou lider e estou conectando nos outros lideres\n")
		new_leader, _ := chooseRandom(new_leaders)
		fmt.Printf("Conectando a: %s\n", new_leader)
		new_leader_conn, err := connect(new_leader)
		clientToLeader(new_leader_conn)
		if err == nil {
			go startChat(new_leader_conn, removeLeader)
		}
	} else {
		fmt.Printf("Nao sou lider e estou escolhendo um lider aleatorio para conectar\n")
		for _, leader := range new_leaders {
			fmt.Printf("Conectando a: %s\n", leader)
			new_leader_conn, err := connect(leader)
			clientToLeader(new_leader_conn)
			if err == nil {
				go startChat(new_leader_conn, removeLeader)
			}
		}
	}
}

func ShowConnections() error {
	for conn, name := range clients {
		fmt.Printf("clients = name: %s | conn %s\n", name[0:10], conn.RemoteAddr())
	}
	for conn, name := range leaders {
		fmt.Printf("leaders = name: %s | conn %s\n", name[0:10], conn.RemoteAddr())
	}
	for conn, name := range nodes {
		fmt.Printf("nodes   = name: %s | conn %s\n", name[0:10], conn.RemoteAddr())
	}
	return nil
}

func handleGetBlockchain(conn net.Conn, parts []string) {
	fmt.Printf("Solicitação de blockchain recebida, enviando arquivo...\n")
	filePath := "data/nether.chain"
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Erro ao ler arquivo blockchain: %v\n", err)
		return
	}

	encodedData := base64.StdEncoding.EncodeToString(data)
	sendMessage(fmt.Sprintf("BLOCKCHAIN_DATA %s", encodedData), conn)
}

func handleBlockchainData(conn net.Conn, parts []string) {
	if len(parts) < 2 {
		fmt.Printf("Dados da blockchain recebidos em formato inválido\n")
		return
	}

	// Reconstruir os dados da mensagem
	dataStr := strings.Join(parts[1:], " ")
	dataStr = strings.TrimSpace(dataStr) // Remover espaços extras

	fmt.Printf("Recebendo e salvando o arquivo blockchain...\n")

	// Decodificar os dados
	data, err := base64.StdEncoding.DecodeString(dataStr)
	if err != nil {
		fmt.Printf("Erro ao decodificar dados do arquivo: %v\n", err)
		return
	}

	// Salvar o arquivo
	filePath := "data/nether.chain"
	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		fmt.Printf("Erro ao salvar arquivo: %v\n", err)
		return
	}

	fmt.Printf("Blockchain salva com sucesso!\n")
}

func RequestBlockchain() {
	leaders_lock.Lock()
	defer leaders_lock.Unlock()

	leaderConn, _, leaderExists := getAny(leaders)
	if !leaderExists {
		fmt.Printf("Nenhum líder disponível para solicitar a blockchain.\n")
		return
	}

	fmt.Printf("Solicitando blockchain ao líder: %s\n", leaderConn.RemoteAddr())
	sendMessage("GET_BLOCKCHAIN", leaderConn)
}
