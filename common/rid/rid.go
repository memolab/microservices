// Package rid generating random IDs.
package rid

import "math/rand"

var az = [62]uint8{
	'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
	'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
	'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
}

// New8 return a new rand char length 8
func New8() string {
	var ar [8]byte
	for i := range ar {
		ar[i] = az[rand.Intn(62)]
	}
	ar[0] = az[rand.Intn(52)]
	return string(ar[:])
}

// New16 return a new rand char length 16
func New16() string {
	var ar [16]byte
	for i := range ar {
		ar[i] = az[rand.Intn(62)]
	}
	ar[0] = az[rand.Intn(52)]
	return string(ar[:])
}
