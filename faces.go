package main

import "time"

type Face struct {
	Name string
	Expressions map[string] * Frame
}

var interrupt = make(chan bool,3)

func (t * TerminalClient) playSingleExpression (face * Face,exp string,zone * Zone,interruptCycles bool) {
	if interruptCycles {
		interrupt <- true
	}
	if frame, ok := face.Expressions[exp]; ok {
		t.drawFrame(frame,zone.Y,zone.X)
	}
}

//delay doesn't consider time to print, also cycles idea is weird, have a current state of face
func (t * TerminalClient) cycleExpressions (face * Face, exps [] string, msDelay int,cycles int,zone * Zone){
	for i := 0 ; i < cycles ; i ++ {
		for _,exp := range exps {
			if len(interrupt) > 0 {
				return
			}
			t.playSingleExpression(face,exp,zone,false)
			time.Sleep(time.Millisecond * time.Duration(msDelay))
		}
	}
}


func buildFace (exps [] string,filenames [] string,name string) * Face{
	expMap := make(map[string] * Frame)
	for i,exp := range exps {
		expMap[exp] = buildFrame(filenames[i])
	}
	return &Face{
		Name:        name,
		Expressions: expMap,
	}
}
