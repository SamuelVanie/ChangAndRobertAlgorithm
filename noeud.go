package main

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
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
	suivant adresse
}

// Récupérer l'adresse du noeud suivant depuis le fichier
// Exécutée uniquement à la création d'un noeud
func (n *noeud) init() {

	fichier, err := os.Open("ip.txt")
	if err != nil {
		fmt.Print("Le fichier de configuration n'existe pas, vous devez definir un noeud suivant\n\n")
		fmt.Print("Veuillez entrer l'adresse ip et le port du suivant sous le format 0.0.0.0:0\n\n")
		reader := bufio.NewReader(os.Stdin)
		ligne, _ := reader.ReadString('\n')
		ligne = strings.Trim(ligne, " \r\n")
		adresseSuivant := strings.Split(ligne, ":")
		portSuivant, err := strconv.Atoi(adresseSuivant[1])
		if err != nil{
			fmt.Print("Veuillez entrer un port valide\n\n")
			os.Exit(1)
		}
		n.suivant = adresse{adresseSuivant[0], portSuivant}
		message := fmt.Sprintf("INFO %s:%d %s:%d", n.ad.ip, n.ad.port, n.suivant.ip, n.suivant.port)
		n.envoyerSuiv(message)
	}

	fileScanner := bufio.NewScanner(fichier)

	fileScanner.Split(bufio.ScanLines)
	if fileScanner.Scan() {
		ligne := fileScanner.Text()
		tokens := strings.Split(ligne, ":")
		port, err := strconv.Atoi(tokens[1])
		if err != nil {
			fmt.Printf("Le caractère entrée dans le fichier de configuration ne correspond pas à un port\n\n")
			os.Exit(1)
		}
		n.suivant = adresse{tokens[0], port}
	}


}

// Lancer le processus d'élection
func (n *noeud) election(num ...int) {

	// message de la forme ELECTION 4395435
	if len(num) == 0 {
		message := fmt.Sprintf("ELECTION %d", n.moi)
		fmt.Printf("Je suis candidat\nEnvoi d'un message pour mon élection : %s\n\n", message)
		n.envoyerSuiv(message)
	} else {
		message := fmt.Sprintf("ELECTION %d", num[0])
		fmt.Printf("J'ai reçu un poids supérieur au mien\nTransmission du message au suivant : %s\n\n", message)
		n.envoyerSuiv(message)
	}

}


// Envoyer un message à son suivant
// message : le message à transmettre
func (n *noeud) envoyerSuiv(message string){
	if n.suivant.ip == n.ad.ip && n.suivant.port == n.ad.port {
		fmt.Println("Plus aucun noeud dans l'anneau")
		os.Exit(1)
	}
	connection, err := net.Dial("tcp", fmt.Sprintf("%s:%d", n.suivant.ip, n.suivant.port))

	if err != nil{
		fmt.Printf("Impossible de communiquer avec le noeud suivant\n\n")
		os.Exit(1)
	}

	defer connection.Close()
	//envoi du message
	io.WriteString(connection, message)
}

// Signaler à tous qui est l'elu
func (n *noeud) elu() {
	fmt.Printf("Le noeud %d a été élu, il a pour poids %d\n\n", n.numeroOrdre, n.moi)
	message := fmt.Sprintf("ELU %d:%d", n.numeroOrdre, n.moi)
	n.envoyerSuiv(message)
}


// Création d'un noeud
func newNoeud(num int, ip string, port int) *noeud {
  rand.Seed(time.Now().UnixNano())
  n := &noeud{
    moi:         rand.Int(),
    numeroOrdre: num,
    ad:          *newAdresse(ip, port),
    candidat:    false,
    suivant:  adresse{},
  }

  n.init()

  return n
}


// recéption d'un message par le noeud
func (n *noeud) reception() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", n.ad.port))
	if err != nil {
		fmt.Println("Ecoute impossible sur ce port")
	  os.Exit(1)
	}

	defer listener.Close()
        // Quitter l'anneau à l'appui sur la commande Ctrl+C ou Ctrl+Z
        ch := make(chan os.Signal, 1)
        signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

        go func(){
          for sig := range ch{
            if sig == syscall.SIGINT || sig == syscall.SIGTERM{
              // Envoi du numeroOrdre du suivant
	      message := fmt.Sprintf("EXIT %s:%d %s:%d", n.ad.ip, n.ad.port, n.suivant.ip, n.suivant.port)	
              n.envoyerSuiv(message)
	      os.Exit(1)
            }
          }
        }()

        for {
          connection, err := listener.Accept()
          if err != nil {
                  fmt.Println(err)
          }

          scanner := bufio.NewScanner(connection)
          for scanner.Scan(){

            recu := scanner.Text()
            fmt.Printf("J'ai reçu le message suivant %s\n\n", recu)
            ligne := strings.Split(recu, " ")
            commande := strings.Trim(ligne[0], " \r\n")
            message := strings.Trim(ligne[1], " \r\n")

            switch commande {
                    case "ELECTION":
                            num, err := strconv.Atoi(message)
                            if err != nil {
                                    fmt.Println("L'élection se passe en comparant des nombres, j'ai reçu un autre type")
                                    os.Exit(1)
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
                            noeudArrivant := strings.Split(ligne[1], ":")
			    noeudSuivNoeudArrivant := strings.Split(ligne[2], ":") 
                            portArrivant, err := strconv.Atoi(noeudArrivant[1])
                            if err != nil {
                                    fmt.Println("Entrez plutôt un nombre comme port svp")
				    os.Exit(1)
                            }
                            portSuivArrivant, err := strconv.Atoi(noeudSuivNoeudArrivant[1])
			    if err != nil {
                                    fmt.Println("Entrez plutôt un nombre comme port svp")
				    os.Exit(1)
                            }
				// Le noeud suivant le noeud qui vient d'arriver est le suivant du recepteur
				// Le noeud qui vient d'arriver devient donc le suivant du recepteur
			    if noeudSuivNoeudArrivant[0] == n.suivant.ip && portSuivArrivant == n.suivant.port{
					n.suivant.ip = noeudArrivant[0]
					n.suivant.port = portArrivant
			    }else if noeudArrivant[0] == n.suivant.ip && portArrivant == n.suivant.port{

			    }else{
					// Le noeud qui vient d'arriver n'a pas pour suivant le suivant du recepteur 
					// transmettre le message
				n.envoyerSuiv(recu)
			    }
                            fmt.Printf("Ajout de %s:%s dans l'anneau\n\n", noeudArrivant[0], noeudArrivant[1])

                    case "EXIT":

			    noeudPartant := strings.Split(ligne[1], ":")
			    noeudSuivPartant := strings.Split(ligne[2], ":")
			    portNoeudPartant, err := strconv.Atoi(noeudPartant[1])
			    if err != nil{
				fmt.Println("Le port reçu est erroné")
				}
			    
			    portSuivPartant, err := strconv.Atoi(noeudSuivPartant[1])
		            if err != nil{
				fmt.Println("Le port reçu est erroné")
			    }
			    // Si le noeud qui part est le suivant du récepteur, alors le suivant de celui-ci remplace celui du recepteur
			    // 1->2->3->1  avec 3 qui part devient 1->2->1
			    // Si non alors envoyer au suivant
			    if noeudPartant[0] == n.suivant.ip && portNoeudPartant == n.suivant.port{
				n.suivant.ip = noeudSuivPartant[0]
				n.suivant.port = portSuivPartant
				n.election()
			    }else{
				n.envoyerSuiv(recu)
			    }

                            connection.Close()

                    case "ELU":
			    elu := strings.Split(message, ":")
			    fmt.Printf("Le noeud %s a été élu, il a pour poids %s\n\n", elu[0], elu[1])
			    ordre, err := strconv.Atoi(elu[0])
			    if err != nil{
				fmt.Println("Erreur de numéro d'ordre")
				break
			    }
			    moi, err := strconv.Atoi(elu[1])
		            if err != nil{
				fmt.Println("Erreur de numéro d'ordre")
				break
			    }
			    if ordre != n.numeroOrdre && moi != n.moi{
					n.envoyerSuiv(recu)
			    }	
                            break

                    default:
                            fmt.Println("Je ne reconnais pas ce message")
                
              }
            }

          defer connection.Close()
        }

}
