package main

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
)

// gzip 対応版
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
		// gzip 対応クライアントであることをサーバーに伝える
		request.Header.Set("Accept-Encoding", "gzip")

		request.Write(conn)
		response, err := http.ReadResponse(bufio.NewReader(conn), request)
		if err != nil {
			// タイムアウトはここでエラーになるためリトライ
			fmt.Println("Retry")
			conn = nil
			continue
		}

		// DumpResponse は圧縮された内容を理解できないため false で Body を無視する
		dump, err := httputil.DumpResponse(response, false)
		if err != nil {
			panic(err)
		}
		fmt.Println(strings.Repeat("-", 30))
		fmt.Println(string(dump))

		defer response.Body.Close()
		if response.Header.Get("Content-Encoding") == "gzip" {
			// gzip 対応サーバーからのレスポンス処理
			reader, err := gzip.NewReader(response.Body)
			if err != nil {
				panic(err)
			}
			io.Copy(os.Stdout, reader)
			reader.Close()
		} else {
			io.Copy(os.Stdout, response.Body)
		}
		fmt.Println()

		current++
		if current >= len(sendMessages) {
			break
		}
	}
	conn.Close()
}
