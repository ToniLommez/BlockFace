package nether

import (
	"bufio"
	"fmt"
	"net"
	"sync"
)

const (
	SERVER_ADRESS string = "data/server.conf"
	SERVER_PORT   string = "666"
)

var (
	clients     = make(map[net.Conn]string)
	clientsLock sync.Mutex
)

func StartConnections() error {
	ipv6 := GetIPv6()
	listenAddress, err := ProcessIpv6(ipv6)
	if err != nil {
		return err
	}

	listener, err := net.Listen("tcp", listenAddress)
	if err != nil {
		fmt.Println("Erro ao iniciar o servidor:", err)
		return err
	}

	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Erro ao aceitar conexão:", err)
			continue
		}

		fmt.Println("Peer conectado:", conn.RemoteAddr())
		go handleConnection(conn, true)
	}
}

func Connect(ipv6 string) error {
	serverAddress := DigestIpv6(ipv6)

	conn, err := net.Dial("tcp", serverAddress)
	if err != nil {
		fmt.Println("Erro ao conectar ao servidor:", err)
		return err
	}

	handleConnection(conn, false)
	return nil
}

func handleConnection(conn net.Conn, isReceiver bool) {
	var name string

	if isReceiver {
		name = readMessage(conn)
		sendSelfId(conn)
	} else {
		sendSelfId(conn)
		name = readMessage(conn)
	}

	addClient(name, conn)
	startChat(conn)
}

func addClient(name string, conn net.Conn) {
	clientsLock.Lock()
	clients[conn] = name
	clientsLock.Unlock()
}

func readMessage(conn net.Conn) string {
	msg, _ := bufio.NewReader(conn).ReadString('\n')
	return msg
}

func sendSelfId(conn net.Conn) error {
	message := EncodePublicKey(userdata.Key.Pk) + "\n"
	_, err := conn.Write([]byte(message))
	if err != nil {
		fmt.Println("Erro ao enviar mensagem:", err)
	}
	return err
}

func startChat(conn net.Conn) error {
	defer func() {
		clientsLock.Lock()
		delete(clients, conn)
		clientsLock.Unlock()
		conn.Close()
	}()

	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		fmt.Println("Mensagem recebida de:", conn.RemoteAddr(), ":", message)
	}

	return nil
}

func PingAll() {
	for conn := range clients {
		message := "Ping" + "\n"
		_, err := conn.Write([]byte(message))
		if err != nil {
			fmt.Println("Erro ao enviar mensagem:", err)
			return
		}
	}
}

// GetIPv6 obtem um único endereço IPv6 global válido e não local
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

		for _, addr := range addrs {
			ip, _, err := net.ParseCIDR(addr.String())
			if err != nil {
				continue
			}

			if ip.To4() == nil && ip.IsGlobalUnicast() && !ip.IsLoopback() {
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
		return "", fmt.Errorf(error)
	}

	listenAddress := DigestIpv6(ipv6Address)
	fmt.Println("Servidor aberto em: ", listenAddress)
	// os.WriteFile(SERVER_ADRESS, []byte(listenAddress), 0644)

	return listenAddress, nil
}
