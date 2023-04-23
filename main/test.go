package main

import (
	"math/rand"
)

type myStruct struct {
}

type aa interface {
	getInt() int
}

type mytest struct {
}

func (m *mytest) getInt() int {
	return 1
}

func ThisFunc(m aa) string {
	switch m.getInt() {
	case 1:
		return "aaa"
	case 2:
		return "bbb"
	case 5:
		return "ccc"
	case 10:
		return "ddd"
	case 30:
		return "eee"
	case 80:
		return "fff"
	default:
		return "not found"
	}
}

func (m *myStruct) getInt() int {
	return rand.Intn(100)
}
