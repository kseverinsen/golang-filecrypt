package main

import (
	"bufio"
	"io"
	"os"
)

type chunk struct {
	data []byte
	id   uint64
}

func chunker(f *os.File, stream chan chunk, chunksize uint64) {
	var id uint64
	id = 0

	reader := bufio.NewReader(f)

	for {
		// Create a new chunk
		c := chunk{make([]byte, chunksize), id}
		// Read to chunk
		n, err := reader.Read(c.data)
		if err != nil && err != io.EOF {
			check(err)
		}
		// If done, send a nil chunk
		if n == 0 {
			stream <- chunk{nil, 0}
			break
		}
		// resize chunk if small file or last chunk
		if uint64(n) < chunksize {
			b := make([]byte, n)
			copy(b, c.data[:n])
			c.data = b
		}

		// send chunk and increment position
		stream <- c
		id++
	}
}

func dechunker(f *os.File, chunkStream chan chunk, chunksize uint64) {
	writer := bufio.NewWriter(f)
	for {
		// Get chunk from channel
		c := <-chunkStream
		if c.data == nil { // hack.. TODO: send done signal on separate channel
			break
		}
		// Seek to right posistion in file
		_, err := f.Seek(int64(c.id*chunksize), 0)
		check(err)
		// Write data to buffer
		_, err = writer.Write(c.data)
		check(err)
		// Flush write buffer
		writer.Flush()
	}

}
