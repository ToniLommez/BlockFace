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
	MESSAGE_END   string = "<END_OF_MESSAGE>"
	BUFFER_SIZE   int    = 4096
	HEADER_SIZE   int    = 128
	PAYLOAD_SIZE  int    = BUFFER_SIZE - HEADER_SIZE
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
			return fmt.Errorf("erro ao enviar fragmento %d: %v", i+1, err)
		}
	}
	return nil
}

func readMessage(conn net.Conn) (string, error) {
	messageParts := make(map[int]string)
	var messageID int64
	var totalFragments int

	for {
		buffer := make([]byte, BUFFER_SIZE)
		n, err := conn.Read(buffer)
		if err != nil {
			return "", err
		}

		message := string(buffer[:n])

		// Extrair cabeçalho
		headerEnd := strings.Index(message, " ")
		if headerEnd == -1 {
			return "", fmt.Errorf("cabeçalho inválido na mensagem")
		}
		header := message[:headerEnd]
		body := message[headerEnd+1:]

		// Parse do cabeçalho
		var partNumber int
		_, err = fmt.Sscanf(header, "ID:%d PART:%d/%d", &messageID, &partNumber, &totalFragments)
		if err != nil {
			return "", fmt.Errorf("erro ao parsear cabeçalho: %v", err)
		}

		// Localizar o delimitador
		endIndex := strings.Index(body, MESSAGE_END)
		if endIndex == -1 {
			return "", fmt.Errorf("delimitador %q não encontrado na mensagem", MESSAGE_END)
		}

		// Salvar a parte da mensagem
		messageParts[partNumber] = strings.TrimSpace(body[:endIndex])

		// Verificar se todas as partes foram recebidas
		if len(messageParts) == totalFragments {
			break
		}
	}

	// Reconstituir a mensagem
	var reconstructedMessage strings.Builder
	for i := 1; i <= totalFragments; i++ {
		part, exists := messageParts[i]
		if !exists {
			return "", fmt.Errorf("fragmento %d está faltando", i)
		}
		reconstructedMessage.WriteString(part)
	}

	return reconstructedMessage.String(), nil
}

func sendSelfId(conn net.Conn) error {
	message := EncodePublicKey(userdata.Key.Pk) + "\n"
	_, err := conn.Write([]byte(message))
	if err != nil {
		fmt.Println("Erro ao enviar mensagem:", err)
	}
	return err
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
