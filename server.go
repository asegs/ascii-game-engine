package main

import (
	"hash/fnv"
	"net"
	"strconv"
)

type Server struct {
	Players map[int] * net.UDPConn
	ConnectKey string
	ZoneIds map[int] int
	ZoneHandlers map[int] * ZoneHandlers

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
		ZoneIds: make(map[int]int),
		ZoneHandlers: make(map[int]*ZoneHandlers),
	}
}

func newServer (connectKey string) * Server {
	return &Server{
		Players:        make(map[int] * net.UDPConn,0),
		ConnectKey:     connectKey,
		ZoneIds: make(map[int]int),
		ZoneHandlers: make(map[int]*ZoneHandlers),
	}
}

func (s * Server) newZoneHandlers (zoneId int) * ZoneHandlers {
	handlers := &ZoneHandlers{
		Server:         s,
		PlayerHandlers: make(map[byte]func(int)),
	}
	s.ZoneHandlers[zoneId] = handlers
	return handlers
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

func (s * Server) broadcastToAll (message [] byte) {
	for _,player := range s.Players {
		player.Write(message)
	}
}

func (s * Server) broadcastCustomPair (key string, data interface{}, from int) {
	s.broadcastToAll(newStateUpdate(from).appendCustom(data,key).toBytes())
}

func (s * Server) broadcastStateUpdate (state interface{}, from int, keys ...string) {
	s.broadcastToAll(newStateUpdate(from).append(state,keys...).toBytes())
}
