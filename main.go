package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

var wg sync.WaitGroup

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Entrez votre adresse ip")
	text, _ := reader.ReadString('\n')
	text = strings.Trim(text, " \r\n")

	fmt.Println("Entrez votre port :")
	port, _ := reader.ReadString('\n')
	port = strings.Trim(port, " \r\n")
	portNum, err := strconv.Atoi(port)

	if err != nil {
		fmt.Println("Entrez un port correct")
		panic(err)
	}

	fmt.Println("Entrez votre numero\n\rTips: Le numéro doit être supérieur ou égal à 1")
	num, _ := reader.ReadString('\n')
	num = strings.Trim(num, " \r\n")
	numConv, err := strconv.Atoi(num)

	if err != nil {
		fmt.Println("Ceci n'est pas un nombre")
		panic(err)
	}

	n1 := newNoeud(numConv, text, portNum)

	fmt.Printf("Bienvenue noeud %d\nAdresse %s:%d\nPoids %d Voulez-vous lancer l'élection? Oui(O) ou Non(N)\n", n1.numeroOrdre, n1.ad.ip, n1.ad.port, n1.moi)

	choix, _ := reader.ReadString('\n')
	choix = strings.Trim(choix, " \r\n")

	wg.Add(1)
	go n1.reception()

	switch choix {
	case "O":
		n1.election()
	default:
		fmt.Println("En attente de communication avec les autres noeuds")
	}

	wg.Wait()

}
