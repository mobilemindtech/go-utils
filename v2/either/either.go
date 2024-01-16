package either

import (
	"github.com/mobilemindtec/go-utils/v2/optional"
)

type Either[L any, R any] struct {
	left  L
	right R
}

func Left[L any, R any](v L) *Either[L, R] {
	return &Either[L, R]{left: v}
}
func Right[L any, R any](v R) *Either[L, R] {
	return &Either[L, R]{right: v}
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
	return !optional.IsNilFixed(this.left)
}

func (this Either[L, R]) IsRight() bool {
	return !optional.IsNilFixed(this.right)
}


// MapIf map either to another either
func Map[L any, R any, LN any, RN any](e *Either[L, R], f func(*Either[L, R]) *Either[LN, RN]) *Either[LN, RN] {
	return f(e)
}

// MapIf map either to another either by conditional left or rigth
func MapIf[L any, R any, LN any, RN any](e *Either[L, R], fl func(*Either[L, R]) LN, fr func(*Either[L, R]) RN) *Either[LN, RN] {
	if e.IsLeft() {
		return Left[LN, RN](fl(e))
	}
	return Right[LN, RN](fr(e))
}