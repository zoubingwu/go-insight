package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Welcome!")

	for {
		fmt.Print("> ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "exit" {
			break
		}

		response := getResponse(input)
		fmt.Println(response)
	}
}

func getResponse(input string) string {
	return "you just said：" + input
}
