package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// adresse d'un noeud du système
type adresse struct {
	ip   string
	port int
}

func newAdresse(ip string, port int) *adresse {
	return &adresse{
		ip:   ip,
		port: port,
	}
}

// Représente un noeud du système
// Une machine de l'anneau
type noeud struct {
	// numero du noeud dans l'anneau
	// utiliser pour déterminer la priorité
	moi int
	// numéro d'ordre dans l'anneau
	numeroOrdre int

	ad       adresse
	candidat bool

	// liste des adresses de tous les noeuds de l'anneau
	listeNoeud []adresse
}

// Récupérer la liste des noeuds depuis le fichier
// Exécutée uniquement à la création d'un noeud
func (n *noeud) init() {

	fichier, err := os.Open("ip.txt")
	if err != nil {
		log.Fatal(err)
	}

	fileScanner := bufio.NewScanner(fichier)

	fileScanner.Split(bufio.ScanLines)
	for fileScanner.Scan() {
		ligne := fileScanner.Text()
		tokens := strings.Split(ligne, " ")
		port, err := strconv.Atoi(tokens[1])
		if err != nil {
			continue
		}
		n.listeNoeud = append(n.listeNoeud, adresse{tokens[0], port})
	}

}

// Lancer le processus d'élection
func (n *noeud) election(num ...int) {

	var connection net.Conn
	var err error

	// Connection au prochain noeud de l'anneau
	// Envoyer le message au prochain dans l'anneau ie noeud avec numeroOrdre + 1
	if n.numeroOrdre >= len(n.listeNoeud) {
		connection, err = net.Dial("tcp", fmt.Sprintf("%s:%d", n.listeNoeud[0].ip, n.listeNoeud[0].port))
	} else {
		connection, err = net.Dial("tcp", fmt.Sprintf("%s:%d", n.listeNoeud[n.numeroOrdre].ip, n.listeNoeud[n.numeroOrdre].port))
	}
	defer connection.Close()

	if err != nil {
		panic(err)
	}

	// message de la forme ELECTION 4395435
	if len(num) == 0 {
		message := fmt.Sprintf("ELECTION %d", n.moi)
		fmt.Printf("J'envoi le message : %s\n", message)
		io.WriteString(connection, message)
	} else {
		message := fmt.Sprintf("ELECTION %d", num[0])
		fmt.Printf("J'envoi le message : %s\n", message)
		io.WriteString(connection, message)
	}

}

// Signaler à tous les autres noeuds, l'arrivée de cet noeud
// ipaddr : addresse ip du noeud
func (n *noeud) broadcast(ipaddr string) {
	for _, ad := range n.listeNoeud {
		if ad.port != n.ad.port {
			connection, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ad.ip, ad.port))
			defer connection.Close()

			if err != nil {
				log.Fatal(err)
			}

			// message de la forme INFO 127.0.0.1:8080
			message := fmt.Sprintf("INFO %s", ipaddr)
			io.WriteString(connection, message)
		}
	}
}

// Signaler à tous qui est l'elu
func (n *noeud) elu() {
	for _, ad := range n.listeNoeud {
		if ad.port != n.ad.port {
			connection, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ad.ip, ad.port))
			defer connection.Close()

			if err != nil {
				log.Fatal(err)
			}

			message := fmt.Sprintf("ELU %d:%d", n.numeroOrdre, n.moi)
			io.WriteString(connection, message)
		}
	}
}

// Création d'un noeud
func newNoeud(num int, ip string, port int) *noeud {
	rand.Seed(time.Now().UnixNano())
	n := &noeud{
		moi:         rand.Int(),
		numeroOrdre: num,
		ad:          *newAdresse(ip, port),
		candidat:    false,
		listeNoeud:  []adresse{},
	}
	n.init()
	return n
}

// recéption d'un message par le noeud
func (n *noeud) reception() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", n.ad.port))
	if err != nil {
		panic(err)
	}

	defer listener.Close()

	for {
		connection, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
		}

		scanner := bufio.NewScanner(connection)

		for scanner.Scan() {
			recu := scanner.Text()
			fmt.Printf("J'ai reçu le message suivant %s\n", recu)
			ligne := strings.Split(recu, " ")
			commande := strings.Trim(ligne[0], " \r\n")
			message := strings.Trim(ligne[1], " \r\n")

			switch commande {
			case "ELECTION":
				num, err := strconv.Atoi(message)
				if err != nil {
					panic("L'élection se passe en comparant des nombres, j'ai reçu un autre type")
				}
				if num > n.moi {
					// le numéro du site qui envoie est supérieur
					n.candidat = true
					connection.Close()
					n.election(num)
				} else if num < n.moi && !n.candidat {
					// le numéro du site qui envoie est inférieur
					n.candidat = true
					connection.Close()
					n.election()
				} else {
					connection.Close()
					n.elu()
				}
				break
			case "INFO":
				break

			case "ELU":
				fmt.Printf("L'élu c'est le noeud %s il a pour priorité %s\n", strings.Split(message, ":")[0], strings.Split(message, ":")[1])
				break

			default:
				fmt.Println("Je ne reconnais pas ce message")
			}

		}

		connection.Close()
	}
}
