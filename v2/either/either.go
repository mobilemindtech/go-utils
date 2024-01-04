package either

import (
	"github.com/mobilemindtec/go-utils/v2/optional"
)

type Either[L any, R any] struct {
	left  L
	right R
}

func (this Either[L, R]) Left() *optional.Optional[L] {
	return optional.Of[L](this.left)
}

func (this Either[L, R]) UnwrapLeft() L {
	return this.left
}

func (this Either[L, R]) Right() *optional.Optional[R] {
	return optional.Of[R](this.right)
}

func (this Either[L, R]) UnwrapRight() R {
	return this.right
}

func (this Either[L, R]) IsLeft() bool {
	return this.Left().NonEmpty()
}

func (this Either[L, R]) IsRight() bool {
	return this.Right().NonEmpty()
}
