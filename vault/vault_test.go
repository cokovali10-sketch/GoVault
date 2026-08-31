package vault

import (
	"path/filepath"
	"testing"
)

func TestSetAndGet(t *testing.T) {
	v := NewVault(filepath.Join(t.TempDir(), "test.vault"), "secret123")

	err := v.Set("db.host", []byte("localhost"))
	if err != nil {
		t.Fatal(err)
	}

	val, err := v.Get("db.host")
	if err != nil {
		t.Fatal(err)
	}
	if string(val) != "localhost" {
		t.Fatalf("expected localhost, got %s", val)
	}
}

func TestGetNotFound(t *testing.T) {
	v := NewVault(filepath.Join(t.TempDir(), "test.vault"), "pass")
	_, err := v.Get("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent key")
	}
}

func TestDelete(t *testing.T) {
	v := NewVault(filepath.Join(t.TempDir(), "test.vault"), "pass")
	v.Set("key1", []byte("val1"))

	if !v.Delete("key1") {
		t.Fatal("delete returned false")
	}
	if v.Delete("key1") {
		t.Fatal("delete returned true for already deleted key")
	}
}

func TestList(t *testing.T) {
	v := NewVault(filepath.Join(t.TempDir(), "test.vault"), "pass")
	v.Set("c", []byte("1"))
	v.Set("a", []byte("2"))
	v.Set("b", []byte("3"))

	keys := v.List()
	if len(keys) != 3 {
		t.Fatalf("expected 3 keys, got %d", len(keys))
	}
	if keys[0] != "a" || keys[1] != "b" || keys[2] != "c" {
		t.Fatalf("keys not sorted: %v", keys)
	}
}

func TestSearch(t *testing.T) {
	v := NewVault(filepath.Join(t.TempDir(), "test.vault"), "pass")
	v.Set("server", []byte("1"), "production")
	v.Set("client", []byte("2"), "staging")
	v.Set("database", []byte("3"), "production")

	results := v.Search("production")
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestSaveAndLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.vault")
	v1 := NewVault(path, "mypassword")
	v1.Set("key1", []byte("value1"))
	v1.Set("key2", []byte("value2"), "important")
	v1.Save()

	v2 := NewVault(path, "mypassword")
	err := v2.Load()
	if err != nil {
		t.Fatal(err)
	}

	val, _ := v2.Get("key1")
	if string(val) != "value1" {
		t.Fatalf("expected value1, got %s", val)
	}

	keys := v2.List()
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys after load, got %d", len(keys))
	}
}

func TestWrongPassword(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.vault")
	v1 := NewVault(path, "correct")
	v1.Set("key", []byte("value"))
	v1.Save()

	v2 := NewVault(path, "wrong")
	err := v2.Load()
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
}

func TestOverwrite(t *testing.T) {
	v := NewVault(filepath.Join(t.TempDir(), "test.vault"), "pass")
	v.Set("key", []byte("old"))
	v.Set("key", []byte("new"))

	val, _ := v.Get("key")
	if string(val) != "new" {
		t.Fatalf("expected new, got %s", val)
	}
}

func TestStats(t *testing.T) {
	v := NewVault(filepath.Join(t.TempDir(), "test.vault"), "pass")
	v.Set("a", []byte("hello"))
	v.Set("b", []byte("world!"))

	stats := v.Stats()
	if stats["entries"] != 2 {
		t.Fatalf("expected 2 entries, got %v", stats["entries"])
	}
}

func TestLoadNonexistent(t *testing.T) {
	v := NewVault(filepath.Join(t.TempDir(), "nonexistent.vault"), "pass")
	err := v.Load()
	if err != nil {
		t.Fatal("loading nonexistent vault should not error")
	}
}

func TestTags(t *testing.T) {
	v := NewVault(filepath.Join(t.TempDir(), "test.vault"), "pass")
	v.Set("key", []byte("val"), "tag1", "tag2")

	entry, err := v.GetEntry("key")
	if err != nil {
		t.Fatal(err)
	}
	if len(entry.Tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(entry.Tags))
	}
}
