package rw

import (
	"bufio"
	"compress/gzip"
	"io"
	"os"
)

func GZipWriter() {
	file, err := os.Create("test.txt.gz")
	if err != nil {
		panic(err)
	}
	writer := gzip.NewWriter(file)
	defer writer.Close()

	writer.Header.Name = "test.txt"
	_, err = io.WriteString(writer, "gzip.Writer example\n")
	if err != nil {
		panic(err)
	}
}

func BufioWriter() {
	buffer := bufio.NewWriterSize(os.Stdout, 8)
	defer buffer.Flush()
	_, err := io.WriteString(buffer, "bufio.Writer ")
	if err != nil {
		panic(err)
	}
	_, err = io.WriteString(buffer, "example\n")
	if err != nil {
		panic(err)
	}
}
