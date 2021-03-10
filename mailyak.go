package mailyak

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"regexp"
	"sync/atomic"
	"time"
)

// MailYak is an easy-to-use email builder.
type MailYak struct {
	auth   atomic.Value
	sender emailSender
}

// New returns an instance of MailYak using host as the SMTP server, and
// authenticating with auth if non-nil.
//
// host must include the port number (i.e. "smtp.itsallbroken.com:25")
//
//      mail := mailyak.New("smtp.itsallbroken.com:25", smtp.PlainAuth(
//          "",
//          "username",
//          "password",
//          "smtp.itsallbroken.com",
//      ))
//
// MailYak instances created with New will switch to using TLS after connecting
// if the remote host supports the STARTTLS command. For an explicit TLS
// connection, or to provide a custom tls.Config, use NewWithTLS() instead.
func New(host string, auth smtp.Auth) *MailYak {
	m := &MailYak{
		sender: newSenderWithStartTLS(host),
	}
	if auth != nil {
		m.auth.Store(auth)
	}
	return m
}

var trimRegex = regexp.MustCompile("\r?\n")

// NewWithTLS returns an instance of MailYak using host as the SMTP server over
// an explicit TLS connection, and authenticating with auth if non-nil.
//
// host must include the port number (i.e. "smtp.itsallbroken.com:25")
//
//      mail := mailyak.NewWithTLS("smtp.itsallbroken.com:25", smtp.PlainAuth(
//          "",
//          "username",
//          "password",
//          "smtp.itsallbroken.com",
//      ), tlsConfig)
//
// If tlsConfig is nil, a sensible default is generated that can connect to
// host.
func NewWithTLS(host string, auth smtp.Auth, tlsConfig *tls.Config) (*MailYak, error) {
	// Construct a default MailYak instance
	m := New(host, auth)

	// Initialise the TLS sender with the (potentially nil) TLS config, swapping
	// it with the default STARTTLS sender.
	var err error
	m.sender, err = newSenderWithExplicitTLS(host, tlsConfig)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// NewMail returns new Mail from pool
//
// Mail used for concurrent mail calls Send
func (m *MailYak) NewMail() *Mail {
	mail := getMail()
	mail.date = time.Now().Format(mailDateFormat)
	mail.auth = m.auth.Load().(smtp.Auth)
	return mail
}

// Send attempts to send the built email via the configured SMTP server.
//
// Attachments are read and the email timestamp is created when Send() is
// called, and any connection/authentication errors will be returned by Send().
func (m *MailYak) Send(mail *Mail) error {
	defer putMail(mail)
	mail.date = time.Now().Format(mailDateFormat)
	return m.sender.Send(mail)
}

// String returns a redacted description of the email state, typically for
// logging or debugging purposes.
//
// Authentication information is not included in the returned string.
func (m *MailYak) String() string {
	_, isTLSSender := m.sender.(*senderExplicitTLS)

	return fmt.Sprintf(
		"&MailYak{auth set: %v, explicit tls: %v}",
		m.auth.Load().(smtp.Auth) != nil,
		isTLSSender,
	)
}
