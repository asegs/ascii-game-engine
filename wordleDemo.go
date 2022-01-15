package main

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

const wHeight int = 40
const wWidth int = 100
const defLen int = 5
const defGuesses int = 6

type Validity int

const (
	RightPlace Validity = iota
	WrongPlace
	NotPresent
)

type LetterHistory struct {
	Validity Validity
	Letter byte
}

var s1 = rand.NewSource(time.Now().UnixNano())
var r1 = rand.New(s1)

func getAllWords (size int,directory string) ([] string,error) {
	numberAsString := strconv.Itoa(size)
	filename := directory+"/"+numberAsString+".txt"
	wordsText,err := ReadToString(filename)
	if err != nil {
		LogString("No words of this length found!")
		return nil,err
	}
	return strings.Split(wordsText,"\n"),nil
}

func pickRandomly (words [] string) (string,error) {
	if len(words) == 0 {
		LogString("No words provided to random picker.")
		return "",errors.New("no words to select on")
	}
	return words[r1.Intn(len(words))],nil
}

func checkLetterValidity (history [] [] LetterHistory, letter byte, place int) Validity {
	for _,attempt := range history {
		for i,item := range attempt {
			if item.Validity == NotPresent && item.Letter == letter{
				return NotPresent
			} else if item.Validity ==  WrongPlace && item.Letter == letter && i == place {
				return WrongPlace
			}
		}
	}
	return RightPlace
}

func letterInWord (word string, letter rune) bool {
	for _,wordLetter := range word {
		if wordLetter == letter {
			return true
		}
	}
	return false
}

func wordInOptions (options [] string, word string) bool {
	for _,option := range options {
		if word == option {
			return true
		}
	}
	return false
}

func makeGuess (realWord string, guess string) (bool, [] LetterHistory) {
	guessStatus := make( [] LetterHistory,len(realWord))
	for i := 0 ; i < len(realWord) ; i++ {
		validity := RightPlace
		if realWord[i] != guess[i] && letterInWord(realWord, rune(guess[i])) {
			validity = WrongPlace
		}else if !letterInWord(realWord,rune(guess[i])) {
			validity = NotPresent
		}
		guessStatus[i] = LetterHistory{
			Validity: validity,
			Letter:   guess[i],
		}
	}
	return realWord == guess,guessStatus
}

func main()  {
	go HandleLog()
	words,err := getAllWords(defLen,"wordleWords")
	if err != nil {
		LogString(err.Error())
		return
	}
	toGuess,err := pickRandomly(words)
	if err != nil {
		LogString(err.Error())
		return
	}
	guessCount := 0
	currentWord := ""
	currentHistory := make([][] LetterHistory,defGuesses)
	for i := 0 ; i < defGuesses ; i ++ {
		currentHistory[i] = make([] LetterHistory,defLen)
	}
	input := initializeInput()
	valid := initContext().addRgbStyleBg(26, 158, 0).addRgbStyleFg(0,0,0).compile()
	wrongPlace := initContext().addRgbStyleBg(245, 201, 105).addRgbStyleFg(0,0,0).compile()
	invalid := initContext().addRgbStyleBg(128, 128, 128).addRgbStyleFg(0,0,0).compile()
	clear := initContext().addSimpleStyle(0).compile()
	struck := initContext().addSimpleStyle(9).addRgbStyleBg(128, 128, 128).addRgbStyleFg(0,0,0).compile()
	terminal := createTerminal(wHeight,wWidth,&Recorded{
		Format:         clear,
		ShownSymbol:    ' ',
		BackgroundCode: '0',
	},8)
	zoning := initZones(wHeight,wWidth,input)
	gameZone,err := zoning.createZone(0,0,wHeight,wWidth,true)
	if err != nil {
		LogString("For some reason no zone was created: " + err.Error())
		return
	}
	err = zoning.cursorEnterZone(gameZone,0)
	if err != nil {
		LogString("For some reason the zone wasn't selected: " + err.Error())
	}
	zoning.setDefaultZone(gameZone)
	var msg * NetworkedMsg
	for {
		topRowKeys := [...]byte{'q','w','e','r','t','y','u','i','o','p'}
		topShift := 8
		middleRowKeys := [...]byte{'a','s','d','f','g','h','j','k','l'}
		midShift :=9
		bottomRowKeys := [...]byte{'z','x','c','v','b','n','m'}
		botShift := 10
		for i,b := range topRowKeys {
			validity := checkLetterValidity(currentHistory,b,len(currentWord))
			properCtx := valid
			switch validity {
			case WrongPlace:
				properCtx = wrongPlace
				break
			case NotPresent:
				properCtx = struck
				break
			}
			terminal.sendPlaceCharFormat(b,0,i + defLen + topShift,properCtx,'0')
		}

		for i,b := range middleRowKeys {
			validity := checkLetterValidity(currentHistory,b,len(currentWord))
			properCtx := valid
			switch validity {
			case WrongPlace:
				properCtx = wrongPlace
				break
			case NotPresent:
				properCtx = struck
				break
			}
			terminal.sendPlaceCharFormat(b,1,i + defLen + midShift,properCtx,'0')
		}

		for i,b := range bottomRowKeys {
			validity := checkLetterValidity(currentHistory,b,len(currentWord))
			properCtx := valid
			switch validity {
			case WrongPlace:
				properCtx = wrongPlace
				break
			case NotPresent:
				properCtx = struck
				break
			}
			terminal.sendPlaceCharFormat(b,2,i + defLen + botShift,properCtx,'0')
		}
		msg = <- gameZone.Events
		if 'a' <= msg.Msg && msg.Msg <= 'z' {
			if checkLetterValidity(currentHistory,msg.Msg,len(currentWord)) == RightPlace && len(currentWord) < len(toGuess){
				currentWord += string(msg.Msg)
				terminal.sendPlaceCharFormat(msg.Msg,guessCount,len(currentWord) - 1,clear,'0')
			}
		}else if msg.Msg == BACKSPACE {
			if len(currentWord) > 0 {
				currentWord = currentWord[0:len(currentWord) - 1]
				terminal.sendUndoAtLocationConditional(guessCount,len(currentWord),'0',false)
			}
		}else if msg.Msg == ENTER {
			if len(currentWord) < len(toGuess) {
				continue
			}
			if !wordInOptions(words,currentWord){
				for i := 0 ; i <= defLen ; i++ {
					terminal.sendPlaceCharFormat(' ',guessCount,i,clear,'0')
				}
				currentWord = ""
				continue
			}
			success,results := makeGuess(toGuess,currentWord)
			for i,result := range results {
				properCtx := valid
				switch result.Validity {
				case WrongPlace:
					properCtx = wrongPlace
					break
				case NotPresent:
					properCtx = invalid
					break
				}
				terminal.sendPlaceCharFormat(result.Letter,guessCount,i,properCtx,'0')
				if success || guessCount > defGuesses {
					if success {
						terminal.sendPrintStyleAtCoord(valid,guessCount + 1,0,"Correct!")
					} else {
						terminal.sendPrintStyleAtCoord(invalid,guessCount + 1,0,"Out of guesses!  Word was: " + toGuess)
					}
					fmt.Println()
					fmt.Println()
					fmt.Println()
					time.Sleep(5 * time.Second)
					return
				}
			}
			guessCount ++
			currentWord = ""
			currentHistory = append(currentHistory,results)

		}else if msg.Msg == BACKSLASH {
				terminal.sendPrintStyleAtCoord(invalid,guessCount + 1,0,"Out of guesses!  Word was: " + toGuess)
				fmt.Println()
				fmt.Println()
				fmt.Println()
				time.Sleep(5 * time.Second)
				return
			}
		}
}

