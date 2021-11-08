package main

import (
	"fmt"
	"io/ioutil"
	"strings"
	"strconv"
)

const MAX_ITEMS = 25
const MAX_DIALOGUES = 25
const USER_DATA_SIZE = 3

var NPCMap map[int]NPC
var DialogueMap map[int]*dialogue

type NPC struct {
	name string
	male bool
	description string
	id int
	//Other attributes
}

type item struct{
	id int64
	name string
	value int64
	damage int32
	description string
	//NEED OTHERS
}

type event struct{
	code int32
	feedback string
	id int64
	parameter_changed string
	parameter_change_amount int64
	items_to_gain [MAX_ITEMS]item
	//Handle both item gains/losses and numerical changes
}

type dialogue struct{
	speaker NPC
	text string
	childrenCount int
	children [MAX_DIALOGUES]*dialogue
	order int
	id int
}


func getOrder(s string)(string,int){
	terms := strings.Split(s,"^")
	for i:=0;i<len(terms);i++{
		if len(terms[i])>1{
			return terms[i],i
		}
	}
	fmt.Println("No message on string: "+s)
	return "",-1
}

func initializeNPCS(){
	NPCMap = make(map[int]NPC)
	data,err := ioutil.ReadFile("speakers.speakers")
	if err!=nil{
		fmt.Println("Speakers file damaged or missing")
		return
	}
	speakersFile := string(data)
	people := strings.Split(speakersFile,"~~~")
	for i := 1;i<len(people);i+=2{
		NPCID,err := strconv.Atoi(people[i])
		if err!=nil{
			fmt.Println("Invalid ID "+ people[i])
		}
		NPCDetails := people[i+1]
		NPCDetailsSlice := strings.Split(NPCDetails,"\n")[1:]
		male,err := strconv.ParseBool(strings.TrimSpace(NPCDetailsSlice[1]))
		if err!=nil{
			fmt.Println("There are only two genders ;)")
		}
		npc1 := NPC{NPCDetailsSlice[0],male,NPCDetailsSlice[2],NPCID}
		NPCMap[NPCID] = npc1

	}
}

func insert(root *dialogue,key string,order int,speaker NPC,id int){
	var addy *dialogue
	addy = root
	for order>1{
		addy = addy.children[addy.childrenCount-1]
		order--
	}
	var emptyArr [MAX_DIALOGUES]*dialogue
	fresh := dialogue{speaker,key,0,emptyArr,order,id}
	addy.children[addy.childrenCount] = &fresh
	addy.childrenCount++
}


func initializeDialogue(){
	DialogueMap = make(map[int]*dialogue)
	data,err := ioutil.ReadFile("dialogue.dialogue")
	if err!=nil{
		fmt.Println("Dialogue file damaged or missing")
		return
	}
	dialogueFile := string(data)
	sections := strings.Split(dialogueFile,"~~~")
	for i := 1;i<len(sections);i+=2{
		//header := strings.Split(sections[i],"<i>")
		//section_name := header[0]
		//section_id,err := strconv.Atoi(string(header[1]))
		//if err!=nil{
			//fmt.Println("Invalid ID "+ header[1])
		//}
		body := sections[i+1]
		lines := strings.Split(body,"\n")
		//to handle forced newline, check for newline in future
		for b := 1;b<len(lines)-1;b++{
			line := lines[b]
			broken_line := strings.Split(line,"<s>")
			message := broken_line[0]
			speaker_id,err := strconv.Atoi(broken_line[1])
			if err!=nil{
				fmt.Println("The message speaker id was improperly configured "+broken_line[1])
			}
			npc_unit := NPCMap[speaker_id]
			var children [MAX_DIALOGUES]*dialogue
			cleanMessage,order := getOrder(message)
			if b==1{
				DialogueMap[i] = &dialogue{npc_unit,cleanMessage,0,children,order,b}
				continue
			}
			insert(DialogueMap[i],cleanMessage,order,npc_unit,b)
		}
	}
}


func converse(index int){
	var addy *dialogue
	addy = DialogueMap[index]
	fmt.Println(addy.speaker.description)
	for true{
		fmt.Println(addy.speaker.name)
		fmt.Println(addy.text+"\n")
		if addy.childrenCount==0{
			return
		}
		if addy.speaker.id==0{
			addy = addy.children[0]
			continue
		}
		fmt.Println("\n///////////////\nChoose the best option: ")
		for i:=0;i<addy.childrenCount;i++{
			fmt.Println(strconv.Itoa(i)+": "+addy.children[i].text)
		}
		fmt.Println("///////////////")
    var input string
    fmt.Scanln(&input)
		choice,err := strconv.Atoi(input)
		if err!=nil{
			fmt.Println("Not a valid choice!")
		}
		addy = addy.children[choice]

	}
}
func main(){
	initializeNPCS()
	initializeDialogue()
	converse(1)
	converse(3)
}

//Allow multiple sets of dialogue now
