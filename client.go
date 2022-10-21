package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"reflect"
	"strconv"
	"time"
)

var DUMP_ID = -2
var GLOBAL_ID = -1
var LOCAL_ID = 0

type Client struct {
	LocalState            interface{}
	PlayerStates          map[int]interface{}
	GlobalState           interface{}
	LocalProcessor        map[string]func(interface{})
	GlobalProcessor       map[string]func(interface{})
	PlayersProcessor      map[string]func(int, interface{})
	CustomProcessor       map[string]func(string)
	Socket                *net.UDPConn
	EventChannel          *chan byte
	LastMessageProcessed  int
	Buffers               chan []byte
	StoredBuffers         map[int]*UpdateMessage
	HighestReceivedBuffer int
	BufferFirstQueryTimes map[int]time.Time
	OnNewPlayerConnect    func(id int)
	OnPlayerDisconnect    func(id int)
	Config                *ClientNetworkConfig
}

type StatePair struct {
	Key  string
	Json string
}

type UpdateMessage struct {
	From    int
	Pairs   []StatePair
	Id      int
	AsyncOk bool
}

type ClientNetworkConfig struct {
	ClientPort       int
	ServerPort       int
	BufferSize       int
	SkipWindowMs     int
	ScanMissedFreqMs int
	PacketRetries    int
}

type IndexUpdate struct {
	Index int
}

type DisconnectUpdate struct {
	DisconnectId int
}

func newClient(serverIp []byte, events *chan byte, localState interface{}, playerStates map[int]interface{}, globalState interface{}, onNewPlayer func(id int), onPlayerDisconnect func(id int), config *ClientNetworkConfig) *Client {
	client := &Client{
		LocalState:            localState,
		PlayerStates:          playerStates,
		GlobalState:           globalState,
		LocalProcessor:        make(map[string]func(interface{})),
		GlobalProcessor:       make(map[string]func(interface{})),
		PlayersProcessor:      make(map[string]func(int, interface{})),
		CustomProcessor:       make(map[string]func(string)),
		LastMessageProcessed:  -1,
		Buffers:               make(chan []byte),
		StoredBuffers:         make(map[int]*UpdateMessage),
		HighestReceivedBuffer: 0,
		BufferFirstQueryTimes: make(map[int]time.Time),
		OnNewPlayerConnect:    onNewPlayer,
		OnPlayerDisconnect:    onPlayerDisconnect,
		Config:                config,
		EventChannel:          events,
	}
	client.addCustomHandler("Index", func(s string) {
		var idx IndexUpdate
		_ = json.Unmarshal([]byte(s), &idx)
		if client.LastMessageProcessed < idx.Index {
			client.LastMessageProcessed = idx.Index
		}

		if client.HighestReceivedBuffer < idx.Index {
			client.HighestReceivedBuffer = idx.Index
		}
	})

	client.addCustomHandler("DisconnectId", func(s string) {
		var disconnect DisconnectUpdate
		err := json.Unmarshal([]byte(s), &disconnect)
		if err != nil {
			LogString("Couldn't process disconnect: " + err.Error())
			return
		}
		client.OnPlayerDisconnect(disconnect.DisconnectId)
	})
	err := client.connectToServer(serverIp)
	if err != nil {
		fmt.Println("Failed to connect to server " + err.Error())
	} else {
		helloBuffer := make([]byte, 1)
		helloBuffer[0] = CONNECT
		err = client.sendWithRetry(helloBuffer)
		if err != nil {
			LogString("Server not accepting connections: " + err.Error())
		}
		client.broadcastActions()
	}
	return client
}

func (c *Client) addLocalHandler(key string, operator func(interface{})) *Client {
	c.LocalProcessor[key] = operator
	return c
}

func (c *Client) addGlobalHandler(key string, operator func(interface{})) *Client {
	c.GlobalProcessor[key] = operator
	return c
}

func (c *Client) addPlayersHandler(key string, operator func(int, interface{})) *Client {
	c.PlayersProcessor[key] = operator
	return c
}

func (c *Client) addCustomHandler(key string, operator func(string)) *Client {
	c.CustomProcessor[key] = operator
	return c
}

func marshal(anything interface{}) []byte {
	output, err := json.Marshal(anything)
	if err != nil {
		LogString(err.Error())
	}
	return output
}

func newStateUpdate(from int, asyncOk bool) *UpdateMessage {
	return &UpdateMessage{
		From:    from,
		Pairs:   make([]StatePair, 0),
		AsyncOk: asyncOk,
	}
}

func wrapWithKey(key string, jsonBody string) string {
	return `{"` + key + `":` + jsonBody + "}"
}

func (u *UpdateMessage) append(state interface{}, keys ...string) *UpdateMessage {
	if len(keys) == 0 {
		valueOf := directIfPointer(state).Type()
		for i := 0; i < valueOf.NumField(); i++ {
			keys = append(keys, valueOf.Field(i).Name)
		}
	}
	for _, key := range keys {
		reflectedState := directIfPointer(state)
		u.Pairs = append(u.Pairs, StatePair{
			Key:  key,
			Json: wrapWithKey(key, string(marshal(reflectedState.FieldByName(key).Interface()))),
		})
	}
	return u
}

func (u *UpdateMessage) appendCustom(data interface{}, key string) *UpdateMessage {
	u.Pairs = append(u.Pairs, StatePair{
		Key:  key,
		Json: wrapWithKey(key, string(marshal(data))),
	})
	return u
}

func (u *UpdateMessage) toBytes() []byte {
	output, _ := json.Marshal(u)
	return output
}

func messageFromBytes(bytes []byte) *UpdateMessage {
	var update UpdateMessage
	err := json.Unmarshal(bytes, &update)
	if err != nil {
		LogString("JSON received: " + string(bytes))
		LogString(err.Error())
	}
	return &update

}

func updateStateFromJson(state interface{}, data string) error {
	err := json.Unmarshal([]byte(data), &state)
	return err
}

func keyInState(key string, state interface{}) bool {
	reflectedState := directIfPointer(state)
	return reflectedState.FieldByName(key).IsValid()
}

func (p StatePair) performCustomFunction(customs map[string]func(string)) {
	customs[p.Key](p.Json)
}

func (u *UpdateMessage) applyToStates(localState interface{}, playerStates map[int]interface{}, globalState interface{}, localHandlers map[string]func(interface{}), playersHandlers map[string]func(int, interface{}), globalHandlers map[string]func(interface{}), customHandlers map[string]func(string2 string)) {
	for _, pair := range u.Pairs {
		switch u.From {
		case LOCAL_ID:
			if keyInState(pair.Key, localState) {
				previousLocalState := directAndCopy(localState)
				updateStateFromJson(&localState, pair.Json)
				if localHandler, ok := localHandlers[pair.Key]; ok {
					localHandler(previousLocalState)
				} else {
					LogString("No local handler defined for key: " + pair.Key)
				}
			} else {
				pair.performCustomFunction(customHandlers)
			}
			break
		case GLOBAL_ID:
			if keyInState(pair.Key, globalState) {
				previousGlobalState := directAndCopy(globalState)
				updateStateFromJson(&globalState, pair.Json)
				if globalHandler, ok := globalHandlers[pair.Key]; ok {
					globalHandler(previousGlobalState)
				} else {
					LogString("No global handler defined for key: " + pair.Key)
				}
			} else {
				pair.performCustomFunction(customHandlers)
			}
			break
		default:
			playerState := playerStates[u.From]
			if keyInState(pair.Key, playerState) {
				previousPlayerState := directAndCopy(playerState)
				updateStateFromJson(&playerState, pair.Json)
				if playerHandler, ok := playersHandlers[pair.Key]; ok {
					playerHandler(u.From, previousPlayerState)
				} else {
					LogString("No global handler defined for key: " + pair.Key)
				}
			} else {
				pair.performCustomFunction(customHandlers)
			}
		}
	}
}

func (c *Client) connectToServer(IP []byte) error {
	Conn, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   IP,
		Port: c.Config.ServerPort,
		Zone: "",
	})
	if err != nil {
		return err
	}
	c.Socket = Conn
	return err
}

func (c *Client) broadcastActions() {
	go func() {
		var message byte
		var err error
		fmtMessage := make([]byte, 1)
		for true {
			message = <-*c.EventChannel
			fmtMessage[0] = message
			err = c.sendWithRetry(fmtMessage)
			if err != nil {
				LogString("Failure to broadcast action: " + err.Error())
			}
		}
	}()
}

func (c *Client) listen() {
	//var addr * net.UDPAddr
	var err error
	var received int
	buf := make([]byte, c.Config.BufferSize)
	go func() {
		for true {
			received, _, err = c.Socket.ReadFromUDP(buf)
			if err != nil {
				LogString("Failed to read from server: " + err.Error())
				continue
			}
			newCopy := copyByteSlice(buf, received)
			c.Buffers <- newCopy
		}
	}()

	var newBuf []byte
	var message *UpdateMessage

	go func() {
		var cont bool
		for true {
			newBuf = <-c.Buffers
			message = messageFromBytes(newBuf)
			if message.Pairs[0].Key == "DisconnectId" {
				LogString("Received disconnect packet")
			}

			if message.Id == DUMP_ID {
				c.applyMessage(message)
				continue
			}

			if c.HighestReceivedBuffer < message.Id {
				c.HighestReceivedBuffer = message.Id
			}
			c.StoredBuffers[message.Id] = message
			for i := c.LastMessageProcessed + 1; i <= c.HighestReceivedBuffer; i++ {
				cont = c.processBuffer(i)
				if !cont {
					break
				}
			}
		}
	}()

	go c.grabExtra()
}

func (c *Client) processBuffer(i int) bool {
	message, ok := c.StoredBuffers[i]
	//message of that id exists and has been stored
	if ok {
		c.applyBufferMessage(message, i)
		c.LastMessageProcessed = i
		return true
	} else {
		//message of that id has not been received
		readTime, ok := c.BufferFirstQueryTimes[i]
		//this message has already been looked for
		if ok {
			if time.Now().Sub(readTime) > time.Millisecond*time.Duration(c.Config.SkipWindowMs) {
				//skip, timed out
				c.LastMessageProcessed = i
				delete(c.BufferFirstQueryTimes, i)
				return true
			} else {
				//still waiting for this packet
				return false
			}
		} else {
			err := c.requestPacketFromServer(i)
			if err != nil {
				LogString("Failed to send packet: " + err.Error())
			}
			c.BufferFirstQueryTimes[i] = time.Now()
			return false
		}
	}
}

func (c *Client) applyBufferMessage(message *UpdateMessage, i int) {
	c.applyMessage(message)
	delete(c.StoredBuffers, i)
	delete(c.BufferFirstQueryTimes, i)
}

func (c *Client) applyMessage(message *UpdateMessage) {
	_, playerExists := c.PlayerStates[message.From]
	if message.From != GLOBAL_ID && message.From != LOCAL_ID && !playerExists {
		c.OnNewPlayerConnect(message.From)
	}
	message.applyToStates(c.LocalState,
		c.PlayerStates,
		c.GlobalState,
		c.LocalProcessor,
		c.PlayersProcessor,
		c.GlobalProcessor,
		c.CustomProcessor,
	)
}

func (c *Client) grabExtra() {
	for true {
		for i, update := range c.StoredBuffers {
			if i < c.LastMessageProcessed && update.AsyncOk {
				c.applyBufferMessage(update, i)
			}
		}
		time.Sleep(time.Duration(c.Config.ScanMissedFreqMs) * time.Millisecond)
	}
}

func (c *Client) requestPacketFromServer(id int) error {
	idText := []byte(strconv.Itoa(id))
	if len(idText) == 1 {
		newIdText := make([]byte, 2)
		newIdText[0] = '0'
		newIdText[1] = idText[0]
		idText = newIdText
	}
	return c.sendWithRetry(idText)
}

func (c *Client) sendWithRetry(buf []byte) error {
	return c.sendWithCustomRetry(buf, c.Config.PacketRetries)
}

func (c *Client) sendWithCustomRetry(buf []byte, maxRetries int) error {
	s, err := c.Socket.Write(buf)
	if err == nil && s < len(buf) {
		err = errors.New("entire body not sent")
	}
	if err != nil {
		for retryCount := 0; retryCount < maxRetries && err != nil; retryCount++ {
			s, err = c.Socket.Write(buf)
			if err == nil && s < len(buf) {
				err = errors.New("entire body not sent")
			}
			if err == nil {
				return err
			}
		}
	}
	return err
}

func directAndCopy(data interface{}) interface{} {
	reflectedValue := directIfPointer(data)
	directedInterface := reflectedValue.Interface()
	toJson, _ := json.Marshal(directedInterface)
	newOfType := reflect.New(reflectedValue.Type()).Interface()
	_ = json.Unmarshal(toJson, &newOfType)

	return newOfType
}

func copyByteSlice(bytes []byte, until int) []byte {
	if len(bytes) < until {
		until = len(bytes)
	}
	newSlice := make([]byte, until)
	for i := 0; i < until; i++ {
		newSlice[i] = bytes[i]
	}
	return newSlice
}
