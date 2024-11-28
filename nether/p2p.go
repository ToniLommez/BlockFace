package nether

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

const (
	SERVER_ADRESS string = "data/server.conf"
	SERVER_PORT   string = "666"
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

	conn, err := net.Dial("tcp", serverAddress)
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
		startChat(conn, nodes, removeNode)
	} else {
		addClient(name, conn)
		startChat(conn, clients, removeClient)
	}
}

func clientHandle(conn net.Conn) net.Conn {
	sendSelfId(conn)
	name, _ := readMessage(conn)

	addClient(name, conn)
	return conn
}

func readMessage(conn net.Conn) (string, error) {
	msg, err := bufio.NewReader(conn).ReadString('\n')
	return strings.TrimSuffix(msg, "\n"), err
}

func sendMessage(message string, conn net.Conn) error {
	message += "\n"
	_, err := conn.Write([]byte(message))
	return err
}

func sendSelfId(conn net.Conn) error {
	message := EncodePublicKey(userdata.Key.Pk) + "\n"
	_, err := conn.Write([]byte(message))
	if err != nil {
		fmt.Println("Erro ao enviar mensagem:", err)
	}
	return err
}

func startChat(conn net.Conn, m map[net.Conn]string, remove func(conn net.Conn)) error {
	defer remove(conn)

	for {
		msg, err := readMessage(conn)
		if err != nil {
			break
		}
		fmt.Println("Mensagem recebida:", msg)
		dealWithRequisition(msg, conn)
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
	fmt.Println("Servidor aberto em:", listenAddress)
	// os.WriteFile(SERVER_ADRESS, []byte(listenAddress), 0644)

	return listenAddress, nil
}
