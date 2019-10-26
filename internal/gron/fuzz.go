// +build gofuzz

package gron

import "bytes"

func Fuzz(b []byte) int {
	//   g: Gron
	//   s: GronStream
	//   u: Ungron
	if len(b) < 1 {
		return 0
	}

	roundTrip := true
	a := Gron

	switch b[0] {
	case 'g':
		a = Gron
		roundTrip = true
	case 'u':
		a = Ungron
		roundTrip = false
	case 's':
		a = GronStream
		roundTrip = false
	default:
		return -1
	}

	r := bytes.NewReader(b[1:])
	w := &bytes.Buffer{}

	_, err := a(r, w, 0)
	if err != nil {
		if len(w.Bytes()) != 0 {
			panic("err not nil and len not zero")
		}
		return 0
	}

	if roundTrip {
		output := w.Bytes()
		json := &bytes.Buffer{}
		_, err := Ungron(bytes.NewReader(output), json, 0)
		if err != nil {
			panic("should be able to make json if we successfully grond")
		}
	}

	return 1
}
