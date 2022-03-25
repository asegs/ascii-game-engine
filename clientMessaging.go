package main

import (
	"encoding/json"
	"net"
	"reflect"
)

var GLOBAL_ID int = -1
var LOCAL_ID int = 0

type Client struct {
	 LocalProcessor map[string]func()
	 GlobalProcessor map[string]func()
	 PlayersProcessor map[string]func(int)
	 CustomProcessor map[string]func(string)
}

type StatePair struct {
	Key string
	Json string
}


type UpdateMessage struct {
	Id int
	Pairs [] StatePair
}

func newClient () * Client {
	return &Client{
		LocalProcessor:   make(map[string]func()),
		GlobalProcessor:  make(map[string]func()),
		PlayersProcessor: make(map[string]func(int)),
		CustomProcessor:  make(map[string]func(string)),
	}
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
		Id:    id,
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

func updateStateFromJson(state interface{},data string) {
	_ = json.Unmarshal([]byte(data),&state)
}

func keyInState (key string, state interface{}) bool{
	return reflect.ValueOf(state).FieldByName(key).IsValid()
}

func (p StatePair) performCustomFunction(customs map[string]func(string)) {
	customs[p.Key](p.Json)
}


func (u * UpdateMessage) applyToStates(localState interface{},playerStates map[int]interface{},globalState interface{},localHandlers map[string]func(),playersHandlers map[string]func(int),globalHandlers map[string]func(),customHandlers map[string]func(string2 string)){
	for _,pair := range u.Pairs {
		switch u.Id {
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
			playerState := playerStates[u.Id]
			if keyInState(pair.Key,playerState) {
				updateStateFromJson(&playerState,pair.Json)
				playersHandlers[pair.Key](u.Id)
			} else {
				pair.performCustomFunction(customHandlers)
			}
		}
	}
}

func connectToServer(IP []byte) error{
	Conn, err := net.DialUDP("udp",nil,&net.UDPAddr{
		IP:   IP,
		Port: n.Port,
		Zone: "",
	})
	if err != nil {
		return err
	}
}