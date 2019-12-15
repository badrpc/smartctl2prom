package smartctldata

import (
	"bufio"
	"encoding/json"
	"io"
)

type OutputOrError struct {
	O   *Output
	Err error
}

func Decode(r io.Reader, decoder func(io.Reader, *Output) error) chan OutputOrError {
	ch := make(chan OutputOrError)
	go func() {
		for {
			var o Output
			if err := decoder(r, &o); err == io.EOF {
				close(ch)
				return
			} else if err != nil {
				ch <- OutputOrError{nil, err}
			}
			ch <- OutputOrError{&o, nil}
		}
	}()
	return ch
}

func DecodeJSON(r io.Reader) chan OutputOrError {
	return Decode(r, func(r io.Reader, o *Output) error {
		return json.NewDecoder(r).Decode(&o)
	})
}

func DecodeText(r io.Reader) chan OutputOrError {
	return Decode(r, func(r io.Reader, o *Output) error {
		return parseSMARTCtl(bufio.NewReader(r), o)
	})
}
