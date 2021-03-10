package mailyak

import (
	"reflect"
	"testing"
)

func TestMailTo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		// Test description.
		name string
		// Parameters.
		addrs []string
		// Want
		want []string
	}{
		{
			"Single email",
			[]string{"dom@itsallbroken.com"},
			[]string{"dom@itsallbroken.com"},
		},
		{
			"Multiple email",
			[]string{"dom@itsallbroken.com", "ohnoes@itsallbroken.com"},
			[]string{"dom@itsallbroken.com", "ohnoes@itsallbroken.com"},
		},
		{
			"Empty last",
			[]string{"dom@itsallbroken.com", "ohnoes@itsallbroken.com", ""},
			[]string{"dom@itsallbroken.com", "ohnoes@itsallbroken.com"},
		},
		{
			"Empty Middle",
			[]string{"dom@itsallbroken.com", "", "ohnoes@itsallbroken.com"},
			[]string{"dom@itsallbroken.com", "ohnoes@itsallbroken.com"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := getMail()
			m.toAddrs = []string{}

			m.To(tt.addrs...)

			if !reflect.DeepEqual(m.toAddrs, tt.want) {
				t.Errorf("%q. Mail.To() = %v, want %v", tt.name, m.toAddrs, tt.want)
			}

		})
	}
}

func TestMailBcc(t *testing.T) {
	t.Parallel()

	tests := []struct {
		// Test description.
		name string
		// Parameters.
		addrs []string
		// Want
		want []string
	}{
		{
			"Single email",
			[]string{"dom@itsallbroken.com"},
			[]string{"dom@itsallbroken.com"},
		},
		{
			"Multiple email",
			[]string{"dom@itsallbroken.com", "ohnoes@itsallbroken.com"},
			[]string{"dom@itsallbroken.com", "ohnoes@itsallbroken.com"},
		},
		{
			"Empty last",
			[]string{"dom@itsallbroken.com", "ohnoes@itsallbroken.com", ""},
			[]string{"dom@itsallbroken.com", "ohnoes@itsallbroken.com"},
		},
		{
			"Empty Middle",
			[]string{"dom@itsallbroken.com", "", "ohnoes@itsallbroken.com"},
			[]string{"dom@itsallbroken.com", "ohnoes@itsallbroken.com"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := getMail()
			m.bccAddrs = []string{}
			m.Bcc(tt.addrs...)

			if !reflect.DeepEqual(m.bccAddrs, tt.want) {
				t.Errorf("%q. Mail.Bcc() = %v, want %v", tt.name, m.bccAddrs, tt.want)
			}
			putMail(m)
		})
	}
}
func TestMailSubject(t *testing.T) {
	t.Parallel()

	tests := []struct {
		// Test description.
		name string
		// Parameters.
		subject string
		// Want
		want string
	}{
		{
			"ASCII",
			"Banana\r\n",
			"Banana",
		},
		{
			"Q-encoded",
			"🍌\r\n",
			"=?UTF-8?q?=F0=9F=8D=8C?=",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := getMail()
			m.Subject(tt.subject)

			if !reflect.DeepEqual(m.subject, tt.want) {
				t.Errorf("%q. Mail.Subject() = %v, want %v", tt.name, m.subject, tt.want)
			}
			putMail(m)
		})
	}
}

func TestMailFromName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		// Test description.
		name string
		// Parameters.
		from string
		// Want
		want string
	}{
		{
			"ASCII",
			"Goat\r\n",
			"Goat",
		},
		{
			"Q-encoded",
			"🐐",
			"=?UTF-8?q?=F0=9F=90=90?=",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := getMail()
			m.FromName(tt.from)

			if !reflect.DeepEqual(m.fromName, tt.want) {
				t.Errorf("%q. Mail.Subject() = %v, want %v", tt.name, m.fromName, tt.want)
			}
			putMail(m)
		})
	}
}

func TestMailAddHeader(t *testing.T) {
	t.Parallel()

	tests := []struct {
		// Test description.
		name string
		// Parameters.
		from map[string]string
		// Want
		want map[string]string
	}{
		{
			"ASCII",
			map[string]string{
				"List-Unsubscribe": "http://example.com",
				"X-NASTY":          "true\r\nBcc: badguy@example.com",
			},
			map[string]string{
				"List-Unsubscribe": "http://example.com",
				"X-NASTY":          "trueBcc: badguy@example.com",
			},
		},
		{
			"Q-encoded",
			map[string]string{
				"X-BEETHOVEN": "für Elise",
			},
			map[string]string{
				"X-BEETHOVEN": "=?UTF-8?q?f=C3=BCr_Elise?=",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := getMail()

			for k, v := range tt.from {
				m.AddHeader(k, v)
			}

			if !reflect.DeepEqual(m.headers, tt.want) {
				t.Errorf("%q. Mail.AddHeader() = %v, want %v", tt.name, m.headers, tt.want)
			}
			putMail(m)
		})
	}
}
