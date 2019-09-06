
package support

import (
	"unicode"
  "golang.org/x/text/transform"
  "golang.org/x/text/unicode/norm"  
)

/*

	go get  github.com/paulrosania/go-charset/charset
	go get -u golang.org/x/text

*/

type Character struct {
	Transformer transform.Transformer
}

func NewCharacter() *Character{
	return &Character{ Transformer: transform.Chain(norm.NFD, transform.RemoveFunc(isMn), norm.NFC) }
}

func isMn(r rune) bool {
  return unicode.Is(unicode.Mn, r) // Mn: nonspacing marks
}

func (this *Character) Transform(text string) string{
  newText, _, _ := transform.String(this.Transformer, text)  
  return newText
}