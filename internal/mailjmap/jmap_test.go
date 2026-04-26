package mailjmap

import (
	"testing"

	"git.sr.ht/~rockorager/go-jmap"
	jmapmail "git.sr.ht/~rockorager/go-jmap/mail"
	"git.sr.ht/~rockorager/go-jmap/mail/email"

	"github.com/glw907/poplar/internal/config"
	"github.com/glw907/poplar/internal/mail"
)

func TestNew_AccountName(t *testing.T) {
	b := New(config.AccountConfig{Name: "alice@example.com"})
	if got := b.AccountName(); got != "alice@example.com" {
		t.Errorf("AccountName = %q, want %q", got, "alice@example.com")
	}
}

func TestBackend_DisconnectWithoutConnect(t *testing.T) {
	b := New(config.AccountConfig{Name: "alice"})
	if err := b.Disconnect(); err != nil {
		t.Fatalf("Disconnect on never-connected: %v", err)
	}
}

// newTestBackend returns a Backend wired to fake with the given folders
// pre-populated and session.PrimaryAccounts set to accountID.
func newTestBackend(fake *fakeClient, accountID string, folders map[string]folderEntry) *Backend {
	b := NewWithClient(config.AccountConfig{Name: "test"}, fake)
	b.session = &jmap.Session{
		PrimaryAccounts: map[jmap.URI]jmap.ID{
			jmapmail.URI: jmap.ID(accountID),
		},
	}
	if folders != nil {
		b.folders = folders
	}
	return b
}

// --- ListFolders ---

func TestListFolders_ReturnsCachedFolders(t *testing.T) {
	fake := &fakeClient{}
	folders := map[string]folderEntry{
		"Inbox": {id: "mb-1", folder: mail.Folder{Name: "Inbox"}},
		"Sent":  {id: "mb-2", folder: mail.Folder{Name: "Sent"}},
	}
	b := newTestBackend(fake, "acct-1", folders)

	got, err := b.ListFolders()
	if err != nil {
		t.Fatalf("ListFolders: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("len = %d, want 2", len(got))
	}
	// No server calls should have been made.
	if len(fake.sent) != 0 {
		t.Errorf("unexpected server call with pre-populated folders")
	}
}

// --- OpenFolder ---

func TestOpenFolder(t *testing.T) {
	tests := []struct {
		name      string
		openName  string
		wantErr   bool
		wantCurr  string
	}{
		{
			name:     "known folder sets current",
			openName: "Inbox",
			wantErr:  false,
			wantCurr: "Inbox",
		},
		{
			name:     "unknown folder returns error",
			openName: "Nonexistent",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeClient{}
			folders := map[string]folderEntry{
				"Inbox": {id: "mb-1", folder: mail.Folder{Name: "Inbox"}},
			}
			b := newTestBackend(fake, "acct-1", folders)

			err := b.OpenFolder(tt.openName)
			if (err != nil) != tt.wantErr {
				t.Errorf("OpenFolder(%q) error = %v, wantErr %v", tt.openName, err, tt.wantErr)
			}
			if !tt.wantErr && b.current != tt.wantCurr {
				t.Errorf("current = %q, want %q", b.current, tt.wantCurr)
			}
		})
	}
}

// --- QueryFolder ---

func TestQueryFolder(t *testing.T) {
	tests := []struct {
		name      string
		folder    string
		offset    int
		limit     int
		respond   func(*jmap.Request) (*jmap.Response, error)
		wantUIDs  []mail.UID
		wantTotal int
		wantErr   bool
	}{
		{
			name:   "happy path returns UIDs and total",
			folder: "Inbox",
			offset: 0,
			limit:  50,
			respond: func(_ *jmap.Request) (*jmap.Response, error) {
				return fakeResponse(&jmap.Invocation{
					Name: "Email/query",
					Args: &email.QueryResponse{
						IDs:   []jmap.ID{"e-1", "e-2", "e-3"},
						Total: 42,
					},
				}), nil
			},
			wantUIDs:  []mail.UID{"e-1", "e-2", "e-3"},
			wantTotal: 42,
		},
		{
			name:    "unknown folder returns error",
			folder:  "Nonexistent",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeClient{respond: tt.respond}
			folders := map[string]folderEntry{
				"Inbox": {id: "mb-1", folder: mail.Folder{Name: "Inbox"}},
			}
			b := newTestBackend(fake, "acct-42", folders)

			uids, total, err := b.QueryFolder(tt.folder, tt.offset, tt.limit)
			if (err != nil) != tt.wantErr {
				t.Errorf("QueryFolder error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if total != tt.wantTotal {
				t.Errorf("total = %d, want %d", total, tt.wantTotal)
			}
			if len(uids) != len(tt.wantUIDs) {
				t.Fatalf("len(uids) = %d, want %d", len(uids), len(tt.wantUIDs))
			}
			for i, uid := range uids {
				if uid != tt.wantUIDs[i] {
					t.Errorf("uids[%d] = %q, want %q", i, uid, tt.wantUIDs[i])
				}
			}
		})
	}
}

func TestQueryFolder_RequestShape(t *testing.T) {
	var capturedReq *jmap.Request
	fake := &fakeClient{
		respond: func(req *jmap.Request) (*jmap.Response, error) {
			capturedReq = req
			return fakeResponse(&jmap.Invocation{
				Name: "Email/query",
				Args: &email.QueryResponse{
					IDs:   []jmap.ID{"e-1"},
					Total: 1,
				},
			}), nil
		},
	}
	folders := map[string]folderEntry{
		"Inbox": {id: "mb-1", folder: mail.Folder{Name: "Inbox"}},
	}
	b := newTestBackend(fake, "acct-42", folders)

	if _, _, err := b.QueryFolder("Inbox", 10, 25); err != nil {
		t.Fatalf("QueryFolder: %v", err)
	}

	if capturedReq == nil {
		t.Fatal("no request was sent")
	}
	if len(capturedReq.Calls) != 1 {
		t.Fatalf("calls count = %d, want 1", len(capturedReq.Calls))
	}

	inv := capturedReq.Calls[0]
	q, ok := inv.Args.(*email.Query)
	if !ok {
		t.Fatalf("invocation args type = %T, want *email.Query", inv.Args)
	}
	if q.Account != "acct-42" {
		t.Errorf("account = %q, want %q", q.Account, "acct-42")
	}
	fc, ok := q.Filter.(*email.FilterCondition)
	if !ok {
		t.Fatalf("filter type = %T, want *email.FilterCondition", q.Filter)
	}
	if fc.InMailbox != "mb-1" {
		t.Errorf("InMailbox = %q, want %q", fc.InMailbox, "mb-1")
	}
	if len(q.Sort) != 1 || q.Sort[0].Property != "receivedAt" || q.Sort[0].IsAscending {
		t.Errorf("unexpected sort: %+v", q.Sort)
	}
	if q.Position != 10 {
		t.Errorf("Position = %d, want 10", q.Position)
	}
	if q.Limit != 25 {
		t.Errorf("Limit = %d, want 25", q.Limit)
	}
	if !q.CalculateTotal {
		t.Error("CalculateTotal should be true")
	}
}
