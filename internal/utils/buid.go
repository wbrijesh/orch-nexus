package utils

import (
	"crypto/rand"
	"encoding/binary"
	"math/big"
	"time"
)

const base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

// base58Encode encodes a byte slice to a base58 string.
func base58Encode(input []byte) string {
	num := new(big.Int).SetBytes(input)
	base := big.NewInt(58)
	zero := big.NewInt(0)
	mod := new(big.Int)
	var result []byte

	for num.Cmp(zero) > 0 {
		num.DivMod(num, base, mod)
		result = append(result, base58Alphabet[mod.Int64()])
	}
	// Handle leading zeros
	for _, b := range input {
		if b != 0 {
			break
		}
		result = append(result, base58Alphabet[0])
	}
	// Reverse result
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	return string(result)
}

// base58Decode decodes a base58 string to a byte slice.
func base58Decode(input string) []byte {
	num := big.NewInt(0)
	base := big.NewInt(58)
	for _, c := range input {
		index := int64(-1)
		for i, ac := range base58Alphabet {
			if c == ac {
				index = int64(i)
				break
			}
		}
		if index == -1 {
			return nil // invalid character
		}
		num.Mul(num, base)
		num.Add(num, big.NewInt(index))
	}
	// Convert big.Int to byte slice
	buf := num.Bytes()
	// Handle leading zeros
	zeros := 0
	for _, c := range input {
		if c != rune(base58Alphabet[0]) {
			break
		}
		zeros++
	}
	return append(make([]byte, zeros), buf...)
}

// NewBUID generates a new, time-sortable, base58-encoded unique ID (≤10 chars).
// It's meant to be used as an alternative to UUIDs for unique identifiers.
func NewBUID() (string, error) {
	// 4 bytes for timestamp (seconds)
	now := uint32(time.Now().Unix())
	// 3 bytes for randomness
	random := make([]byte, 3)
	_, err := rand.Read(random)
	if err != nil {
		return "", err
	}
	// Combine
	buf := make([]byte, 7)
	binary.BigEndian.PutUint32(buf[:4], now)
	copy(buf[4:], random)
	id := base58Encode(buf)
	// Pad or trim to 10 chars if needed
	if len(id) < 10 {
		id = id + base58Alphabet[:10-len(id)]
	} else if len(id) > 10 {
		id = id[:10]
	}
	return id, nil
}

// DecodeBUIDTimestamp decodes the timestamp from the BUID string.
// Returns the timestamp as time.Time and the raw uint32 value.
func DecodeBUIDTimestamp(id string) (time.Time, uint32) {
	// Decode base58
	buf := base58Decode(id)
	if len(buf) < 4 {
		return time.Time{}, 0
	}
	timestamp := binary.BigEndian.Uint32(buf[:4])
	return time.Unix(int64(timestamp), 0), timestamp
}
