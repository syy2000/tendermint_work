package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/tendermint/tendermint/crypto/tmhash"
)

const (
	hashTimes = 1000000
)

func TestHash(t *testing.T) {
	var set0 = []byte("d")
	var set = []byte{}
	start := time.Now()
	for i := 0; i < hashTimes; i++ {
		set = tmhash.Sum(set0)
	}
	fmt.Printf("%x : %s\n", set[:16], time.Since(start))
}
