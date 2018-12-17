package library

import (
	"log"
	"net/http"
	"errors"
	"fmt"
	"crypto/rand"
)

func GetSize(urls string) (int64, error) {
	cl := http.Client{}
	resp, err := cl.Head(urls)
	if err != nil {
		log.Printf("error: when try get file size %v \n", err)
		return 0, err
	}
	if resp.StatusCode != 200 {
		log.Printf("error: file not found or moved status:%d", resp.StatusCode)
		return 0, errors.New("error: file not found or moved")
	}
	log.Printf("info: file size is %d bytes \n", resp.ContentLength)
	return resp.ContentLength, nil
}

func GenUid() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

