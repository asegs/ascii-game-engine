package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)
/**
Lets use JSON here and have some sort of online builder.
Some sort of format like:
{
	NPCs: ["---","---",...]
	Dialogues: [
		{
			Speaker: 1,
			Text: "......",
			Children: [...]
		}
	]
}

Sounds like very deeply recursive JSON.
 */




type NPC struct {
	name string
	description string
	id int
}

type Dialogue struct{
	speaker NPC
	text string
	childrenCount int
	children []*Dialogue
	order int
	id int
}

type DialogueSystem struct {
	Dialogues map[string] * Dialogue
	NPCMap map[int] * NPC
}

func initDialogueSystem (speakersDirectoryName string, dialogueDirectoryName string) (* DialogueSystem,error) {
	ds := &DialogueSystem{
		Dialogues: make(map[string] * Dialogue),
		NPCMap: make(map[int] * NPC),
	}
	speakersDirectory,err := os.Open(speakersDirectoryName)
	if err != nil {
		return nil, err
	}
	//read through speakers directory, parse files into NPCMap

	dialogueDirectory,err := os.Open(dialogueDirectoryName)
	if err != nil {
		return nil, err
	}

	//read through dialogue directory, parse files into dialogues
	return ds,nil
}


func getOrder(s string)(string,int){
	counter := 0
	for i,char := range s {
		if char == '^' {
			counter ++
		}else {
			return s[i:],counter
		}
	}
	//Empty line
	return "",-1
}

func insertDialogue (root * Dialogue, key string, order int,speaker NPC,id int) {
	if order == 1 {
		root.children = append(root.children,&Dialogue{speaker,key,0,make([] * Dialogue,0),order,id})
		root.childrenCount++
	}else {
		insertDialogue(root.children[root.childrenCount - 1],key,order - 1,speaker,id)
	}
}

func (d * Dialogue) converse(){
	root := d
	fmt.Println(root.speaker.description)
	for true{
		fmt.Println(root.speaker.name)
		fmt.Println(root.text+"\n")
		if root.childrenCount==0{
			return
		}
		if root.speaker.id==0{
			root = root.children[0]
			continue
		}
		fmt.Println("\n///////////////\nChoose the best option: ")
		for i:=0;i<root.childrenCount;i++{
			fmt.Println(strconv.Itoa(i)+": "+root.children[i].text)
		}
		fmt.Println("///////////////")
    var input string
    fmt.Scanln(&input)
		choice,err := strconv.Atoi(input)
		if err!=nil{
			fmt.Println("Not a valid choice!")
		}
		root = root.children[choice]

	}
}
// func main(){
// 	initializeNPCS()
// 	initializeDialogue()
// 	converse(1)
// 	converse(3)
// }

//Allow multiple sets of dialogue now
