package tokens

import "fmt"

type Position struct {
	Offset int
	Line   int
	Column int
}

func (p Position) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}

func (p Position) Difference(other Position) Position {
	p.Column = other.Column - p.Column
	return p
}

func (p Position) MarshalText() ([]byte, error) {
	return []byte{}, nil
}
