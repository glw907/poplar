package content

import "testing"

func TestParseHeaders(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		from    []Address
		to      []Address
		subject string
	}{
		{
			name:    "simple",
			input:   "From: Alice <alice@example.com>\r\nTo: Bob <bob@example.com>\r\nDate: Mon, 5 Jan 2026\r\nSubject: Hello\r\n\r\n",
			from:    []Address{{Name: "Alice", Email: "alice@example.com"}},
			to:      []Address{{Name: "Bob", Email: "bob@example.com"}},
			subject: "Hello",
		},
		{
			name:    "bare email",
			input:   "From: alice@example.com\r\nSubject: Test\r\n\r\n",
			from:    []Address{{Email: "alice@example.com"}},
			subject: "Test",
		},
		{
			name:  "multiple recipients",
			input: "From: Alice <alice@example.com>\r\nTo: Bob <bob@example.com>, Carol <carol@example.com>\r\nSubject: Group\r\n\r\n",
			from:  []Address{{Name: "Alice", Email: "alice@example.com"}},
			to: []Address{
				{Name: "Bob", Email: "bob@example.com"},
				{Name: "Carol", Email: "carol@example.com"},
			},
			subject: "Group",
		},
		{
			name:    "folded header",
			input:   "From: Alice <alice@example.com>\r\nSubject: This is a very\r\n long subject line\r\n\r\n",
			from:    []Address{{Name: "Alice", Email: "alice@example.com"}},
			subject: "This is a very long subject line",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := ParseHeaders(tt.input)
			if len(h.From) != len(tt.from) {
				t.Fatalf("From count: got %d, want %d", len(h.From), len(tt.from))
			}
			for i, a := range h.From {
				if a != tt.from[i] {
					t.Errorf("From[%d]: got %v, want %v", i, a, tt.from[i])
				}
			}
			if len(h.To) != len(tt.to) {
				t.Fatalf("To count: got %d, want %d", len(h.To), len(tt.to))
			}
			for i, a := range h.To {
				if a != tt.to[i] {
					t.Errorf("To[%d]: got %v, want %v", i, a, tt.to[i])
				}
			}
			if h.Subject != tt.subject {
				t.Errorf("Subject: got %q, want %q", h.Subject, tt.subject)
			}
		})
	}
}
