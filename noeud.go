package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

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

	ad       adresse
	candidat chan bool

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

			message := fmt.Sprintf("INFO %s", ipaddr)
			io.WriteString(connection, message)
		}
	}
}

// Création d'un noeud
func newNoeud(moi int, ip string, port int) *noeud {
	n := &noeud{
		moi:        moi,
		ad:         *newAdresse(ip, port),
		candidat:   make(chan bool),
		listeNoeud: []adresse{},
	}
	n.init()
	return n
}

//traitement d'un message par le noeud
//connection : la connection entrante
func traitement(connection net.Conn) {
	scanner := bufio.NewScanner(connection)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
	connection.Close()
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

		go traitement(connection)
	}
}
