// This package provides a easy to use MIME email composer with support for
// attachments.
package mailyak

import (
	"bytes"
	"crypto/tls"
	"io"
	"net/smtp"
)

func Example() {
	// Create a new email - specify the SMTP host:port and auth (or nil if not
	// needed).
	//
	// If you want to connect using TLS, use NewWithTLS() instead.
	my := New("mail.host.com:25", smtp.PlainAuth("", "user", "pass", "mail.host.com"))
	mail := my.NewMail()
	mail.To("dom@itsallbroken.com")
	mail.From("jsmith@example.com")
	mail.FromName("Prince Anybody")

	mail.Subject("Business proposition")

	// Add a custom header
	mail.AddHeader("X-TOTALLY-NOT-A-SCAM", "true")

	// mail.HTMLWriter() and mail.PlainWriter() implement io.Writer, so you can
	// do handy things like parse a template directly into the email body - here
	// we just use io.WriteString()
	if _, err := io.WriteString(mail.HTML(), "So long, and thanks for all the fish."); err != nil {
		panic(" :( ")
	}

	// Or set the body using a string helper
	mail.Plain().SetString("Get a real email client")

	// And you're done!
	if err := my.Send(mail); err != nil {
		panic(" :( ")
	}
}

func Example_attachments() {
	// This will be our attachment data
	buf := &bytes.Buffer{}
	_, _ = io.WriteString(buf, "We're in the stickiest situation since Sticky the Stick Insect got stuck on a sticky bun.")

	// Create a new email - specify the SMTP host:port and auth (or nil if not
	// needed).
	my := New("mail.host.com:25", smtp.PlainAuth("", "user", "pass", "mail.host.com"))
	mail := my.NewMail()

	mail.To("dom@itsallbroken.com")
	mail.From("jsmith@example.com")
	mail.HTML().SetString("I am an email")

	// buf could be anything that implements io.Reader, like a file on disk or
	// an in-memory buffer.
	mail.Attach("sticky.txt", buf)

	if err := my.Send(mail); err != nil {
		panic(" :( ")
	}
}

func ExampleNewWithTLS() {
	// Create a new Mail instance that uses an explicit TLS connection. This
	// ensures no communication is performed in plain-text.
	//
	// Specify the SMTP host:port to connect to, the authentication credentials
	// (or nil if not needed), and use an automatically generated TLS
	// configuration by passing nil as the tls.Config argument.
	my, err := NewWithTLS("mail.host.com:25", smtp.PlainAuth("", "user", "pass", "mail.host.com"), nil)
	if err != nil {
		panic("failed to initialise a TLS instance :(")
	}
	mail := my.NewMail()

	mail.Plain().SetString("Have some encrypted goodness")
	if err := my.Send(mail); err != nil {
		panic(" :( ")
	}
}

func ExampleNewWithTLS_with_config() {
	// Create a new Mail instance that uses an explicit TLS connection. This
	// ensures no communication is performed in plain-text.
	//
	// Specify the SMTP host:port to connect to, the authentication credentials
	// (or nil if not needed), and use the tls.Config provided.
	my, err := NewWithTLS(
		"mail.host.com:25",
		smtp.PlainAuth("", "user", "pass", "mail.host.com"),
		&tls.Config{
			// ServerName is used to verify the hostname on the returned
			// certificates unless InsecureSkipVerify is given. It is also included
			// in the client's handshake to support virtual hosting unless it is
			// an IP address.
			ServerName: "mail.host.com",

			// Negotiate a connection that uses at least TLS v1.2, or refuse the
			// connection if the server does not support it. Most do, and it is
			// a very good idea to enforce it!
			MinVersion: tls.VersionTLS12,
		},
	)
	if err != nil {
		panic("failed to initialise a TLS instance :(")
	}
	mail := my.NewMail()

	mail.Plain().SetString("Have some encrypted goodness")
	if err := my.Send(mail); err != nil {
		panic(" :( ")
	}
}

func ExampleMail_AttachInline() {
	// Create a new email
	my := New("mail.host.com:25", smtp.PlainAuth("", "user", "pass", "mail.host.com"))
	mail := my.NewMail()
	mail.To("dom@itsallbroken.com")
	mail.From("jsmith@example.com")

	// Initialise an io.Reader that contains your image (typically read from
	// disk, or embedded in memory).
	//
	// Here we use an empty buffer as a mock.
	imageBuffer := &bytes.Buffer{}

	// Add the image as an attachment.
	//
	// To reference it, use the name as the cid value.
	mail.AttachInline("myimage", imageBuffer)

	// Set the HTML body, which includes the inline CID reference.
	mail.HTML().SetString(`
		<html>
		<body>
			<img src="cid:myimage"/>
		</body>
		</html>
	`)

	// Send it!
	if err := my.Send(mail); err != nil {
		panic(" :( ")
	}
}
