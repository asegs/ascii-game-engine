package main

import (
	"fmt"
	"hash/fnv"
	"net"
	"strconv"
)

var symbolicMapping map[byte]string = map[byte]string{
MOVE_LEFT: "<-",
MOVE_UP: "^",
MOVE_RIGHT: "->",
MOVE_DOWN: "v",
}

func symbolicMap (buf []byte) string {
	if str,ok := symbolicMapping[buf[0]]; ok {
		return str
	}
	return string(buf)
}


type ServerNetworkConfig struct {
	ClientPort int
	ServerPort int
	Strikes int
	BufferSize int
	StoredUpdates int
}

type Server struct {
	Players map[int] * net.UDPConn
	ConnectKey string
	ZoneIndexes map[int] int
	ZoneHandlers [] * ZoneHandlers
	PlayerJoined func(int)
	PlayerLeft func(int)
	Strikes map[int] int
	MessagesSent int
	MessageMap map[int] [] byte
	ZoneMap map[string] int
	PlayerState map[int] interface{}
	GlobalState interface{}
	Config * ServerNetworkConfig
}

type ZoneHandlers struct {
	Server * Server
	PlayerHandlers map[byte]func(int)
}

func newServerDefault (PlayerJoined func(int),PlayerLeft func(int),config * ServerNetworkConfig, globalState interface{}, playerStates map[int] interface{}) * Server {
	return newServer("connect",PlayerJoined,PlayerLeft,config,globalState,playerStates)
}

func newServer (connectKey string,PlayerJoined func(int),PlayerLeft func(int),config * ServerNetworkConfig, globalState interface{}, playerStates map[int] interface{}) * Server {
	return &Server{
		Players:        make(map[int] * net.UDPConn,0),
		ConnectKey:     connectKey,
		ZoneIndexes: make(map[int]int),
		ZoneHandlers: make([] *ZoneHandlers,0),
		Strikes: make(map[int] int),
		PlayerJoined: PlayerJoined,
		PlayerLeft: PlayerLeft,
		MessagesSent: 0,
		MessageMap: make(map[int] []byte),
		ZoneMap: make(map[string] int),
		PlayerState: playerStates,
		GlobalState: globalState,
		Config: config,
	}
}

func (s * Server) start() {
	err := s.listen()
	if err != nil {
		fmt.Println(err.Error())
	}
}

func (s * Server) newZoneHandlers (name string) * ZoneHandlers {
	handlers := &ZoneHandlers{
		Server:         s,
		PlayerHandlers: make(map[byte]func(int)),
	}
	s.ZoneMap[name] = len(s.ZoneHandlers)
	s.ZoneHandlers = append(s.ZoneHandlers, handlers)
	return handlers
}

func (s * Server) zoneByName (name string) * ZoneHandlers {
	return s.ZoneHandlers[s.ZoneMap[name]]
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
	_,_ = h.Write([]byte(s))
	return h.Sum32()
}

func permuteIp (addr * net.UDPAddr) int{
	return int(hash(addr.IP.String()))
}

func (s * Server) performHandler (addr * net.UDPAddr, msg byte) {
	id := permuteIp(addr)
	s.ZoneHandlers[s.ZoneIndexes[id]].PlayerHandlers[msg](id)
}

func (s * Server) broadcastToAll (stateUpdate * UpdateMessage) {
	stateUpdate.Id = s.MessagesSent
	message := stateUpdate.toBytes()
	s.MessageMap[s.MessagesSent] = message
	if s.Config.BufferSize < len(message) {
		LogString("Buffer limit exceeded with: " + string(message))
		message = message[0:s.Config.BufferSize]
	}
	for id,player := range s.Players {
		s.sendToConn(message,id,player)
	}
	s.MessagesSent++
	if _,ok := s.MessageMap[s.MessagesSent - s.Config.StoredUpdates]; ok {
		delete(s.MessageMap,s.MessagesSent - s.Config.StoredUpdates)
	}
}

func (s * Server) sendToConn(buf [] byte, id int, conn * net.UDPConn)  {
	n,err := conn.Write(buf)
	if err != nil {
		LogString("Failed to write: " + err.Error())
	}
	if err != nil || n < len(buf) {
		s.Strikes[id] ++
		if s.Strikes[id] > s.Config.Strikes {
			s.removePlayerSaveState(id)
		}
	}else {
		s.Strikes[id] = 0
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

func (s * Server) broadcastCustomPair (key string, data interface{}, from int,asyncOk bool) {
	s.broadcastToAll(newStateUpdate(from,asyncOk).appendCustom(data,key))
}

func (s * Server) broadcastStateUpdate (state interface{}, from int, asyncOk bool, keys ...string) {
	s.broadcastToAll(newStateUpdate(from,asyncOk).append(state,keys...))
}

func (s * Server) listen () error{
	ServerConn, err := net.ListenUDP("udp",&net.UDPAddr{
		IP:[]byte{0,0,0,0},
		Port:s.Config.ServerPort,
		Zone:"",
	})
	fmt.Println("Started UDP listen server")
	if err != nil {
		return err
	}
	var received int
	var addr * net.UDPAddr
	var id int
	buf := make([]byte,64)
	go func() {
		for true {
			received, addr, err = ServerConn.ReadFromUDP(buf)
			fmt.Printf("Received %s from %s\n",symbolicMap(buf),addr)
			if err != nil {
				LogString("Failed to read from connection: " + err.Error())
				continue
			}
			id = permuteIp(addr)
			if _, ok := s.Players[id]; !ok {
				NewConn, err := net.DialUDP("udp", nil, &net.UDPAddr{
					IP:   addr.IP,
					Port: s.Config.ClientPort,
					Zone: "",
				})
				if err != nil {
					LogString("Failed to add client to players set: " + err.Error())
				}
				s.PlayerJoined(id)
				s.addNewDefaultPlayer(id,NewConn)
			}
			if received > 1 {
				packetId,err := strconv.Atoi(string(buf[0:received]))
				if err != nil {
					LogString(fmt.Sprintf("Failed to convert packet info to id, info was: %s",buf[0:received]))
				}
				if message,ok := s.MessageMap[packetId] ; ok {
					s.sendToConn(message,id,s.Players[id])
				}
			}else {
				if operation,ok := s.ZoneHandlers[s.ZoneIndexes[id]].PlayerHandlers[buf[0]] ; ok {
					operation(id)
				}else {
					LogString("Zone has no implemented function for key: " + string(buf[0]))
				}
			}

		}
	}()
	return nil
}

func (s * Server) addNewDefaultPlayer (id int, conn * net.UDPConn) {
	s.Players[id] = conn
	s.ZoneIndexes[id] = 0
	s.Strikes[id] = 0
	s.dumpStateToPlayer(id)
}

func (s * Server) removePlayerSaveState (id int) {
	s.PlayerLeft(id)
	delete(s.Players,id)
	delete(s.ZoneIndexes,id)
	delete(s.Strikes,id)
}

func (s * Server) dumpStateToPlayer (id int) {
	for i, state := range s.PlayerState {
		playerUpdate := newStateUpdate(i,true).append(state)
		playerUpdate.Id = DUMP_ID
		s.sendToConn(playerUpdate.toBytes(),id,s.Players[id])
	}

	globalUpdate := newStateUpdate(GLOBAL_ID,true).append(s.GlobalState)
	globalUpdate.Id = DUMP_ID
	s.sendToConn(globalUpdate.toBytes(),id,s.Players[id])

	indexUpdate := newStateUpdate(GLOBAL_ID,true).appendCustom(s.MessagesSent,"Index")
	indexUpdate.Id = DUMP_ID
	s.sendToConn(indexUpdate.toBytes(),id,s.Players[id])
}