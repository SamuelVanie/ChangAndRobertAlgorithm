package main

import (
  "fmt"
  "net"
  "os"
  "bufio"
  "log"
  "strings"
)

var PORT int = 5999

// Représente un noeud du système
// Une machine de l'anneau
type noeud struct{
  // numero du noeud dans l'anneau
  // utiliser pour déterminer la priorité
  moi int

  // adresse ip du noeud
  ip string
  candidat chan bool

  // liste des adresses ip de tous les noeuds de l'anneau
  listeNoeud []string
}

// Récupérer la liste des noeuds depuis le fichier
// Exécutée uniquement à la création d'un noeud
func (n *noeud) init(){

  fichier, err := os.Open("ip.txt")
  if err != nil{
    log.Fatal(err)
  }

  fileScanner := bufio.NewScanner(fichier)

  fileScanner.Split(bufio.ScanLines)
  for fileScanner.Scan(){
    n.listeNoeud = append(n.listeNoeud, fileScanner.Text())
  }

}

// Signaler à tous les autres noeuds, l'arrivée de cet noeud
func (n *noeud) broadcast(ipaddr string){
  for _, ip := range n.listeNoeud{
    connection, err := net.Dial("tcp", fmt.Sprintf("%s:%v", ip, PORT))
    log.Fatal(err)

    defer connection.Close()

    if err != nil{
      log.Fatal(err)
    }
    
  }
}

func newNoeud(moi int, ip string) *noeud {
  n := &noeud{
    moi: moi,
    ip: ip,
    candidat: make(chan bool),
    listeNoeud: make([]string, 0),
  }
  n.init()
  n.broadcast(ip)
  return n
}


