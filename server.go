package main

import (
	"fmt"
	"hash/fnv"
	"net"
	"strconv"
)

var ServerNetworkConfig NetworkConfig

type Server struct {
	Players map[int] * net.UDPConn
	ConnectKey string
	ZoneIndexes map[int] int
	ZoneHandlers [] * ZoneHandlers
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
		ZoneIndexes: make(map[int]int),
		ZoneHandlers: make([] *ZoneHandlers,0),
	}
}

func newServer (connectKey string) * Server {
	return &Server{
		Players:        make(map[int] * net.UDPConn,0),
		ConnectKey:     connectKey,
		ZoneIndexes: make(map[int]int),
		ZoneHandlers: make([] *ZoneHandlers,0),
	}
}

func (s * Server) start() {
	err := s.listen()
	if err != nil {
		fmt.Println(err.Error())
	}
}

func (s * Server) newZoneHandlers () * ZoneHandlers {
	handlers := &ZoneHandlers{
		Server:         s,
		PlayerHandlers: make(map[byte]func(int)),
	}
	return handlers
}

func (s * Server) addZoneHandlers (zoneHandlers * ZoneHandlers) {
	s.ZoneHandlers = append(s.ZoneHandlers, zoneHandlers)
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

func (s * Server) performHandler (addr * net.UDPAddr, msg byte) {
	id := permuteIp(addr)
	s.ZoneHandlers[s.ZoneIndexes[id]].PlayerHandlers[msg](id)
}

func (s * Server) connect(addr * net.UDPConn) {

}

func (s * Server) broadcastToAll (message [] byte) {
	for _,player := range s.Players {
		player.Write(message)
	}
}

func (s * Server) nextZone (from int) {
	index := s.ZoneIndexes[from]
	if index == len(s.ZoneHandlers) - 1 {
		s.ZoneIndexes[from] = 0
	}else {
		s.ZoneIndexes[from] ++
	}
}

func (s * Server) broadcastCustomPair (key string, data interface{}, from int) {
	s.broadcastToAll(newStateUpdate(from).appendCustom(data,key).toBytes())
}

func (s * Server) broadcastStateUpdate (state interface{}, from int, keys ...string) {
	s.broadcastToAll(newStateUpdate(from).append(state,keys...).toBytes())
}

func (s * Server) listen () error{
	ServerConn, err := net.ListenUDP("udp",&net.UDPAddr{
		IP:[]byte{0,0,0,0},
		Port:ClientNetworkConfig.defaultPort,
		Zone:"",
	})
	if err != nil {
		return err
	}
	//var s int
	var addr * net.UDPAddr
	var id int
	buf := make([]byte,16)
	go func() {
		for true {
			_, addr, err = ServerConn.ReadFromUDP(buf)
			if err != nil {
				LogString("Failed to read from connection: " + err.Error())
				continue
			}
			id = permuteIp(addr)
			if _, ok := s.Players[id]; !ok {
				NewConn, err := net.DialUDP("udp", nil, &net.UDPAddr{
					IP:   addr.IP,
					Port: ServerNetworkConfig.defaultPort,
					Zone: "",
				})
				if err != nil {
					LogString("Failed to add client to players set: " + err.Error())
				}
				s.Players[id] = NewConn
			}
			s.ZoneHandlers[s.ZoneIndexes[id]].PlayerHandlers[buf[0]](id)
		}
	}()
	return nil
}
