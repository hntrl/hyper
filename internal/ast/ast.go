package ast

import (
	"fmt"

	"github.com/hntrl/hyper/internal/tokens"
)

type Node interface {
	Validate() error
	Pos() tokens.Position
}

func ExpectedError(pos tokens.Position, expected tokens.Token, lit string) error {
	return fmt.Errorf("syntax (%s): expected %s but got %s", pos.String(), expected.String(), lit)
}
