package main

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
)

// Keep-Alive 対応版
func main() {
	sendMessages := []string{
		"ASCII",
		"PROGRAMMING",
		"PLUS",
	}
	current := 0
	var conn net.Conn = nil
	// リトライするためループで囲む
	for {
		var err error
		// まだコネクションを張ってない or エラーでリトライ
		if conn == nil {
			conn, err = net.Dial("tcp", "localhost:8888")
			if err != nil {
				panic(err)
			}
			fmt.Printf("Access: %d\n", current)
		}
		request, err := http.NewRequest("POST", "http://localhost:8888", strings.NewReader(sendMessages[current]))
		if err != nil {
			panic(err)
		}
		request.Write(conn)
		response, err := http.ReadResponse(bufio.NewReader(conn), request)
		if err != nil {
			// タイムアウトはここでエラーになるためリトライ
			fmt.Println("Retry")
			conn = nil
			continue
		}
		dump, err := httputil.DumpResponse(response, true)
		if err != nil {
			panic(err)
		}
		fmt.Println(strings.Repeat("-", 30))
		fmt.Println(string(dump))
		current++
		if current >= len(sendMessages) {
			break
		}
	}
	conn.Close()
}
