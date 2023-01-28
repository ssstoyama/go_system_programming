package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"
)

// パイプライニング 対応版
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
		go processSession(conn)
	}
}

func isGZipAcceptable(request *http.Request) bool {
	return strings.Index(
		strings.Join(request.Header["Accept-Encoding"], ","),
		"gzip",
	) != -1
}

// セッション 1 つを処理
func processSession(conn net.Conn) {
	fmt.Printf("Accept %v\n", conn.RemoteAddr())
	// セッション内のリクエストを順に処理するためのチャネル
	sessionResponses := make(chan chan *http.Response, 50)
	defer close(sessionResponses)
	// レスポンスを直列化してソケットに書き出す専用の goroutine
	go writeToConn(sessionResponses, conn)
	reader := bufio.NewReader(conn)
	// レスポンスを受け取ってセッションのキューに入れる
	for {
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		request, err := http.ReadRequest(reader)
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
		sessionResponse := make(chan *http.Response)
		sessionResponses <- sessionResponse
		// 非同期で実行
		go handleRequest(request, sessionResponse)
	}
}

func writeToConn(sessionResponses chan chan *http.Response, conn net.Conn) {
	defer conn.Close()
	for sessionResponse := range sessionResponses {
		response := <-sessionResponse
		response.Write(conn)
		close(sessionResponse)
	}
}

func handleRequest(request *http.Request, resultReceiver chan *http.Response) {
	dump, err := httputil.DumpRequest(request, true)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(dump))
	var content string
	if message := request.URL.Query().Get("message"); message != "" {
		content = fmt.Sprintf("%s\n", message)
	} else {
		content = "Hello World\n"
	}
	// レスポンスを書き込む
	// セッションを維持するために Keep-Alive でないといけない
	response := &http.Response{
		StatusCode:    200,
		ProtoMajor:    1,
		ProtoMinor:    1,
		ContentLength: int64(len(content)),
		Body:          ioutil.NopCloser(strings.NewReader(content)),
	}
	// 処理が終わったらチャネルに書き込み、
	// ブロックされていた writeToConn の処理を再始動する
	resultReceiver <- response
}
