package mailyak

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"strings"
	"testing"
	"time"
)

// TestMailFromHeader ensures the fromHeader method returns valid headers
func TestMailFromHeader(t *testing.T) {
	t.Parallel()

	tests := []struct {
		// Test description.
		name string
		// Receiver fields.
		rfromAddr string
		rfromName string
		// Expected results.
		want string
	}{
		{
			"With name",
			"dom@itsallbroken.com",
			"Dom",
			"From: Dom <dom@itsallbroken.com>\r\n",
		},
		{
			"Without name",
			"dom@itsallbroken.com",
			"",
			"From: dom@itsallbroken.com\r\n",
		},
		{
			"Without either",
			"",
			"",
			"From: \r\n",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := Mail{
				fromAddr: tt.rfromAddr,
				fromName: tt.rfromName,
			}

			if got := m.fromHeader(); got != tt.want {
				t.Errorf("%q. Mail.fromHeader() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

// TestMailWriteHeaders ensures the Mime-Version, Date, Reply-To, From, To and
// Subject headers are correctly wrote
func TestMailWriteHeaders(t *testing.T) {
	t.Parallel()

	now := time.Now().Format(time.RFC1123Z)
	tests := []struct {
		// Test description.
		name string
		// Receiver fields.
		rtoAddrs        []string
		rccAddrs        []string
		rbccAddrs       []string
		rsubject        string
		rreplyTo        string
		rwriteBccHeader bool
		// Expected results.
		wantBuf string
	}{
		{
			"All fields",
			[]string{"test@itsallbroken.com"},
			[]string{},
			[]string{},
			"Test",
			"help@itsallbroken.com",
			true,
			"From: Dom <dom@itsallbroken.com>\r\nMime-Version: 1.0\r\nDate: " + now + "\r\nReply-To: help@itsallbroken.com\r\nSubject: Test\r\nTo: test@itsallbroken.com\r\n",
		},
		{
			"No reply-to",
			[]string{"test@itsallbroken.com"},
			[]string{},
			[]string{},
			"",
			"",
			true,
			"From: Dom <dom@itsallbroken.com>\r\nMime-Version: 1.0\r\nDate: " + now + "\r\nSubject: \r\nTo: test@itsallbroken.com\r\n",
		},
		{
			"Multiple To addresses",
			[]string{"test@itsallbroken.com", "repairs@itsallbroken.com"},
			[]string{},
			[]string{},
			"",
			"",
			true,
			"From: Dom <dom@itsallbroken.com>\r\nMime-Version: 1.0\r\nDate: " + now + "\r\nSubject: \r\nTo: test@itsallbroken.com\r\nTo: repairs@itsallbroken.com\r\n",
		},
		{
			"Single Cc address, Multiple To addresses",
			[]string{"test@itsallbroken.com", "repairs@itsallbroken.com"},
			[]string{"cc@itsallbroken.com"},
			[]string{},
			"",
			"",
			true,
			"From: Dom <dom@itsallbroken.com>\r\nMime-Version: 1.0\r\nDate: " + now + "\r\nSubject: \r\nTo: test@itsallbroken.com\r\nTo: repairs@itsallbroken.com\r\nCC: cc@itsallbroken.com\r\n",
		},
		{
			"Multiple Cc addresses, Multiple To addresses",
			[]string{"test@itsallbroken.com", "repairs@itsallbroken.com"},
			[]string{"cc1@itsallbroken.com", "cc2@itsallbroken.com"},
			[]string{},
			"",
			"",
			true,
			"From: Dom <dom@itsallbroken.com>\r\nMime-Version: 1.0\r\nDate: " + now + "\r\nSubject: \r\nTo: test@itsallbroken.com\r\nTo: repairs@itsallbroken.com\r\nCC: cc1@itsallbroken.com\r\nCC: cc2@itsallbroken.com\r\n",
		},
		{
			"Single Bcc address, Multiple To addresses",
			[]string{"test@itsallbroken.com", "repairs@itsallbroken.com"},
			[]string{},
			[]string{"bcc@itsallbroken.com"},
			"",
			"",
			true,
			"From: Dom <dom@itsallbroken.com>\r\nMime-Version: 1.0\r\nDate: " + now + "\r\nSubject: \r\nTo: test@itsallbroken.com\r\nTo: repairs@itsallbroken.com\r\nBCC: bcc@itsallbroken.com\r\n",
		},
		{
			"Multiple Bcc addresses, Multiple To addresses",
			[]string{"test@itsallbroken.com", "repairs@itsallbroken.com"},
			[]string{},
			[]string{"bcc1@itsallbroken.com", "bcc2@itsallbroken.com"},
			"",
			"",
			true,
			"From: Dom <dom@itsallbroken.com>\r\nMime-Version: 1.0\r\nDate: " + now + "\r\nSubject: \r\nTo: test@itsallbroken.com\r\nTo: repairs@itsallbroken.com\r\nBCC: bcc1@itsallbroken.com\r\nBCC: bcc2@itsallbroken.com\r\n",
		},
		{
			"Multiple Bcc addresses, Multiple To addresses",
			[]string{"test@itsallbroken.com", "repairs@itsallbroken.com"},
			[]string{},
			[]string{"bcc1@itsallbroken.com", "bcc2@itsallbroken.com"},
			"",
			"",
			false,
			"From: Dom <dom@itsallbroken.com>\r\nMime-Version: 1.0\r\nDate: " + now + "\r\nSubject: \r\nTo: test@itsallbroken.com\r\nTo: repairs@itsallbroken.com\r\n",
		},
		{
			"All together now",
			[]string{"test@itsallbroken.com", "repairs@itsallbroken.com"},
			[]string{"cc1@itsallbroken.com", "cc2@itsallbroken.com"},
			[]string{"bcc1@itsallbroken.com", "bcc2@itsallbroken.com"},
			"",
			"",
			true,
			"From: Dom <dom@itsallbroken.com>\r\nMime-Version: 1.0\r\nDate: " + now + "\r\nSubject: \r\nTo: test@itsallbroken.com\r\nTo: repairs@itsallbroken.com\r\nCC: cc1@itsallbroken.com\r\nCC: cc2@itsallbroken.com\r\nBCC: bcc1@itsallbroken.com\r\nBCC: bcc2@itsallbroken.com\r\n",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := Mail{
				toAddrs:        tt.rtoAddrs,
				subject:        tt.rsubject,
				fromAddr:       "dom@itsallbroken.com",
				fromName:       "Dom",
				replyTo:        tt.rreplyTo,
				ccAddrs:        tt.rccAddrs,
				bccAddrs:       tt.rbccAddrs,
				writeBccHeader: tt.rwriteBccHeader,
				date:           now,
			}

			buf := &bytes.Buffer{}
			if err := m.writeHeaders(buf); err != nil {
				t.Fatal(err)
			}

			if gotBuf := buf.String(); gotBuf != tt.wantBuf {
				t.Errorf("%q. Mail.writeHeaders() = %v, want %v", tt.name, gotBuf, tt.wantBuf)
			}
		})
	}
}

// TestMailWriteBody ensures the correct MIME parts are wrote for the body
func TestMailWriteBody(t *testing.T) {
	t.Parallel()

	tests := []struct {
		// Test description.
		name string
		// Receiver fields.
		rHTML  string
		rPlain string
		// Parameters.
		boundary string
		// Expected results.
		wantW   string
		wantErr bool
	}{
		{
			"Empty",
			"",
			"",
			"test",
			"--test\r\nContent-Transfer-Encoding: quoted-printable\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n\r\n--test\r\nContent-Transfer-Encoding: quoted-printable\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n\r\n--test--\r\n",
			false,
		},
		{
			"HTML",
			"HTML",
			"",
			"t",
			"--t\r\nContent-Transfer-Encoding: quoted-printable\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n\r\n--t\r\nContent-Transfer-Encoding: quoted-printable\r\nContent-Type: text/html; charset=UTF-8\r\n\r\nHTML\r\n--t--\r\n",
			false,
		},
		{
			"Plain text",
			"",
			"Plain",
			"t",
			"--t\r\nContent-Transfer-Encoding: quoted-printable\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\nPlain\r\n--t\r\nContent-Transfer-Encoding: quoted-printable\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n\r\n--t--\r\n",
			false,
		},
		{
			"Both",
			"HTML",
			"Plain",
			"t",
			"--t\r\nContent-Transfer-Encoding: quoted-printable\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\nPlain\r\n--t\r\nContent-Transfer-Encoding: quoted-printable\r\nContent-Type: text/html; charset=UTF-8\r\n\r\nHTML\r\n--t--\r\n",
			false,
		},
		{
			"Both with long lines",
			"Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.",
			"Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.",
			"t",
			"--t\r\nContent-Transfer-Encoding: quoted-printable\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\nLorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tem=\r\npor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, q=\r\nuis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo cons=\r\nequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillu=\r\nm dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non pr=\r\noident, sunt in culpa qui officia deserunt mollit anim id est laborum.\r\n--t\r\nContent-Transfer-Encoding: quoted-printable\r\nContent-Type: text/html; charset=UTF-8\r\n\r\nLorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tem=\r\npor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, q=\r\nuis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo cons=\r\nequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillu=\r\nm dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non pr=\r\noident, sunt in culpa qui officia deserunt mollit anim id est laborum.\r\n--t--\r\n",
			false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := getMail()
			_, _ = m.HTML().WriteString(tt.rHTML)
			_, _ = m.Plain().WriteString(tt.rPlain)

			w := &bytes.Buffer{}
			if err := m.writeBody(w, tt.boundary); (err != nil) != tt.wantErr {
				t.Fatalf("%q. Mail.writeBody() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}

			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("%q. Mail.writeBody() = %v, want %v", tt.name, gotW, tt.wantW)
			}
			putMail(m)
		})
	}
}

// TestMailBuildMime tests all the other mime-related bits combine in a sane way
func TestMailBuildMime(t *testing.T) {
	t.Parallel()

	now := time.Now().Format(time.RFC1123Z)
	tests := []struct {
		// Test description.
		name string
		// Receiver fields.
		rHTML     []byte
		rPlain    []byte
		rtoAddrs  []string
		rsubject  string
		rfromAddr string
		rfromName string
		rreplyTo  string
		// Expected results.
		want    string
		wantErr bool
	}{
		{
			"Empty",
			[]byte{},
			[]byte{},
			[]string{""},
			"",
			"",
			"",
			"",
			"From: \r\nMime-Version: 1.0\r\nDate: " + now + "\r\nSubject: \r\nTo: \r\nContent-Type: multipart/mixed;\r\n\tboundary=\"mixed\"; charset=UTF-8\r\n\r\n--mixed\r\nContent-Type: multipart/alternative;\r\n\tboundary=\"alt\"\r\n\r\n--alt\r\nContent-Transfer-Encoding: quoted-printable\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n\r\n--alt\r\nContent-Transfer-Encoding: quoted-printable\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n\r\n--alt--\r\n\r\n--mixed--\r\n",
			false,
		},
		{
			"HTML",
			[]byte("HTML"),
			[]byte{},
			[]string{""},
			"",
			"",
			"",
			"",
			"From: \r\nMime-Version: 1.0\r\nDate: " + now + "\r\nSubject: \r\nTo: \r\nContent-Type: multipart/mixed;\r\n\tboundary=\"mixed\"; charset=UTF-8\r\n\r\n--mixed\r\nContent-Type: multipart/alternative;\r\n\tboundary=\"alt\"\r\n\r\n--alt\r\nContent-Transfer-Encoding: quoted-printable\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n\r\n--alt\r\nContent-Transfer-Encoding: quoted-printable\r\nContent-Type: text/html; charset=UTF-8\r\n\r\nHTML\r\n--alt--\r\n\r\n--mixed--\r\n",
			false,
		},
		{
			"Plain",
			[]byte{},
			[]byte("Plain"),
			[]string{""},
			"",
			"",
			"",
			"",
			"From: \r\nMime-Version: 1.0\r\nDate: " + now + "\r\nSubject: \r\nTo: \r\nContent-Type: multipart/mixed;\r\n\tboundary=\"mixed\"; charset=UTF-8\r\n\r\n--mixed\r\nContent-Type: multipart/alternative;\r\n\tboundary=\"alt\"\r\n\r\n--alt\r\nContent-Transfer-Encoding: quoted-printable\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\nPlain\r\n--alt\r\nContent-Transfer-Encoding: quoted-printable\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n\r\n--alt--\r\n\r\n--mixed--\r\n",
			false,
		},
		{
			"Reply-To",
			[]byte{},
			[]byte{},
			[]string{""},
			"",
			"",
			"",
			"reply",
			"From: \r\nMime-Version: 1.0\r\nDate: " + now + "\r\nReply-To: reply\r\nSubject: \r\nTo: \r\nContent-Type: multipart/mixed;\r\n\tboundary=\"mixed\"; charset=UTF-8\r\n\r\n--mixed\r\nContent-Type: multipart/alternative;\r\n\tboundary=\"alt\"\r\n\r\n--alt\r\nContent-Transfer-Encoding: quoted-printable\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n\r\n--alt\r\nContent-Transfer-Encoding: quoted-printable\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n\r\n--alt--\r\n\r\n--mixed--\r\n",
			false,
		},
		{
			"From name",
			[]byte{},
			[]byte{},
			[]string{""},
			"",
			"",
			"name",
			"",
			"From: name <>\r\nMime-Version: 1.0\r\nDate: " + now + "\r\nSubject: \r\nTo: \r\nContent-Type: multipart/mixed;\r\n\tboundary=\"mixed\"; charset=UTF-8\r\n\r\n--mixed\r\nContent-Type: multipart/alternative;\r\n\tboundary=\"alt\"\r\n\r\n--alt\r\nContent-Transfer-Encoding: quoted-printable\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n\r\n--alt\r\nContent-Transfer-Encoding: quoted-printable\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n\r\n--alt--\r\n\r\n--mixed--\r\n",
			false,
		},
		{
			"From name + address",
			[]byte{},
			[]byte{},
			[]string{""},
			"",
			"addr",
			"name",
			"",
			"From: name <addr>\r\nMime-Version: 1.0\r\nDate: " + now + "\r\nSubject: \r\nTo: \r\nContent-Type: multipart/mixed;\r\n\tboundary=\"mixed\"; charset=UTF-8\r\n\r\n--mixed\r\nContent-Type: multipart/alternative;\r\n\tboundary=\"alt\"\r\n\r\n--alt\r\nContent-Transfer-Encoding: quoted-printable\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n\r\n--alt\r\nContent-Transfer-Encoding: quoted-printable\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n\r\n--alt--\r\n\r\n--mixed--\r\n",
			false,
		},
		{
			"From",
			[]byte{},
			[]byte{},
			[]string{""},
			"",
			"from",
			"",
			"",
			"From: from\r\nMime-Version: 1.0\r\nDate: " + now + "\r\nSubject: \r\nTo: \r\nContent-Type: multipart/mixed;\r\n\tboundary=\"mixed\"; charset=UTF-8\r\n\r\n--mixed\r\nContent-Type: multipart/alternative;\r\n\tboundary=\"alt\"\r\n\r\n--alt\r\nContent-Transfer-Encoding: quoted-printable\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n\r\n--alt\r\nContent-Transfer-Encoding: quoted-printable\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n\r\n--alt--\r\n\r\n--mixed--\r\n",
			false,
		},
		{
			"Subject",
			[]byte{},
			[]byte{},
			[]string{""},
			"subject",
			"",
			"",
			"",
			"From: \r\nMime-Version: 1.0\r\nDate: " + now + "\r\nSubject: subject\r\nTo: \r\nContent-Type: multipart/mixed;\r\n\tboundary=\"mixed\"; charset=UTF-8\r\n\r\n--mixed\r\nContent-Type: multipart/alternative;\r\n\tboundary=\"alt\"\r\n\r\n--alt\r\nContent-Transfer-Encoding: quoted-printable\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n\r\n--alt\r\nContent-Transfer-Encoding: quoted-printable\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n\r\n--alt--\r\n\r\n--mixed--\r\n",
			false,
		},
		{
			"To addresses",
			[]byte{},
			[]byte{},
			[]string{"one", "two"},
			"",
			"",
			"",
			"",
			"From: \r\nMime-Version: 1.0\r\nDate: " + now + "\r\nSubject: \r\nTo: one\r\nTo: two\r\nContent-Type: multipart/mixed;\r\n\tboundary=\"mixed\"; charset=UTF-8\r\n\r\n--mixed\r\nContent-Type: multipart/alternative;\r\n\tboundary=\"alt\"\r\n\r\n--alt\r\nContent-Transfer-Encoding: quoted-printable\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n\r\n--alt\r\nContent-Transfer-Encoding: quoted-printable\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n\r\n--alt--\r\n\r\n--mixed--\r\n",
			false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := getMail()
			m.toAddrs = tt.rtoAddrs
			m.subject = tt.rsubject
			m.fromAddr = tt.rfromAddr
			m.fromName = tt.rfromName
			m.replyTo = tt.rreplyTo
			m.date = now

			_, _ = m.HTML().Write(tt.rHTML)
			_, _ = m.Plain().Write(tt.rPlain)

			buf := &bytes.Buffer{}
			err := m.buildMimeWithBoundaries(buf, "mixed", "alt")
			if (err != nil) != tt.wantErr {
				t.Fatalf("%q. Mail.buildMime() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}

			if bytes.Equal(buf.Bytes(), []byte(tt.want)) {
				t.Errorf("%q. Mail.buildMime() = %v, want %v", tt.name, buf.Bytes(), []byte(tt.want))
			}
		})
	}
}

// TestMailBuildMime_withAttachments ensures attachments are correctly added to the MIME message
func TestMailBuildMime_withAttachments(t *testing.T) {
	t.Parallel()

	tests := []struct {
		// Test description.
		name string
		// Receiver fields.
		rHTML        []byte
		rPlain       []byte
		rtoAddrs     []string
		rsubject     string
		rfromAddr    string
		rfromName    string
		rreplyTo     string
		rattachments []attachment
		// Expected results.
		wantAttach []string
		wantErr    bool
	}{
		{
			"No attachment",
			[]byte{},
			[]byte{},
			[]string{""},
			"",
			"",
			"",
			"",
			[]attachment{},
			[]string{},
			false,
		},
		{
			"One attachment",
			[]byte{},
			[]byte{},
			[]string{""},
			"",
			"",
			"",
			"",
			[]attachment{
				{"test.txt", strings.NewReader("content"), false, false, ""},
			},
			[]string{"Y29udGVudA=="},
			false,
		},
		{
			"One attachment raw",
			[]byte{},
			[]byte{},
			[]string{""},
			"",
			"",
			"",
			"",
			[]attachment{
				{"test.pdf", strings.NewReader("JVBERi0xLjcKCjEgMCBvYmogICUgZW50cnkgcG9pbnQKPDwKICAvVHlwZSAvQ2F0YWxvZwogIC9QYWdlcyAyIDAgUgo+PgplbmRvYmoKCjIgMCBvYmoKPDwKICAvVHlwZSAvUGFnZXMKICAvTWVkaWFCb3ggWyAwIDAgMjAwIDIwMCBdCiAgL0NvdW50IDEKICAvS2lkcyBbIDMgMCBSIF0KPj4KZW5kb2JqCgozIDAgb2JqCjw8CiAgL1R5cGUgL1BhZ2UKICAvUGFyZW50IDIgMCBSCiAgL1Jlc291cmNlcyA8PAogICAgL0ZvbnQgPDwKICAgICAgL0YxIDQgMCBSIAogICAgPj4KICA+PgogIC9Db250ZW50cyA1IDAgUgo+PgplbmRvYmoKCjQgMCBvYmoKPDwKICAvVHlwZSAvRm9udAogIC9TdWJ0eXBlIC9UeXBlMQogIC9CYXNlRm9udCAvVGltZXMtUm9tYW4KPj4KZW5kb2JqCgo1IDAgb2JqICAlIHBhZ2UgY29udGVudAo8PAogIC9MZW5ndGggNDQKPj4Kc3RyZWFtCkJUCjcwIDUwIFRECi9GMSAxMiBUZgooSGVsbG8sIHdvcmxkISkgVGoKRVQKZW5kc3RyZWFtCmVuZG9iagoKeHJlZgowIDYKMDAwMDAwMDAwMCA2NTUzNSBmIAowMDAwMDAwMDEwIDAwMDAwIG4gCjAwMDAwMDAwNzkgMDAwMDAgbiAKMDAwMDAwMDE3MyAwMDAwMCBuIAowMDAwMDAwMzAxIDAwMDAwIG4gCjAwMDAwMDAzODAgMDAwMDAgbiAKdHJhaWxlcgo8PAogIC9TaXplIDYKICAvUm9vdCAxIDAgUgo+PgpzdGFydHhyZWYKNDkyCiUlRU9G"), false, true, "application/pdf"},
			},
			[]string{"JVBERi0xLjcKCjEgMCBvYmogICUgZW50cnkgcG9pbnQKPDwKICAvVHlwZSAv\r\nQ2F0YWxvZwogIC9QYWdlcyAyIDAgUgo+PgplbmRvYmoKCjIgMCBvYmoKPDwK\r\nICAvVHlwZSAvUGFnZXMKICAvTWVkaWFCb3ggWyAwIDAgMjAwIDIwMCBdCiAg\r\nL0NvdW50IDEKICAvS2lkcyBbIDMgMCBSIF0KPj4KZW5kb2JqCgozIDAgb2Jq\r\nCjw8CiAgL1R5cGUgL1BhZ2UKICAvUGFyZW50IDIgMCBSCiAgL1Jlc291cmNl\r\ncyA8PAogICAgL0ZvbnQgPDwKICAgICAgL0YxIDQgMCBSIAogICAgPj4KICA+\r\nPgogIC9Db250ZW50cyA1IDAgUgo+PgplbmRvYmoKCjQgMCBvYmoKPDwKICAv\r\nVHlwZSAvRm9udAogIC9TdWJ0eXBlIC9UeXBlMQogIC9CYXNlRm9udCAvVGlt\r\nZXMtUm9tYW4KPj4KZW5kb2JqCgo1IDAgb2JqICAlIHBhZ2UgY29udGVudAo8\r\nPAogIC9MZW5ndGggNDQKPj4Kc3RyZWFtCkJUCjcwIDUwIFRECi9GMSAxMiBU\r\nZgooSGVsbG8sIHdvcmxkISkgVGoKRVQKZW5kc3RyZWFtCmVuZG9iagoKeHJl\r\nZgowIDYKMDAwMDAwMDAwMCA2NTUzNSBmIAowMDAwMDAwMDEwIDAwMDAwIG4g\r\nCjAwMDAwMDAwNzkgMDAwMDAgbiAKMDAwMDAwMDE3MyAwMDAwMCBuIAowMDAw\r\nMDAwMzAxIDAwMDAwIG4gCjAwMDAwMDAzODAgMDAwMDAgbiAKdHJhaWxlcgo8\r\nPAogIC9TaXplIDYKICAvUm9vdCAxIDAgUgo+PgpzdGFydHhyZWYKNDkyCiUl\r\nRU9G"},
			false,
		},
		{
			"One inline attachment",
			[]byte{},
			[]byte{},
			[]string{""},
			"",
			"",
			"",
			"",
			[]attachment{
				{"test.txt", strings.NewReader("content"), true, false, ""},
			},
			[]string{"Y29udGVudA=="},
			false,
		},
		{
			"Two attachments",
			[]byte{},
			[]byte{},
			[]string{""},
			"",
			"",
			"",
			"",
			[]attachment{
				{"test.txt", strings.NewReader("content"), false, false, ""},
				{"another.txt", strings.NewReader("another"), false, false, ""},
			},
			[]string{"Y29udGVudA==", "YW5vdGhlcg=="},
			false,
		},
		{
			"Two inline attachments",
			[]byte{},
			[]byte{},
			[]string{""},
			"",
			"",
			"",
			"",
			[]attachment{
				{"test.txt", strings.NewReader("content"), true, false, ""},
				{"another.txt", strings.NewReader("another"), true, false, ""},
			},
			[]string{"Y29udGVudA==", "YW5vdGhlcg=="},
			false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := getMail()
			m.toAddrs = tt.rtoAddrs
			m.subject = tt.rsubject
			m.fromAddr = tt.rfromAddr
			m.fromName = tt.rfromName
			m.replyTo = tt.rreplyTo
			m.attachments = tt.rattachments

			_, _ = m.HTML().Write(tt.rHTML)
			_, _ = m.Plain().Write(tt.rPlain)

			buf := &bytes.Buffer{}
			err := m.buildMimeWithBoundaries(buf, "mixed", "alt")
			if (err != nil) != tt.wantErr {
				t.Fatalf("%q. Mail.buildMime() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}

			seen := 0
			mr := multipart.NewReader(buf, "mixed")

			// Itterate over the mime parts, look for attachments
			for {
				p, err := mr.NextPart()
				if err == io.EOF {
					break
				}
				if err != nil {
					t.Errorf("%q. Mail.buildMime() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				}

				// Read the attachment data
				slurp, err := ioutil.ReadAll(p)
				if err != nil {
					t.Errorf("%q. Mail.buildMime() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				}

				// Skip non-attachments
				if p.Header.Get("Content-Disposition") == "" {
					continue
				}

				// Run through our attachments looking for a match
				for i, attch := range tt.rattachments {
					// Check Disposition header
					var disp string
					if attch.inline {
						disp = "inline; filename=%q"
					} else {
						disp = "attachment; filename=%q"
					}
					if p.Header.Get("Content-Disposition") != fmt.Sprintf(disp, attch.filename) {
						continue
					}

					// Check data
					if !bytes.Equal(slurp, []byte(tt.wantAttach[i])) {
						fmt.Printf("Part %q: %q\n", p.Header.Get("Content-Disposition"), slurp)
						continue
					}

					seen++
				}

			}

			// Did we see all the expected attachments?
			if seen != len(tt.rattachments) {
				t.Errorf("%q. Mail.buildMime() didn't find all attachments in mime body", tt.name)
			}
		})
	}
}
