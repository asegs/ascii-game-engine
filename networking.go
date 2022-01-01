package main

import (
	"fmt"
	"net"
)

type Network struct {
	Outbound chan byte
	Connections [] * net.UDPConn
	Port int
	Input * NetworkedStdIn
	Server * net.UDPConn
}

func initNetwork (port int,input * NetworkedStdIn) (* Network,error) {
	ServerConn, err := net.ListenUDP("udp",&net.UDPAddr{IP:[]byte{0,0,0,0},Port:port,Zone:""})
	if err != nil {
		return nil,err
	}
	network := &Network{
		Outbound:    make(chan byte, 1000),
		Connections: make([] * net.UDPConn,0),
		Port: port,
		Input: input,
		Server: ServerConn,
	}
	go network.sendToConnections()
	go network.readUDPConn()
	return network,nil
}

func (n * Network) addConnection (IP [] byte) error {
	Conn, err := net.DialUDP("udp",nil,&net.UDPAddr{
		IP:   IP,
		Port: n.Port,
		Zone: "",
	})
	if err != nil {
		return err
	}
	n.Connections = append(n.Connections,Conn)
	return err
}

func (n * Network) sendToConnections () {
	var message byte
	fmtMessage := make([]byte,1)
	for true {
		message = <- n.Outbound
		fmtMessage[0] = message
		for _,conn := range n.Connections {
			//returns number of chars sent, err
			_,err := conn.Write(fmtMessage)
			if err != nil {
				fmt.Println(err.Error())
			}
		}
	}
}

//only using single char messages
func (n * Network) readUDPConn () {
	//var s int
	var addr * net.UDPAddr
	var err error
	buf := make([]byte,16)
	for true {
		_,addr,err = n.Server.ReadFromUDP(buf)
		if err != nil {
			//log error here
			continue
		}
		n.Input.events <- &NetworkedMsg{
			Msg: buf[0],
			From: addr.Port,
		}
	}
}

func (n * Network) broadcast (char byte) {
	n.Outbound <- char
}