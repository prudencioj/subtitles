package subdb

import (
	"testing"
	"net/http"
)

func TestSearch(t *testing.T) {
	s := NewSubDB(http.DefaultClient)
	if s == nil {
		t.Fatal("subdb client failed to be created")
	}

	l, err := s.Search("")
	if len(l) > 0 && err == nil {
		t.Fatal("fail")
	}
}

func TestDownload(t *testing.T) {
	s := NewSubDB(http.DefaultClient)
	if s == nil {
		t.Fatal("subdb client failed to be created")
	}

	subs, err := s.Download("", "")
	if subs != nil && err == nil {
		t.Fatal("fail")
	}
}