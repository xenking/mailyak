package mailyak

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/valyala/bytebufferpool"
	"io"
	"mime/multipart"
	"mime/quotedprintable"
	"net/textproto"
	"strings"
)

func (m *Mail) buildMime(w io.Writer) error {
	mb, err := randomBoundary()
	if err != nil {
		return err
	}

	ab, err := randomBoundary()
	if err != nil {
		return err
	}

	return m.buildMimeWithBoundaries(w, mb, ab)
}

// randomBoundary returns a random hexadecimal string used for separating MIME
// parts.
//
// The returned string must be sufficiently random to prevent malicious users
// from performing content injection attacks.
func randomBoundary() (string, error) {
	buf := make([]byte, 30)
	_, err := io.ReadFull(rand.Reader, buf)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

// buildMimeWithBoundaries creates the MIME message using mb and ab as MIME
// boundaries, and returns the generated MIME data as a buffer.
func (m *Mail) buildMimeWithBoundaries(w io.Writer, mb, ab string) error {
	if err := m.writeHeaders(w); err != nil {
		return err
	}

	// Start our multipart/mixed part
	mixed := multipart.NewWriter(w)
	if err := mixed.SetBoundary(mb); err != nil {
		return err
	}

	// To avoid deferring a mixed.Close(), run the write in a closure and
	// close the mixed after.
	tryWrite := func() error {
		buf := bytebufferpool.Get()
		_, _ = buf.WriteString("Content-Type: multipart/mixed;\r\n\tboundary=\"")
		_, _ = buf.WriteString(mixed.Boundary())
		_, _ = buf.WriteString("\"; charset=UTF-8\r\n\r\n")
		_, _ = w.Write(buf.Bytes())
		bytebufferpool.Put(buf)

		var ctype strings.Builder
		ctype.WriteString("multipart/alternative;\n\tboundary=\"")
		ctype.WriteString(ab)
		ctype.WriteByte('"')

		altPart, err := mixed.CreatePart(textproto.MIMEHeader{"Content-Type": {ctype.String()}})
		if err != nil {
			return err
		}

		if err := m.writeBody(altPart, ab); err != nil {
			return err
		}

		return m.writeAttachments(mixed, lineSplitterBuilder{})
	}

	if err := tryWrite(); err != nil {
		return err
	}

	if err := mixed.Close(); err != nil {
		return err
	}

	return nil
}

// writeHeaders writes the Mime-Version, Date, Reply-To, From, To and Subject headers,
// plus any custom headers set via AddHeader().
//goland:noinspection GoUnhandledErrorResult
func (m *Mail) writeHeaders(w io.Writer) error {
	buf := bytebufferpool.Get()
	buf.WriteString(m.fromHeader())
	buf.WriteString("Mime-Version: 1.0\r\n")

	buf.WriteString("Date: ")
	buf.WriteString(m.date)
	buf.WriteString("\r\n")

	if m.replyTo != "" {
		buf.WriteString("Reply-To: ")
		buf.WriteString(m.replyTo)
		buf.WriteString("\r\n")
	}

	buf.WriteString("Subject: ")
	buf.WriteString(m.subject)
	buf.WriteString("\r\n")

	for _, to := range m.toAddrs {
		buf.WriteString("To: ")
		buf.WriteString(to)
		buf.WriteString("\r\n")
	}

	for _, cc := range m.ccAddrs {
		buf.WriteString("CC: ")
		buf.WriteString(cc)
		buf.WriteString("\r\n")
	}

	if m.writeBccHeader {
		for _, bcc := range m.bccAddrs {
			buf.WriteString("BCC: ")
			buf.WriteString(bcc)
			buf.WriteString("\r\n")
		}
	}

	for k, v := range m.headers {
		buf.WriteString(k)
		buf.WriteString(": ")
		buf.WriteString(v)
		buf.WriteString("\r\n")
	}
	w.Write(buf.Bytes())
	bytebufferpool.Put(buf)
	return nil
}

// fromHeader returns a correctly formatted From header, optionally with a name
// component.
func (m *Mail) fromHeader() string {
	if m.fromName == "" {
		return fmt.Sprintf("From: %s\r\n", m.fromAddr)
	}

	return fmt.Sprintf("From: %s <%s>\r\n", m.fromName, m.fromAddr)
}

// writeBody writes the text/plain and text/html mime parts.
func (m *Mail) writeBody(w io.Writer, boundary string) error {
	if m.plain.Len() == 0 && m.html.Len() == 0 {
		// No body to write - just skip it
		return nil
	}

	alt := multipart.NewWriter(w)

	if err := alt.SetBoundary(boundary); err != nil {
		return err
	}

	var err error
	writePart := func(ctype string, data []byte) {
		if len(data) == 0 || err != nil {
			return
		}

		c := fmt.Sprintf("%s; charset=UTF-8", ctype)

		var part io.Writer
		part, err = alt.CreatePart(textproto.MIMEHeader{"Content-Type": {c}, "Content-Transfer-Encoding": {"quoted-printable"}})
		if err != nil {
			return
		}

		var buf bytes.Buffer
		qpw := quotedprintable.NewWriter(&buf)
		_, _ = qpw.Write(data)
		_ = qpw.Close()

		_, err = part.Write(buf.Bytes())
	}

	writePart("text/plain", m.plain.Bytes())
	writePart("text/html", m.html.Bytes())

	// If closing the alt fails, and there's not already an error set, return the
	// close error.
	if closeErr := alt.Close(); err == nil && closeErr != nil {
		return closeErr
	}

	return err
}
