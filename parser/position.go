package parser

import "fmt"

type Position struct {
	StartedAt int
	EndedAt   int
}

func (p Position) String() string {
	return fmt.Sprintf("%v< ... <= %v", p.StartedAt, p.EndedAt)
}

func NewPosition(s, e int) *Position {
	return &Position{StartedAt: s, EndedAt: e}
}
