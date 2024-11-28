package main

import (
	"bufio"
	"fmt"
	"nether/nether"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

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

func seeUserdata() {
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

func startConnection() {
	go func() {
		if err := nether.StartConnections(); err != nil {
			fmt.Println(err)
		}
	}()
}

func connect() {
	if err := nether.Connect(input("Type the ipv6 server address: ")); err != nil {
		fmt.Println(err)
	}
}

func pingAll() {
	nether.PingAll()
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	// nether.StartLog()

	fmt.Println("Welcome to nether blockchain - type your command:")

	for {
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch input {
		case "exit":
			return
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
		case "see userdata":
			seeUserdata()
		case "new blockchain":
			nether.NewBlockchain()
		case "load blockchain":
			nether.LoadBlockchain()
		case "write random block":
			nether.WriteRandomBlock()
		case "print blockchain":
			nether.PrintBlockchain()
		case "start server":
			startConnection()
		case "start client":
			connect()
		case "ping all":
			pingAll()
		default:
			fmt.Println("No command found")
		}
	}
}
