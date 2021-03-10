package mailyak

import (
	"github.com/valyala/bytebufferpool"
)

// BodyPart is a buffer holding the contents of an email MIME part.
type BodyPart struct{ *bytebufferpool.ByteBuffer }

// Set accepts a string s as the contents of a BodyPart, replacing any existing
// data.
func (w *BodyPart) Set(s string) {
	w.Reset()
	w.WriteString(s)
}
func GetBodyPart() *BodyPart {
	return &BodyPart{ByteBuffer: bytebufferpool.Get()}
}

func PutBodyPart(bp *BodyPart) {
	bytebufferpool.Put(bp.ByteBuffer)
	bp.ByteBuffer = nil
}
