package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Entrez votre adresse ip")
	text, _ := reader.ReadString('\n')

	fmt.Println("Entrez votre port :")
	port, _ := reader.ReadString('\n')
	port = strings.Trim(port, " \r\n")
	portNum, err := strconv.Atoi(port)

	if err != nil {
		fmt.Println("Entrez un port correct")
		panic(err)
	}

	fmt.Println("Entrez votre numero")
	num, _ := reader.ReadString('\n')
	num = strings.Trim(num, " \r\n")
	numConv, err := strconv.Atoi(num)

	if err != nil {
		fmt.Println("Ceci n'est pas un nombre")
		panic(err)
	}

	n1 := newNoeud(numConv, text, portNum)
	go n1.broadcast(n1.ad.ip)
	n1.reception()

}
