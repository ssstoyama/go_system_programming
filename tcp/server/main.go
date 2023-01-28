package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"time"
)

// Keep-Alive 対応版
func main() {
	listener, err := net.Listen("tcp", "localhost:8888")
	if err != nil {
		panic(err)
	}
	fmt.Println("Server is running at localhost:8888")

	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		go func() {
			defer conn.Close()
			fmt.Printf("Accept %v\n", conn.RemoteAddr())
			// Accept 後のソケットで何度も応答を返すためにループ
			// TCP のコネクションが張られたあと何度もリクエストを受けられる
			for {
				// 通信がしばらくない場合タイムアウトする
				conn.SetReadDeadline(time.Now().Add(5 * time.Second))

				request, err := http.ReadRequest(bufio.NewReader(conn))
				if err != nil {
					neterr, ok := err.(net.Error)
					if ok && neterr.Timeout() {
						fmt.Println("Timeout")
						break
					} else if err == io.EOF {
						break
					}
					panic(err)
				}
				dump, err := httputil.DumpRequest(request, true)
				if err != nil {
					panic(err)
				}
				fmt.Println(string(dump))

				// レスポンスを書き込む
				// HTTP/1.1 かつ、ContentLength の設定が必要
				response := http.Response{
					StatusCode:    http.StatusOK,
					ProtoMajor:    1,
					ProtoMinor:    1,
					ContentLength: int64(len(dump)),
					Body:          ioutil.NopCloser(bytes.NewReader(dump)),
				}
				response.Write(conn)
			}
		}()
	}
}
