package mailyak

import (
	"bytes"
	"fmt"
	"net/smtp"
	"strings"
	"sync"
	"time"

	"github.com/valyala/bytebufferpool"
)

// Email Date timestamp format
const mailDateFormat = time.RFC1123Z

var mailPool = sync.Pool{New: func() interface{} {
	return new(Mail)
}}

func getMail() *Mail {
	m := mailPool.Get().(*Mail)
	m.html = bytebufferpool.Get()
	m.plain = bytebufferpool.Get()
	m.headers = map[string]string{}

	return m

}

func putMail(m *Mail) {
	m.Reset()
	mailPool.Put(m)
}

type Mail struct {
	html  *bytebufferpool.ByteBuffer
	plain *bytebufferpool.ByteBuffer

	auth           smtp.Auth
	headers        map[string]string // arbitrary headers
	attachments    []attachment
	toAddrs        []string
	ccAddrs        []string
	bccAddrs       []string
	subject        string
	fromAddr       string
	fromName       string
	replyTo        string
	date           string
	writeBccHeader bool
}

func (m *Mail) Reset() {
	bytebufferpool.Put(m.html)
	m.html = nil
	bytebufferpool.Put(m.plain)
	m.plain = nil
	m.auth = nil
	m.headers = nil
	m.attachments = nil
	m.toAddrs = nil
	m.ccAddrs = nil
	m.bccAddrs = nil
	m.subject = ""
	m.fromAddr = ""
	m.fromName = ""
	m.replyTo = ""
	m.date = ""
	m.writeBccHeader = false
}

// String returns a redacted description of the email state, typically for
// logging or debugging purposes.
//
// Authentication information is not included in the returned string.
func (m *Mail) String() string {
	var att []string
	for _, a := range m.attachments {
		att = append(att, "{filename: "+a.filename+"}")
	}

	var custom string
	if len(m.headers) > 0 {
		var hdrs []string
		for k, v := range m.headers {
			hdrs = append(hdrs, fmt.Sprintf("%s: %q", k, v))
		}
		custom = strings.Join(hdrs, ", ") + ", "
	}

	return fmt.Sprintf(
		"&Mail{date: %q, from: %q, fromName: %q, html: %v bytes, plain: %v bytes, toAddrs: %v, "+
			"bccAddrs: %v, subject: %q, %vattachments (%v): %v, auth set: %v}",
		m.date,
		m.fromAddr,
		m.fromName,
		len(m.HTML().String()),
		len(m.Plain().String()),
		m.toAddrs,
		m.bccAddrs,
		m.subject,
		custom,
		len(att),
		att,
		m.auth != nil,
	)
}

// MimeBuf returns the buffer containing all the RAW MIME data.
//
// MimeBuf is typically used with an API service such as Amazon SES that does
// not use an SMTP interface.
func (m *Mail) MimeBuf() (*bytes.Buffer, error) {
	m.date = time.Now().Format(mailDateFormat)

	buf := &bytes.Buffer{}
	if err := m.buildMime(buf); err != nil {
		return nil, err
	}

	return buf, nil
}

// HTML returns a BodyPart for the HTML email body.
func (m *Mail) HTML() *bytebufferpool.ByteBuffer {
	return m.html
}

// Plain returns a BodyPart for the plain-text email body.
func (m *Mail) Plain() *bytebufferpool.ByteBuffer {
	return m.plain
}

// getFromAddr should return the address to be used in the MAIL FROM
// command.
func (m *Mail) getFromAddr() string {
	return m.fromAddr
}

// getAuth should return the smtp.Auth if configured, nil if not.
func (m *Mail) getAuth() smtp.Auth {
	return m.auth
}

// getToAddrs should return a slice of email addresses to be added to the
// RCPT TO command.
func (m *Mail) getToAddrs() []string {
	// Pre-allocate the slice to avoid growing it, we already know how big it
	// needs to be.
	addrs := len(m.toAddrs) + len(m.ccAddrs) + len(m.bccAddrs)
	out := make([]string, 0, addrs)

	out = append(out, m.toAddrs...)
	out = append(out, m.ccAddrs...)
	out = append(out, m.bccAddrs...)

	return out
}
