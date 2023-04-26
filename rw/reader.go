package rw

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"os"
)

func CopyReader() {
	buf := bytes.NewBuffer([]byte("ABCDEFGHIJKLMN\n"))

	for {
		_, err := io.CopyN(os.Stdout, buf, 2)
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
	}

	buf = bytes.NewBuffer([]byte("ABCDEFGHIJKLMN\n"))
	buffer := make([]byte, 4)
	io.CopyBuffer(os.Stdout, buf, buffer)
}

func dumpChunk(chunk io.Reader) error {
	var length int32
	if err := binary.Read(chunk, binary.BigEndian, &length); err != nil {
		return err
	}
	buffer := make([]byte, 4)
	if _, err := chunk.Read(buffer); err != nil {
		return err
	}
	fmt.Printf("chunk '%v' (%d bytes)\n", string(buffer), length)
	if bytes.Equal(buffer, []byte("tEXt")) {
		rawText := make([]byte, length)
		chunk.Read(rawText)
		fmt.Println(string(rawText))
	}
	return nil
}

func readChunks(file *os.File) ([]io.Reader, error) {
	var chunks []io.Reader
	var offset int64 = 8
	if _, err := file.Seek(offset, 0); err != nil {
		return nil, err
	}

	for {
		var length int32
		err := binary.Read(file, binary.BigEndian, &length)
		if err == io.EOF {
			break
		}
		chunks = append(chunks, io.NewSectionReader(file, offset, int64(length)+12))
		offset, _ = file.Seek(int64(length+8), 1)
	}
	return chunks, nil
}

func textChunk(text string) (io.Reader, error) {
	byteData := []byte(text)
	var buffer bytes.Buffer
	// 長さ
	if err := binary.Write(&buffer, binary.BigEndian, int32(len(byteData))); err != nil {
		return nil, err
	}
	// 種類
	if _, err := io.WriteString(&buffer, "tEXt"); err != nil {
		return nil, err
	}
	// データ
	if _, err := buffer.Write(byteData); err != nil {
		return nil, err
	}
	// CRC
	crc := crc32.NewIEEE()
	if _, err := io.WriteString(crc, "tEXt"); err != nil {
		return nil, err
	}
	if _, err := crc.Write(byteData); err != nil {
		return nil, err
	}
	if err := binary.Write(&buffer, binary.BigEndian, crc.Sum32()); err != nil {
		return nil, err
	}
	return &buffer, nil
}

func PngReader() {
	file, err := os.Open("Lenna.png")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	newFile, err := os.Create("Lenna2.png")
	if err != nil {
		panic(err)
	}
	defer newFile.Close()

	chunks, err := readChunks(file)
	if err != nil {
		panic(err)
	}
	// シグニチャ
	if _, err := io.WriteString(newFile, "\x89PNG\r\n\x1a\n"); err != nil {
		panic(err)
	}
	if _, err := io.Copy(newFile, chunks[0]); err != nil {
		panic(err)
	}
	chunk, err := textChunk("ASCII PROGRAMMING++")
	if err != nil {
		panic(err)
	}
	if _, err := io.Copy(newFile, chunk); err != nil {
		panic(err)
	}
	for _, chunk := range chunks[1:] {
		if _, err := io.Copy(newFile, chunk); err != nil {
			panic(err)
		}
	}

	if _, err := newFile.Seek(0, 0); err != nil {
		panic(err)
	}
	newChunks, err := readChunks(newFile)
	if err != nil {
		panic(err)
	}
	for _, chunk := range newChunks {
		dumpChunk(chunk)
	}
}
