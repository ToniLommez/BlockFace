package nether

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"os"
	"sync"
	"time"
)

const (
	SERVER_ADRESS string = "data/server.conf"
	port          string = "666"
)

var (
	clients     = make(map[net.Conn]string)
	clientsLock sync.Mutex
)

func StartServer() {
	ipv6Address := GetValidIPv6Address()
	if ipv6Address == "" {
		fmt.Println("Nenhum endereço IPv6 válido encontrado.")
		return
	}

	listenAddress := fmt.Sprintf("[%s]:%s", ipv6Address, port)
	fmt.Println("Servidor aberto em: ", listenAddress)
	os.WriteFile(SERVER_ADRESS, []byte(listenAddress), 0644)

	listener, err := net.Listen("tcp", listenAddress)
	if err != nil {
		fmt.Println("Erro ao iniciar o servidor:", err)
		return
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Erro ao aceitar conexão:", err)
			continue
		}

		fmt.Println("Peer conectado:", conn.RemoteAddr())
		go handleConnection(conn)
	}
}

// Função para iniciar o cliente e conectar ao servidor
func StartClient(ipv6 string) {
	// Garante que o endereço IPv6 esteja entre colchetes
	if ipv6[0] != '[' && ipv6[len(ipv6)-1] != ']' {
		ipv6 = fmt.Sprintf("[%s]", ipv6)
	}

	serverAddress := fmt.Sprintf("%s:%s", ipv6, port)
	conn, err := net.Dial("tcp", serverAddress)
	if err != nil {
		fmt.Println("Erro ao conectar ao servidor:", err)
		return
	}
	fmt.Println("Conectado ao servidor em", serverAddress)

	_, err = conn.Write([]byte(EncodePublicKey(userdata.Key.Pk)))
	if err != nil {
		fmt.Println("Erro ao enviar mensagem:", err)
		return
	}

	// Inicia o chat bidirecional
	startChat(conn)
}

// Função para iniciar o chat bidirecional do cliente
func startChat(conn net.Conn) {
	defer conn.Close()

	// Inicia uma goroutine para receber mensagens
	go func() {
		reader := bufio.NewReader(conn)
		for {
			message, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Conexão encerrada:", err)
				break
			}
			fmt.Println("Mensagem recebida:", message)
		}
	}()

	// Lê e envia mensagens digitadas para o servidor
	for {
		time.Sleep(5 * time.Second)
		ping := rand.Intn(100) + 1

		_, err := conn.Write([]byte(fmt.Sprintf("Ping: %d\n", ping)))
		if err != nil {
			fmt.Println("Erro ao enviar mensagem:", err)
			return
		}

	}
}

// Função para lidar com uma conexão recebida e distribuir mensagens para outros clientes
func handleConnection(conn net.Conn) {
	defer func() {
		clientsLock.Lock()
		delete(clients, conn)
		clientsLock.Unlock()
		conn.Close()
		fmt.Println("Conexão encerrada com:", conn.RemoteAddr())
	}()

	reader := bufio.NewReader(conn)
	name, _ := reader.ReadString('\n')
	clientsLock.Lock()
	clients[conn] = name
	clientsLock.Unlock()

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Conexão encerrada:", err)
			break
		}
		fmt.Println("Mensagem recebida de:", conn.RemoteAddr(), ":", message)
	}
}

// Função para obter um único endereço IPv6 global válido
func GetValidIPv6Address() string {
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
