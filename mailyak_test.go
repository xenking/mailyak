package mailyak

import (
	"fmt"
	"net/smtp"
	"strings"
	"testing"
	"time"
)

// TestMailStringer ensures MailYak struct conforms to the Stringer interface.
func TestMailStringer(t *testing.T) {
	t.Parallel()

	my := New("mail.host.com:25", smtp.PlainAuth("", "user", "pass", "mail.host.com"))
	mail := my.NewMail()
	mail.From("from@example.org")
	mail.FromName("From Example")
	mail.To("to@example.org")
	mail.Bcc("bcc1@example.org", "bcc2@example.org")
	mail.Subject("Test subject")
	mail.ReplyTo("replies@example.org")
	mail.HTML().SetString("HTML part: this is just a test.")
	mail.Plain().SetString("Plain text part: this is also just a test.")
	mail.Attach("test.html", strings.NewReader("<html><head></head></html>"))
	mail.Attach("test2.html", strings.NewReader("<html><head></head></html>"))

	mail.AddHeader("Precedence", "bulk")

	mail.date = "a date"

	want := "&Mail{date: \"a date\", from: \"from@example.org\", fromName: \"From Example\", html: 31 bytes, plain: 42 bytes, toAddrs: [to@example.org], bccAddrs: [bcc1@example.org bcc2@example.org], subject: \"Test subject\", Precedence: \"bulk\", attachments (2): [{filename: test.html} {filename: test2.html}], auth set: true}"
	got := fmt.Sprintf("%+v", mail)
	if got != want {
		t.Errorf("Mail.String() = %v, want %v", got, want)
	}
}

// TestMailYakDate ensures two emails sent with the same MailYak instance use
// different (updated) date timestamps.
func TestMailYakDate(t *testing.T) {
	t.Parallel()

	my := New("mail.host.com:25", smtp.PlainAuth("", "user", "pass", "mail.host.com"))
	mail := my.NewMail()
	mail.From("from@example.org")
	mail.To("to@example.org")
	mail.Subject("Test subject")

	// send two emails at different times (discarding any errors)
	_, _ = mail.MimeBuf()
	dateOne := mail.date

	time.Sleep(1 * time.Second)

	_, _ = mail.MimeBuf()
	dateTwo := mail.date

	if dateOne == dateTwo {
		t.Errorf("MailYak.Send(): timestamp not updated: %v", dateOne)
	}
}
