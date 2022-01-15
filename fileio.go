package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

func readAll(r io.Reader, capacity int64) (b []byte, err error) {
	var buf bytes.Buffer
	defer func() {
		e := recover()
		if e == nil {
			return
		}
		if panicErr, ok := e.(error); ok && panicErr == bytes.ErrTooLarge {
			err = panicErr
		} else {
			panic(e)
		}
	}()
	if int64(int(capacity)) == capacity {
		buf.Grow(int(capacity))
	}
	_, err = buf.ReadFrom(r)
	return buf.Bytes(), err
}

func ReadFile(filename string) ([]byte, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var n int64 = bytes.MinRead

	if fi, err := f.Stat(); err == nil {
		if size := fi.Size() + bytes.MinRead; size > n {
			n = size
		}
	}
	return readAll(f, n)
}

func ReadToString(filename string) (string,error) {
	s,err:=ReadFile(filename)
	if err != nil {
		return "",err
	}
	return string(s),err
}

func Write(filename string, body string) {

	f, err := os.Create(filename)

	if err != nil {
		fmt.Println(err.Error())
	}

	defer f.Close()

	_, err2 := f.WriteString(body)

	if err2 != nil {
		fmt.Println(err2.Error())
	}

	//fmt.Printf("Wrote to file: %s\n", filename)
}