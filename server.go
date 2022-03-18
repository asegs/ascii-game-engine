package main

import (
	"hash/fnv"
	"net"
	"strconv"
)

//Permute IP + Local Port into ID.  Receive byte + this id, have handler for byte.
type Server struct {
	PlayerHandlers map[byte]func(int)
	Players map[int] * net.UDPConn
	ConnectKey string
}

func newServerDefault () * Server {
	return &Server{
		PlayerHandlers: make(map[byte]func(int)),
		Players:        make(map[int] * net.UDPConn,0),
		ConnectKey:     "connect",
	}
}

func newServer (connectKey string) * Server {
	return &Server{
		PlayerHandlers: make(map[byte]func(int)),
		Players:        make(map[int] * net.UDPConn,0),
		ConnectKey:     connectKey,
	}
}

func (s * Server) addPlayerHandler (key byte,operator func(int)) * Server{
	s.PlayerHandlers[key] = operator
	return s
}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func permuteIp (addr * net.UDPAddr) int{
	return int(hash(addr.IP.String()+strconv.Itoa(addr.Port)))
}

func (s * Server) performHandler (addr * net.UDPAddr, msg byte) {
	s.PlayerHandlers[msg](permuteIp(addr))
}

func (s * Server) connect(addr * net.UDPConn) {

}

func (s * Server) broadcastCustomPair (key string, data interface{}, from int, toUser int) {
	update := newStateUpdate(from).appendCustom(data,key)
	s.Players[toUser].Write(update.toBytes())
}
