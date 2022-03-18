package main

import (
	"hash/fnv"
	"net"
	"strconv"
)

type Server struct {
	Players map[int] * net.UDPConn
	ConnectKey string
}

//Permute IP + Local Port into ID.  Receive byte + this id, have handler for byte.
type ZoneHandlers struct {
	Server * Server
	PlayerHandlers map[byte]func(int)
}

func newServerDefault () * Server {
	return &Server{
		Players:        make(map[int] * net.UDPConn,0),
		ConnectKey:     "connect",
	}
}

func newServer (connectKey string) * Server {
	return &Server{
		Players:        make(map[int] * net.UDPConn,0),
		ConnectKey:     connectKey,
	}
}

func (s * Server) newZoneHandlers () * ZoneHandlers {
	return &ZoneHandlers{
		Server:         s,
		PlayerHandlers: make(map[byte]func(int)),
	}
}

func (z * ZoneHandlers) addPlayerHandler (key byte,operator func(int)) * ZoneHandlers{
	z.PlayerHandlers[key] = operator
	return z
}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func permuteIp (addr * net.UDPAddr) int{
	return int(hash(addr.IP.String()+strconv.Itoa(addr.Port)))
}

func (z * ZoneHandlers) performHandler (addr * net.UDPAddr, msg byte) {
	z.PlayerHandlers[msg](permuteIp(addr))
}

func (s * Server) connect(addr * net.UDPConn) {

}

func (s * Server) broadcastCustomPair (key string, data interface{}, from int, toUser int) {
	update := newStateUpdate(from).appendCustom(data,key)
	s.Players[toUser].Write(update.toBytes())
}
