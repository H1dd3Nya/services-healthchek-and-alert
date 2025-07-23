package storage

import (
	"os"
	"testing"
)

func TestBoltStorage_SiteCRUD(t *testing.T) {
	os.Remove("test.db")
	store, err := NewBoltStorage("test.db")
	if err != nil {
		t.Fatal(err)
	}
	site := Site{ID: "1", Name: "Test", URL: "http://test", CheckType: CheckHTTP, IntervalSeconds: 10}
	if err := store.AddSite(site); err != nil {
		t.Fatal(err)
	}
	got, err := store.GetSite("1")
	if err != nil || got.Name != "Test" {
		t.Fatal("get failed")
	}
	sites, _ := store.ListSites()
	if len(sites) != 1 {
		t.Fatal("list failed")
	}
	if err := store.DeleteSite("1"); err != nil {
		t.Fatal(err)
	}
}

func TestBoltStorage_History(t *testing.T) {
	os.Remove("test2.db")
	store, _ := NewBoltStorage("test2.db")
	h := SiteCheckHistory{SiteID: "1", CheckedAt: 123, Duration: 100, HTTPCode: 200, Success: true}
	if err := store.AddCheckHistory(h); err != nil {
		t.Fatal(err)
	}
	hist, err := store.ListCheckHistory("1", 10)
	if err != nil || len(hist) == 0 {
		t.Fatal("history failed")
	}
}

// Для PostgresStorage можно реализовать аналогичные тесты с использованием тестовой БД или mock.
