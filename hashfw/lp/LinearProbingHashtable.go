package lp

import (
	"hashfw/hashfw/fw"
	"hashfw/hashfw/fw/hash"
	"unsafe"
)

type hashtable fw.Hashtable[string, int, uint64, uint64]

type LinearProbingHashtable struct {
	fw hashtable
}

func New(htFw hashtable) *LinearProbingHashtable {
	return &LinearProbingHashtable{
		fw: htFw,
	}
}

func (lp *LinearProbingHashtable) Finalizer(l []string) uint64 {
	data := []byte(l[0])
	toHash := *(*uint64)(unsafe.Pointer(&data[0]))
	return hash.MurmurFinalizerHash64(toHash)
}
