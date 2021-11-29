package main

import "time"

type Face struct {
	Name string
	Expressions map[string] * Frame
}

func (t * Terminal) playSingleExpression (face * Face,exp string) {
	if frame, ok := face.Expressions[exp]; ok {
		t.drawFrame(frame)
	}
}

//delay doesn't consider time to print, also cycles idea is weird, have a current state of face
//also doesn't terminate
func (t * Terminal) cycleExpressions (face * Face, exps [] string, msDelay int,cycles int){
	for cycles == -1 || 0 < cycles {
		for _,exp := range exps {
			t.playSingleExpression(face,exp)
			time.Sleep(time.Millisecond * time.Duration(msDelay))
		}
		if cycles != -1 {
			cycles --
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
