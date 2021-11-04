package main

const MAX_MESSAGES int = 1000

type ContextMessage struct {
	Format string
	Body string
	Row int
	Col int
}

type Terminal struct {
	Row int
	Col int
	Height int
	Width int
	Feed chan ContextMessage
}

func createTerminal(height int,width int)*Terminal{
	feed := make(chan ContextMessage,MAX_MESSAGES)
	terminal := &Terminal{
		Row:  0,
		Col:  height + 1,
		Feed: feed,
	}
	for i := 0;i<height;i++ {
		for b := 0;b<width;b++ {
			print(" ")
		}
		println()
	}
	return terminal
}

func (t * Terminal) send (context string,body string,row int,col int){
	t.Feed <- ContextMessage{
		Format: context,
		Body:   body,
		Row:    row,
		Col:    col,
	}
}

func (t * Terminal) cursorLeft(n int){
	//
}

func (t * Terminal) handleRenders(){
	var ctx ContextMessage
	for true{
		ctx = <- t.Feed

	}
}

/*
func main() {
    // disable input buffering
    exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()
    // do not display entered characters on the screen
    exec.Command("stty", "-F", "/dev/tty", "-echo").Run()
    // restore the echoing state when exiting
    defer exec.Command("stty", "-F", "/dev/tty", "echo").Run()

    var b []byte = make([]byte, 1)
    for {
        os.Stdin.Read(b)
        fmt.Println("I got the byte", b, "("+string(b)+")")
    }
}

 */