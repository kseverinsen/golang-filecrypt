package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"runtime"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

//Read key from STDIN
func readKey() []byte {
	fmt.Print("Encryption key: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()

	key := scanner.Bytes()
	if len(key) > 32 {
		panic("Key too long: max 32 characters")
	}
	if len(key) != 32 { // pad key to 32 bytes
		pad := 32 - len(key)%32
		key = append(key, bytes.Repeat([]byte{0}, pad)...)
	}
	//log.Println("key: ", key)
	return key
}

func main() {
	runtime.GOMAXPROCS(4)
	fmt.Println("Toy file encrypter!")
	scanner := bufio.NewScanner(os.Stdin)
	//	Choose mode
	fmt.Print("Encrypt[0] or Decrypt[1] file: ") // hack TODO: use command line flags :)
	scanner.Scan()
	var ENCRYPT bool
	if scanner.Bytes()[0] == 48 {
		ENCRYPT = true
	} else if scanner.Bytes()[0] == 49 {
		ENCRYPT = false
	} else {
		panic("allowed input is 0 or 1")
	}

	// Read input filename from STDIN
	fmt.Print("Input file: ")
	scanner.Scan()
	inFilename := scanner.Text()
	outFilename := string(append([]byte(inFilename), ".out"...))
	// Open input file
	fi, err := os.Open(inFilename)
	check(err)
	defer fi.Close()
	// create ouput file
	fo, err := os.Create(outFilename)
	check(err)
	defer fo.Close()

	// get key
	key := readKey()

	chunkChan := make(chan chunk)
	cryptChan := make(chan chunk)

	chunksize := uint64(4096)
	cryptSize := uint64(4124) //magic nuber hack..  chunksize + overhead TODO: caclulate overhead.

	if ENCRYPT {
		go chunker(fi, chunkChan, chunksize)
		go encrypter(key, chunkChan, cryptChan)
		dechunker(fo, cryptChan, cryptSize)
	} else {
		go chunker(fi, chunkChan, cryptSize)
		go decrypter(key, chunkChan, cryptChan)
		dechunker(fo, cryptChan, chunksize)
	}

}
