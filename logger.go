package main
import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)


var loggingChannel = make(chan Log,1000)

type Log struct {
	Time time.Time
	Message string
}

//straight from go docs https://golang.org/pkg/os/#example_OpenFile_append
func AppendToFile(data string,filename string)error{
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil{
		return err
	}
	if _, err := f.WriteString(data); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return nil
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func SafeEventLog(log Log){
	toLog := fmt.Sprintf("Event logged at %v :: %s\n",log.Time,log.Message)
	err := AppendToFile(toLog,"logs/server.log")
	if err != nil{
		fmt.Println("Failed to log: "+err.Error())
	}
	return
}

func LogString(msg string){
	l := Log{
		Time:    time.Now(),
		Message: msg,
	}
	loggingChannel <- l
}

func HandleLog(){
	newLogsPath := filepath.Join(".","logs")
	err := os.MkdirAll(newLogsPath,os.ModePerm)
	if err != nil {
		//already existed, fine!
	}
	if !fileExists("logs/server.log"){
		_,err = os.Create("logs/server.log")
		if err != nil {
			//something doesn't work
		}
	}


	for true{
		log :=<- loggingChannel
		SafeEventLog(log)
	}

}

//call go handleLog() in main func
