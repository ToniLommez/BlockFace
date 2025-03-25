package main

import (
	"bufio"
	"fmt"
	"nether/nether"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

func help() {
	fmt.Println("Commands:")
	fmt.Println("help ---------------- You're here")
	fmt.Println("clear --------------- clear the console")
	fmt.Println("register ------------ register a new user")
	fmt.Println("login --------------- login with a user")
	fmt.Println("test userdata ------- test userdata sanity checks")
	fmt.Println("show userdata ------- print userdata in console")
	fmt.Println("new blockchain ------ create a new blockchain")
	fmt.Println("load blockchain ----- load blockchain from secundary memory to primary memory")
	fmt.Println("show blockchain ----- print blockchain in console")
	fmt.Println("start server -------- start server")
	fmt.Println("start client -------- start client")
	fmt.Println("ping all ------------ ping all connections")
	fmt.Println("start election ------ start election for new leaders(only a leader can start an election)")
	fmt.Println("show connections ---- show all connections of the node")
	fmt.Println("download blockchain - download blockchain from a leader")
	fmt.Println("start endpoint ------ start endpoint connection to receive camera aplication infos")
	fmt.Println("exit ---------------- exit the program")
}

func clear() {
	// Determina o sistema operacional
	switch runtime.GOOS {
	case "windows":
		// Comando para limpar terminal no Windows
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	default:
		// Comando para limpar terminal no Linux e macOS
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

func input(prompt string) string {
	fmt.Println(prompt)
	data, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	return strings.TrimSpace(data)
}

func inputNumber(prompt string) int {
	for {
		fmt.Println(prompt)
		data, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		data = strings.TrimSpace(data)

		num, err := strconv.Atoi(data)
		if err != nil {
			fmt.Println("Entrada inválida, tente novamente.")
			continue
		}
		return num
	}
}

func register() {
	nether.Register(input("Type your password: "))
}

func testUserdata() {
	fmt.Println(nether.GetUserdata())

	fmt.Println("Resetting userdata")
	nether.ResetUserdata()
	fmt.Println(nether.GetUserdata())

	fmt.Println("Reopening ...")
	load()
	fmt.Println(nether.GetUserdata())
}

func showUserdata() {
	fmt.Println(nether.GetUserdata())
}

func load() {
	if success := nether.LoadData(input("Please type your password: ")); success {
		fmt.Println("Successfully logged")
	} else {
		fmt.Println("Wrong password")
	}
}

func autoLogin() {
	nether.LoadData("teste123")
	nether.LoadBlockchain()
}

func startServer() {
	go func() {
		if err := nether.StartAsLeader(); err != nil {
			fmt.Println(err)
		}
	}()
}

func startClient() {
	ipv6 := input("Type the ipv6 server address: ")
	go func() {
		if err := nether.EnterToNetwork(ipv6); err != nil {
			fmt.Println(err)
		}
	}()
}

func pingAll() {
	nether.PingAll()
}

func startElection() {
	numberOfLeaders := inputNumber("Type the number of leaders: ")
	numberOfZeroes := inputNumber("Type the number of zeroes: ")
	go func() {
		if err := nether.StartElection(numberOfLeaders, numberOfZeroes); err != nil {
			fmt.Println(err)
		}
	}()
}

func showConnections() {
	go func() {
		if err := nether.ShowConnections(); err != nil {
			fmt.Println(err)
		}
	}()
}

func downloadBlockchain() {
	go nether.RequestBlockchain()
}

func startEndpoint() {
	go nether.InitServer()
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	// nether.StartLog()
	nether.Start()

	fmt.Println("Welcome to nether blockchain - type your command:")

	for {
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch input {
		case "exit":
			return
		case "help":
			help()
		case "clear":
			clear()
		case "register":
			register()
		case "login":
			load()
		case "auto login":
			autoLogin()
		case "test userdata":
			testUserdata()
		case "show userdata":
			showUserdata()
		case "new blockchain":
			nether.NewBlockchain()
		case "load blockchain":
			nether.LoadBlockchain()
		case "show blockchain":
			nether.PrintBlockchain()
		case "start server":
			startServer()
		case "start client":
			startClient()
		case "ping all":
			pingAll()
		case "start election":
			startElection()
		case "show connections":
			showConnections()
		case "download blockchain":
			downloadBlockchain()
		case "start endpoint":
			startEndpoint()
		default:
			fmt.Println("No command found")
		}
	}
}
