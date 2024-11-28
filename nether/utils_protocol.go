package nether

import (
	"net"
	"sync"
)

func deleteFromMap[K comparable, V any](m map[K]V, mu *sync.Mutex, key K) {
	mu.Lock()
	delete(m, key)
	mu.Unlock()
}

func addToMap[K comparable, V any](m map[K]V, mu *sync.Mutex, key K, v V) {
	mu.Lock()
	m[key] = v
	mu.Unlock()
}

func getAny[K comparable, V any](m map[K]V) (K, V, bool) {
	for key, value := range m {
		return key, value, true
	}
	var zeroK K
	var zeroV V
	return zeroK, zeroV, false
}

func getIPv6(conn net.Conn) string {
	tcpAddr := conn.RemoteAddr().(*net.TCPAddr)
	return tcpAddr.IP.String()
}

func addClient(name string, conn net.Conn) {
	addToMap(clients, &clients_lock, conn, name)
}

func addLeader(name string, conn net.Conn) {
	addToMap(leaders, &leaders_lock, conn, name)
}

func addNode(name string, conn net.Conn) {
	addToMap(nodes, &nodes_lock, conn, name)
}

func removeClient(conn net.Conn) {
	deleteFromMap(clients, &clients_lock, conn)
}

func removeLeader(conn net.Conn) {
	deleteFromMap(leaders, &leaders_lock, conn)
}

func removeNode(conn net.Conn) {
	deleteFromMap(nodes, &nodes_lock, conn)
}

func clientToLeader(conn net.Conn) {
	addLeader(clients[conn], conn)
	removeClient(conn)
}

func clientToNode(conn net.Conn) {
	addNode(clients[conn], conn)
	removeClient(conn)
}

func disconnectClient(conn net.Conn) {
	removeClient(conn)
	conn.Close()
}

func disconnectLeader(conn net.Conn) {
	removeLeader(conn)
	conn.Close()
}

func disconnectNode(conn net.Conn) {
	removeNode(conn)
	conn.Close()
}
