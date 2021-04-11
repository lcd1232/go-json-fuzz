// +build gofuzz

package json

import (
	"bytes"
	stdjson "encoding/json"
	"fmt"
	"reflect"

	"github.com/dvyukov/go-fuzz-corpus/fuzz"
	"github.com/goccy/go-json"
)

func fixS(v interface{}) {
	if s, ok := v.(*S); ok {
		if len(s.P) == 0 {
			s.P = []byte(`""`)
		}
	}
}

func Fuzz(data []byte) int {
	score := 0
	for _, factory := range []func() interface{}{
		func() interface{} { return nil },
		func() interface{} { return new([]interface{}) },
		func() interface{} { m := map[string]string{}; return &m },
		func() interface{} { m := map[string]interface{}{}; return &m },
		func() interface{} { return new(S) },
	} {
		v1 := factory()
		v2 := factory()
		err1 := stdjson.Unmarshal(data, &v1)
		err2 := json.Unmarshal(data, &v2)

		if err1 != nil {
			if err2 != nil {
				if reflect.TypeOf(err1) != reflect.TypeOf(err2) {
					fmt.Printf("input: %s\n", string(data))
					fmt.Printf("encoding/json.Unmarshal(%T): %T: %s\n", v1, err1, err1)
					fmt.Printf("goccy/go-json.Unmarshal(%T): %T: %s\n", v2, err2, err2)
					panic("error types mismatch")
				}
				continue
			} else {
				fmt.Printf("input: %s\n", string(data))
				fmt.Printf("encoding/json.Unmarshal(%T): %T: %s\n", v1, err1, err1)
				fmt.Printf("goccy/go-json.Unmarshal(%T): <nil>\n")
				panic("error values mismatch")
			}
		} else {
			if err2 != nil {
				fmt.Printf("input: %s\n", string(data))
				fmt.Printf("encoding/json.Unmarshal(%T): <nil>\n")
				fmt.Printf("goccy/go-json.Unmarshal(%T): %T: %s\n", v2, err2, err2)
				panic("error values mismatch")
			} else {
				// both implementations pass
			}
		}

		score = 1
		fixS(v1)
		fixS(v2)
		if !fuzz.DeepEqual(v1, v2) {
			fmt.Printf("input: %s\n", string(data))
			fmt.Printf("encoding/json: %#v\n", v1)
			fmt.Printf("goccy/go-json: %#v\n", v2)
			panic("not equal")
		}

		data1, err := stdjson.Marshal(v1)
		if err != nil {
			panic(err)
		}
		data2, err := json.Marshal(v2)
		if err != nil {
			panic(err)
		}
		if !bytes.Equal(data1, data2) {
			fmt.Printf("input: %q\n", string(data))
			fmt.Printf("encoding/json: %q\n", string(data1))
			fmt.Printf("goccy/go-json: %q\n", string(data2))
			panic("not equal")
		}
	}
	return score
}

type S struct {
	A int    `json:",omitempty"`
	B string `json:"B1,omitempty"`
	C float64
	D bool
	E uint8
	F []byte
	G interface{}
	H map[string]interface{}
	I map[string]string
	J []interface{}
	K []string
	L S1
	M *S1
	N *int
	O **int
	P json.RawMessage
	Q Marshaller
	R int `json:"-"`
	S int `json:",string"`
}

type S1 struct {
	A int
	B string
}

type Marshaller struct {
	v string
}

func (m *Marshaller) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.v)
}

func (m *Marshaller) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &m.v)
}
