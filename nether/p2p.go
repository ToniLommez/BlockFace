package nether

import (
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"
)

const (
	SERVER_ADRESS string = "data/server.conf"
	SERVER_PORT   string = "666"
	HEADER_SIZE   int    = 128
	BUFFER_SIZE          = 4096
	PAYLOAD_SIZE         = BUFFER_SIZE - 128
	MESSAGE_END          = "<END_OF_MESSAGE>"
)

var (
	self_ipv6 = ""
)

func startServer() error {
	ipv6 := GetIPv6()
	listenAddress, err := ProcessIpv6(ipv6)
	if err != nil {
		return err
	}

	self_ipv6 = ipv6
	listener, err := net.Listen("tcp", listenAddress)
	if err != nil {
		fmt.Println("Erro ao iniciar o servidor:", err)
		return err
	}

	go handleServerConnections(listener)

	if i_am_leader {
		time.Sleep(1 * time.Second)
		selfConn, err := connect(ipv6)
		if err != nil {
			return err
		}
		clientToLeader(selfConn)
	}

	return nil
}

func handleServerConnections(listener net.Listener) {
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Erro ao aceitar conexão:", err)
			continue
		}

		fmt.Println("Peer conectado:", conn.RemoteAddr())
		go serverHandle(conn)
	}
}

func connect(ipv6 string) (net.Conn, error) {
	serverAddress := DigestIpv6(ipv6)

	localAddr := &net.TCPAddr{
		IP: net.ParseIP(self_ipv6),
	}

	dialer := &net.Dialer{
		LocalAddr: localAddr,
		Timeout:   5 * time.Second,
	}

	conn, err := dialer.Dial("tcp", serverAddress)

	if err != nil {
		fmt.Println("Erro ao conectar ao servidor:", err)
		return nil, err
	}

	fmt.Printf("Conexao tcp realizada com %v\n", serverAddress)
	return clientHandle(conn), nil
}

func serverHandle(conn net.Conn) {
	name, _ := readMessage(conn)
	sendSelfId(conn)

	if i_am_leader {
		ip, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
		if ip == self_ipv6 {
			addLeader(name, conn)
		} else {
			addNode(name, conn)
		}
		startChat(conn, removeNode)
	} else {
		addClient(name, conn)
		startChat(conn, removeClient)
	}
}

func clientHandle(conn net.Conn) net.Conn {
	sendSelfId(conn)
	name, _ := readMessage(conn)

	addClient(name, conn)
	return conn
}

func sendSelfId(conn net.Conn) error {
	message := EncodePublicKey(userdata.Key.Pk)

	err := sendMessage(message, conn)
	if err != nil {
		fmt.Printf("Erro ao enviar SelfID: %v\n", err)
		return err
	}

	return nil
}

func startChat(conn net.Conn, remove func(conn net.Conn)) error {
	defer remove(conn)

	for {
		msg, err := readMessage(conn)
		if err != nil {
			break
		}
		go dealWithRequisition(msg, conn)
	}

	return nil
}

func GetIPv6() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("Erro ao obter interfaces:", err)
		return ""
	}

	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		jump := true
		for _, addr := range addrs {
			ip, _, err := net.ParseCIDR(addr.String())
			if err != nil {
				continue
			}

			if ip.To4() == nil && ip.IsGlobalUnicast() && !ip.IsLoopback() {
				if jump {
					jump = false
					continue
				}
				return ip.String()
			}
		}
	}
	return ""
}

func DigestIpv6(ipv6 string) string {
	if ipv6[0] != '[' && ipv6[len(ipv6)-1] != ']' {
		ipv6 = fmt.Sprintf("[%s]:%s", ipv6, SERVER_PORT)
	}

	return ipv6
}

func ProcessIpv6(ipv6Address string) (string, error) {
	if ipv6Address == "" {
		error := "nenhum endereço IPv6 válido encontrado"
		fmt.Println(error)
		return "", fmt.Errorf("%v", error)
	}

	listenAddress := DigestIpv6(ipv6Address)
	fmt.Println("Servidor aberto em:", listenAddress)
	// os.WriteFile(SERVER_ADRESS, []byte(listenAddress), 0644)

	return listenAddress, nil
}

func sendMessage(message string, conn net.Conn) error {
	messageID := rand.Int63() // Gera um ID único para a mensagem
	// Dividir a mensagem em fragmentos
	fragments := []string{}
	for len(message) > 0 {
		if len(message) > PAYLOAD_SIZE {
			fragments = append(fragments, message[:PAYLOAD_SIZE])
			message = message[PAYLOAD_SIZE:]
		} else {
			fragments = append(fragments, message)
			message = ""
		}
	}
	totalFragments := len(fragments)
	// Enviar cada fragmento com cabeçalho
	for i, fragment := range fragments {
		header := fmt.Sprintf("ID:%d PART:%d/%d ", messageID, i+1, totalFragments)
		paddedFragment := fragment + MESSAGE_END
		padding := PAYLOAD_SIZE - len(fragment)
		paddedFragment += strings.Repeat(" ", padding)

		fullMessage := header + paddedFragment
		_, err := conn.Write([]byte(fullMessage))
		if err != nil {
			fmt.Printf("[ERROR] Erro ao enviar fragmento %d: %v\n", i+1, err)
			return fmt.Errorf("erro ao enviar fragmento %d: %v", i+1, err)
		}
	}
	return nil
}

func readMessage(conn net.Conn) (string, error) {
	messageParts := make(map[int]string)
	var messageID int64
	var totalFragments int
	fmt.Println("[DEBUG] Iniciando leitura da mensagem")

	for {
		buffer := make([]byte, BUFFER_SIZE)
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Printf("[ERROR] Erro ao ler do buffer: %v\n", err)
			return "", err
		}

		message := string(buffer[:n])
		// Localizar o delimitador de fim de mensagem
		endIndex := strings.Index(message, MESSAGE_END)
		if endIndex == -1 {
			fmt.Printf("[ERROR] Delimitador %q não encontrado na mensagem\n", MESSAGE_END)
			return "", fmt.Errorf("delimitador %q não encontrado na mensagem", MESSAGE_END)
		}

		// Remover padding extra
		message = strings.TrimSpace(message[:endIndex])
		// Separar cabeçalho e corpo
		headerEnd := strings.Index(message, " ")
		if headerEnd == -1 {
			fmt.Println("[ERROR] Cabeçalho inválido ou mensagem incompleta")
			return "", fmt.Errorf("cabeçalho inválido ou mensagem incompleta")
		}

		header := message[:headerEnd]
		body := message[headerEnd+1:]
		// Parse do cabeçalho completo
		var partNumber int
		_, err = fmt.Sscanf(header+" "+body, "ID:%d PART:%d/%d", &messageID, &partNumber, &totalFragments)
		if err != nil {
			fmt.Printf("[ERROR] Erro ao parsear cabeçalho: %v\n", err)
			return "", fmt.Errorf("erro ao parsear cabeçalho: %v", err)
		}
		// Salvar a parte da mensagem
		messageParts[partNumber] = body[strings.Index(body, " ")+1:]
		// Verificar se todas as partes foram recebidas
		if len(messageParts) == totalFragments {
			fmt.Println("[DEBUG] Todos os fragmentos recebidos")
			break
		}
	}

	// Reconstituir a mensagem
	var reconstructedMessage strings.Builder
	for i := 1; i <= totalFragments; i++ {
		part, exists := messageParts[i]
		if !exists {
			fmt.Printf("[ERROR] Fragmento %d está faltando\n", i)
			return "", fmt.Errorf("fragmento %d está faltando", i)
		}
		reconstructedMessage.WriteString(part)
	}

	fmt.Println("[DEBUG] Mensagem reconstruída com sucesso")
	return reconstructedMessage.String(), nil
}
