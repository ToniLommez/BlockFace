package main

import (
	"bufio"
	"fmt"
	"nether/nether"
	"os"
	"strings"
)

func register() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Type your password: ")
	password, _ := reader.ReadString('\n')

	nether.Register(password)
}

func testMetadata() {
	fmt.Printf("%v\n\n", nether.GetMetadata().String())

	fmt.Println("Resetting metadata")
	nether.ResetMetadata()
	fmt.Printf("%v\n\n", nether.GetMetadata().String())

	fmt.Println("Reopening ...")
	load()
	fmt.Printf("%v\n\n", nether.GetMetadata().String())
}

func load() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Please type your password: ")
	password, _ := reader.ReadString('\n')

	if success := nether.LoadConfig(password); success {
		fmt.Println("Successfully logged")
	} else {
		fmt.Println("Wrong password")
	}

}

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Welcome to nether blockchain - type your command:")

	for {
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch input {
		case "exit":
			return
		case "register":
			register()
		case "load data":
			load()
		case "test metadata":
			testMetadata()
		default:
			fmt.Println("No command found")
		}
	}
}
