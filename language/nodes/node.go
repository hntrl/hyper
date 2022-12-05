package nodes

import (
	"fmt"

	"github.com/hntrl/lang/language/tokens"
)

type Node interface {
	Validate() error
	Pos() tokens.Position
}

// type SymbolParser interface {
// 	Scan() (tokens.Position, tokens.Token, string)
// 	ScanIgnore(...tokens.Token) (tokens.Position, tokens.Token, string)
// 	Unscan()
// 	Rollback(int)
// 	Index() int
// }

func ExpectedError(pos tokens.Position, expected tokens.Token, lit string) error {
	return fmt.Errorf("syntax (%s): expected %s but got %s", pos.String(), expected.String(), lit)
}
