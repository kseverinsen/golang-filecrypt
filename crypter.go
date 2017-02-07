// AES-GCM for file chunks

package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"io"
)

// encryp plaintext and additonal data (can be nil) using key.
func encrypt(key, plaintext, ad []byte) ([]byte, []byte) {
	// Init block
	block, err := aes.NewCipher(key)
	check(err)

	// Random nonce
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}

	// init cipher
	aesgcm, err := cipher.NewGCM(block)
	check(err)

	// return nonce and ciphertext
	return nonce, aesgcm.Seal(nil, nonce, plaintext, ad)
}

// Decrypt ciphertext using key, nonce and additonal data
func decrypt(key, nonce, ciphertext, ad []byte) []byte {
	// Init block
	block, err := aes.NewCipher(key)
	check(err)

	// init cipher
	aesgcm, err := cipher.NewGCM(block)
	check(err)

	// decryption and integrity check
	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, ad)
	check(err)

	return plaintext
}

// chunker hook to encrypt data field of chunk
func encrypter(key []byte, pt, ct chan chunk) {

	for {
		// receive unencypted chunks
		c := <-pt
		if c.data == nil { // done
			ct <- c // forward signal
			break
		}
		// encode chunk id to byte array
		ad := make([]byte, 8)
		binary.LittleEndian.PutUint64(ad, c.id)

		// encrypt data field, and add chunk id (for integrity check)
		nonce, ciphertext := encrypt(key, c.data, ad)
		c.data = append(nonce, ciphertext...)

		// forward encrypted chunk
		ct <- c
	}

}

// hook to decrypt data field of chunks
func decrypter(key []byte, ct, pt chan chunk) {

	for {
		// receive encrypted chunks
		c := <-ct
		if c.data == nil { // done
			pt <- c // forward signal
			break
		}
		// split data into nonce and ciphertext
		nonce := c.data[:12]
		ciphertext := c.data[12:]

		// encode chunk id to byte array
		ad := make([]byte, 8)
		binary.LittleEndian.PutUint64(ad, c.id)

		// decrypt and integrity check data field
		plaintext := decrypt(key, nonce, ciphertext, ad)
		c.data = plaintext

		// forward plaintext chunk
		pt <- c

	}
}
