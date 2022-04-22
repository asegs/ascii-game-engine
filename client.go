package main

import (
	"encoding/json"
	"fmt"
	"net"
	"reflect"
	"time"
)

var GLOBAL_ID int = -1
var LOCAL_ID int = 0

type Client struct {
	 LocalState interface{}
	 PlayerStates map[int]interface{}
	 GlobalState interface{}
	 LocalProcessor map[string]func()
	 GlobalProcessor map[string]func()
	 PlayersProcessor map[string]func(int)
	 CustomProcessor map[string]func(string)
	 ToSend * net.UDPConn
	 ToReceive * net.UDPConn
	 Input * NetworkedStdIn
	 LastMessageProcessed int
	 Buffers chan [] byte
	 StoredBuffers map [int] * UpdateMessage
	 HighestReceivedBuffer int
	 BufferFirstQueryTimes map [int] time.Time
}

type StatePair struct {
	Key string
	Json string
}


type UpdateMessage struct {
	From int
	Pairs [] StatePair
	Id int
}

type ClientNetworkConfig struct {
	defaultPort int
	bufferSize int
	skipWindowMs int
}

var clientNetworkConfig ClientNetworkConfig

func newClient (serverIp []byte,input * NetworkedStdIn,localState interface{},playerStates map[int]interface{},globalState interface{}) * Client {
	client := &Client{
		LocalState: localState,
		PlayerStates: playerStates,
		GlobalState: globalState,
		LocalProcessor:   make(map[string]func()),
		GlobalProcessor:  make(map[string]func()),
		PlayersProcessor: make(map[string]func(int)),
		CustomProcessor:  make(map[string]func(string)),
		LastMessageProcessed: 0,
		Buffers: make(chan [] byte),
		StoredBuffers: make(map[int] * UpdateMessage),
		HighestReceivedBuffer: 0,
		BufferFirstQueryTimes: make(map[int] time.Time),
	}
	client.Input = input
	err := client.connectToServer(serverIp)
	if err != nil {
		fmt.Println("Failed to connect to server " + err.Error())
	}else {
		client.broadcastActions()
	}
	return client
}

func (c * Client) addLocalHandler (key string,operator func()) * Client{
	c.LocalProcessor[key] = operator
	return c
}

func (c * Client) addGlobalHandler (key string,operator func()) * Client{
	c.GlobalProcessor[key] = operator
	return c
}

func (c * Client) addPlayersHandler (key string,operator func(int)) * Client{
	c.PlayersProcessor[key] = operator
	return c
}

func (c * Client) addCustomHandler (key string,operator func(string)) * Client{
	c.CustomProcessor[key] = operator
	return c
}

func marshal(anything interface{}) []byte {
	output,_ := json.Marshal(anything)
	return output
}

func newStateUpdate(id int) * UpdateMessage {
	return &UpdateMessage{
		From:    id,
		Pairs: make([] StatePair,0),
	}
}

func wrapWithKey (key string, jsonBody string) string {
	return `{"` + key + `":` + jsonBody + "}"
}

func (u * UpdateMessage) append(state interface{}, keys ...string) * UpdateMessage {
	for _,key := range keys {
		u.Pairs = append(u.Pairs,StatePair{
			Key:  key,
			Json: wrapWithKey(key,string(marshal(reflect.ValueOf(state).FieldByName(key).Interface()))),
		})
	}
	return u
}

func (u * UpdateMessage) appendCustom (data interface{}, key string) * UpdateMessage{
	u.Pairs = append(u.Pairs,StatePair{
		Key:  key,
		Json: wrapWithKey(key,string(marshal(data))),
	})
	return u
}

func (u * UpdateMessage) toBytes() []byte{
	output,_ := json.Marshal(u)
	return output
}

func messageFromBytes (bytes []byte) * UpdateMessage {
	var update UpdateMessage
	_ = json.Unmarshal(bytes,&update)
	return &update

}

func updateStateFromJson(state interface{},data string) error{
	err := json.Unmarshal([]byte(data),&state)
	return err
}

func keyInState (key string, state interface{}) bool{
	return reflect.ValueOf(state).FieldByName(key).IsValid()
}

func (p StatePair) performCustomFunction(customs map[string]func(string)) {
	customs[p.Key](p.Json)
}


func (u * UpdateMessage) applyToStates(localState interface{},playerStates map[int]interface{},globalState interface{},localHandlers map[string]func(),playersHandlers map[string]func(int),globalHandlers map[string]func(),customHandlers map[string]func(string2 string)){
	for _,pair := range u.Pairs {
		switch u.From {
		case LOCAL_ID:
			if keyInState(pair.Key,localState) {
				updateStateFromJson(&localState,pair.Json)
				localHandlers[pair.Key]()
			} else {
				pair.performCustomFunction(customHandlers)
			}
			break
		case GLOBAL_ID:
			if keyInState(pair.Key,globalState) {
				updateStateFromJson(&globalState,pair.Json)
				globalHandlers[pair.Key]()
			} else {
				pair.performCustomFunction(customHandlers)
			}
			break
		default:
			playerState := playerStates[u.From]
			if keyInState(pair.Key,playerState) {
				updateStateFromJson(&playerState,pair.Json)
				playersHandlers[pair.Key](u.From)
			} else {
				pair.performCustomFunction(customHandlers)
			}
		}
	}
}

func (c * Client) connectToServer(IP []byte) error{
	Conn, err := net.DialUDP("udp",nil,&net.UDPAddr{
		IP:   IP,
		Port: clientNetworkConfig.defaultPort,
		Zone: "",
	})
	ServerConn, err := net.ListenUDP("udp",&net.UDPAddr{
		IP:[]byte{0,0,0,0},
		Port:clientNetworkConfig.defaultPort,
		Zone:"",
	})
	if err != nil {
		return err
	}
	c.ToSend = Conn
	c.ToReceive = ServerConn
	return err
}

func (c * Client) broadcastActions () {
	go func() {
		var message * NetworkedMsg
		fmtMessage := make([]byte,1)
		for true {
			message = <- c.Input.events
			fmtMessage[0] = message.Msg
			_,err := c.ToSend.Write(fmtMessage)
			if err != nil {
				LogString("Failure to broadcast action: " + err.Error())
			}
		}
	}()
}

func (c * Client) listen() {
	var addr * net.UDPAddr
	var err error
	buf := make([]byte,serverNetworkConfig.bufferSize)
	go func() {
		for true {
			_, addr, err = c.ToReceive.ReadFromUDP(buf)
			if err != nil {
				LogString("Failed to read from server: " + err.Error())
				continue
			}
			//copy?
			newBuffer := processJsonFromBuffer(buf)
			c.Buffers <- newBuffer
		}
	}()

	var newBuf []byte
	var message * UpdateMessage
	var ok bool
	var readTime time.Time

	go func() {
		//Perform every n ms to scoop up old data via map keys when no new packets have come in or when expired, make threadsafe?
		for true {
			newBuf = <- c.Buffers
			message = messageFromBytes(newBuf)
			if c.HighestReceivedBuffer < message.Id {
				c.HighestReceivedBuffer = message.Id
			}
			c.StoredBuffers[message.Id] = message
			for i := c.LastMessageProcessed + 1 ; i < c.HighestReceivedBuffer ; i ++ {
				//map of time first checked
				message,ok = c.StoredBuffers[i]
				//message of that id exists and has been stored
				if ok {
					message.applyToStates(c.LocalState,
						c.PlayerStates,
						c.GlobalState,
						c.LocalProcessor,
						c.PlayersProcessor,
						c.GlobalProcessor,
						c.CustomProcessor,
						)
					delete(c.StoredBuffers,i)
					delete(c.BufferFirstQueryTimes,i)
					c.LastMessageProcessed = i
				}else {
					//message of that id has not been received
					readTime,ok = c.BufferFirstQueryTimes[i]
					//this message has already been looked for
					if ok {
						if time.Now().Sub(readTime) > time.Millisecond * time.Duration(clientNetworkConfig.skipWindowMs) {
							//skip, timed out
							c.LastMessageProcessed = i
							continue
						}else {
							//still waiting for this packet
							break
						}
					}else {
						//haven't checked for this packet yet, not here, will request and wait
						c.requestPacketFromServer(i)
						c.BufferFirstQueryTimes[i] = time.Now()
					}
				}
			}
			message.applyToStates(
					c.LocalState,
					c.PlayerStates,
					c.GlobalState,
					c.LocalProcessor,
					c.PlayersProcessor,
					c.GlobalProcessor,
					c.CustomProcessor,
				)
		}
	}()
}

/**
This may be a problem, it will process arrays like:
[{,",k,e,y,",:,1,2,3,}] fine, but will struggle with:
{"key":"hello\""}
{"key":"hello\"}"}
Very rare case but may be a pain.
 */
func processJsonFromBuffer (buf [] byte) []byte {
	bracketDepth := 0
	inQuotes := false
	var char byte
	for i := 0 ; i < len(buf) ; i ++ {
		char = buf[i]
		if bracketDepth == 0 && i > 0 {
			return buf[0:i]
		}
		if !inQuotes && char == '{' {
			bracketDepth++
		}else if !inQuotes && char == '}' {
			bracketDepth--
		} else if char == '"' {
			inQuotes = !inQuotes
		}
	}
	return buf
}

func (c * Client) requestPacketFromServer (id int) {

}