package main

import (
       "fmt"
       "bytes"
       "net/http"
       "io/ioutil"
       "sync"
       "strconv"
)

type dropperDownloader struct{
	numConnections int
	uri string
	fileChunks map[int][]byte
	contentLength int
	*sync.Mutex
}

func newDropperDownloader() *dropperDownloader {
	return &dropperDownloader{
		numConnections: 0,
		uri: "",
		fileChunks: make(map[int][]byte, 0),
		contentLength:0,
		Mutex: &sync.Mutex{},
	}
}

func (dd *dropperDownloader) downloadForRange(wg *sync.WaitGroup, r string, index int, e chan error) {
	if wg != nil {
		defer wg.Done()
	}
	
	req, err := http.NewRequest("GET", dd.uri, nil)
	if err != nil {
		e <- fmt.Errorf("error when creating %d connection request: %v",index, err)
		return
	}
	
	if r != "" {
		req.Header.Add("Range", "bytes="+r)
	}
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		e <- fmt.Errorf("error sending %d connection request: %v", index, err)
		return
	}
	
	if resp.StatusCode != 200 && resp.StatusCode != 206{
		e <- fmt.Errorf("did not get appropriate response on %d connection response: %v", index, err)
		return
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		e <- fmt.Errorf("malformed body on %d connection body: %v", index, err)
		return
	}
	
	dd.Lock()
	dd.fileChunks[index] = append(dd.fileChunks[index], data...)
	dd.Unlock()
} 

func (dd *dropperDownloader) process() ([]byte, error) {
	if (dd.numConnections < 2) {
		return dd.singleDownload()
	}
	
	potions := dd.contentLength / dd.numConnections
	var err chan error
	wg := &sync.WaitGroup{}
	
	index := 0
	for i := 0; i < dd.contentLength; i+= potions + 1 {
		j := i + potions
		if j < dd.contentLength {
			j = dd.contentLength
		}
		
		dd.fileChunks[index] = make([]byte, 0)
		wg.Add(1)
		go dd.downloadForRange(wg, strconv.Itoa(i)+"-"+strconv.Itoa(j), index, err)
		index++
	}
	wg.Wait()
	
	return dd.combineChunks(), nil
}

func (dd *dropperDownloader) combineChunks() []byte {
	buf := bytes.NewBuffer(nil)
	for i := 0; i < len(dd.fileChunks); i++ {
		buf.Write(dd.fileChunks[i])
	}
	
	return buf.Bytes()
}

// check if Range header is supported and fill content-length
func (dd *dropperDownloader)isRangeSupported() (bool, error) {
	resp, err := http.Head(dd.uri)
	if err != nil {
		return false, err
	}
	
	if resp.StatusCode != 200 && resp.StatusCode != 206 {
		return false, fmt.Errorf("Range is not supported. Status Code %d\n", resp.StatusCode)
	}
	
	dd.contentLength, err = strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
		return false, fmt.Errorf("Error with 'Content-Length'")
	}
	
	// Check Accept-Ranges
	if resp.Header.Get("Accept-Ranges") == "bytes" {
		return true, nil
	}
	
	return false, nil
}

// download is used when numConnections is 1
// mostly for use when range is not specified
func(dd *dropperDownloader) singleDownload() ([]byte, error) {
	client := &http.Client{}
	response, err := client.Get(dd.uri)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	return ioutil.ReadAll(response.Body)
}
