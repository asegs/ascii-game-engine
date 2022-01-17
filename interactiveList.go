package main

import "fmt"

var validPunctuation = [...]uint8{' ',',','.','-','!','?'}

func inPunctuation (letter uint8) bool {
	for _,punc := range validPunctuation {
		if letter == punc {return true}
	}
	return false
}

func makeMultilineString(body string,maxWidth int) [] string {
	maxWidth ++
	toReturn := make([] string,0)
	if maxWidth <= 2 {
		return toReturn
	}
	lengthHandled := 0
	bodyLength := len(body)
	for lengthHandled <= bodyLength {
		for lengthHandled < bodyLength && body[lengthHandled] == ' ' {
			lengthHandled ++
		}
		newLength := lengthHandled + maxWidth - 1
		if newLength >= bodyLength {
			toReturn = append(toReturn,body[lengthHandled:])
			lengthHandled = newLength
		}else {
			if inPunctuation(body[newLength]) {
				toReturn = append(toReturn,body[lengthHandled:newLength])
				lengthHandled = newLength
			}else {
				newLength --
				if inPunctuation(body[newLength]){
					toReturn = append(toReturn,body[lengthHandled:newLength+1])
					lengthHandled = newLength + 1
				}else {
					toReturn = append(toReturn,body[lengthHandled:newLength] + "-")
					lengthHandled = newLength
				}
			}
		}
	}
	return toReturn
}

//func (t * Terminal) drawList (options [] string,selected int,highlighted * Context, standard * Context, zone * Zone) {
//	for i,option := range options {
//		if i == selected {
//			t.sendPrintStyleAtCoord()
//		}
//	}
//}

func printArr (words [] string){
	for _,word := range words {
		fmt.Println(word)
	}
}

func main () {
	for i := 1 ; i < 40 ; i ++ {
		printArr(makeMultilineString("Hello my dearest friend, I need to ask you a favor.",i))
	}
}
