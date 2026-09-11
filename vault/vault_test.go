package vault

import (
	"encoding/json"
	"os"
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

func TestSearchPrefix(t *testing.T) {
	v := NewVault(filepath.Join(t.TempDir(), "test.vault"), "pass")
	v.Set("db.host", []byte("1"))
	v.Set("db.port", []byte("2"))
	v.Set("web.port", []byte("3"))

	results := v.SearchPrefix("db.")
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if r != "db.host" && r != "db.port" {
			t.Fatalf("unexpected result: %s", r)
		}
	}
}

func TestSearchPrefixCaseInsensitive(t *testing.T) {
	v := NewVault(filepath.Join(t.TempDir(), "test.vault"), "pass")
	v.Set("DB.Host", []byte("1"))
	v.Set("db.port", []byte("2"))

	results := v.SearchPrefix("db")
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestSearchRegex(t *testing.T) {
	v := NewVault(filepath.Join(t.TempDir(), "test.vault"), "pass")
	v.Set("server.prod", []byte("1"))
	v.Set("server.staging", []byte("2"))
	v.Set("client.prod", []byte("3"))

	results, err := v.SearchRegex("^server\\.")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if r != "server.prod" && r != "server.staging" {
			t.Fatalf("unexpected result: %s", r)
		}
	}
}

func TestSearchRegexInvalid(t *testing.T) {
	v := NewVault(filepath.Join(t.TempDir(), "test.vault"), "pass")
	_, err := v.SearchRegex("[invalid")
	if err == nil {
		t.Fatal("expected error for invalid regex")
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

func TestRename(t *testing.T) {
	v := NewVault(filepath.Join(t.TempDir(), "test.vault"), "pass")
	v.Set("old.key", []byte("secret"), "important")

	if !v.Rename("old.key", "new.key") {
		t.Fatal("rename returned false")
	}

	val, err := v.Get("new.key")
	if err != nil {
		t.Fatal(err)
	}
	if string(val) != "secret" {
		t.Fatalf("expected secret, got %s", val)
	}

	if v.Exists("old.key") {
		t.Fatal("old key should not exist")
	}

	entry, _ := v.GetEntry("new.key")
	if len(entry.Tags) != 1 || entry.Tags[0] != "important" {
		t.Fatalf("tags not preserved: %v", entry.Tags)
	}
}

func TestRenameOldNotFound(t *testing.T) {
	v := NewVault(filepath.Join(t.TempDir(), "test.vault"), "pass")
	if v.Rename("nonexistent", "new") {
		t.Fatal("rename should return false for nonexistent key")
	}
}

func TestRenameNewExists(t *testing.T) {
	v := NewVault(filepath.Join(t.TempDir(), "test.vault"), "pass")
	v.Set("key1", []byte("a"))
	v.Set("key2", []byte("b"))

	if v.Rename("key1", "key2") {
		t.Fatal("rename should return false when new key exists")
	}
}

func TestRenamePersistsSaveLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.vault")
	v1 := NewVault(path, "pass")
	v1.Set("alpha", []byte("one"))
	v1.Rename("alpha", "beta")
	v1.Save()

	v2 := NewVault(path, "pass")
	if err := v2.Load(); err != nil {
		t.Fatal(err)
	}
	if !v2.Exists("beta") {
		t.Fatal("renamed key should exist after load")
	}
	if v2.Exists("alpha") {
		t.Fatal("old key should not exist after load")
	}
}

func TestExportJSON(t *testing.T) {
	v := NewVault(filepath.Join(t.TempDir(), "test.vault"), "pass")
	v.Set("key1", []byte("value1"), "tag1")
	v.Set("key2", []byte("value2"))

	data, err := v.ExportJSON()
	if err != nil {
		t.Fatal(err)
	}

	var exported map[string]interface{}
	if err := json.Unmarshal(data, &exported); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if len(exported) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(exported))
	}

	k1 := exported["key1"].(map[string]interface{})
	if k1["value"] != "value1" {
		t.Fatalf("expected value1, got %v", k1["value"])
	}
}

func TestImportJSON(t *testing.T) {
	v := NewVault(filepath.Join(t.TempDir(), "test.vault"), "pass")
	importData := `{
		"imported.key": {
			"value": "imported-value",
			"tags": ["imported"]
		}
	}`

	if err := v.ImportJSON([]byte(importData)); err != nil {
		t.Fatal(err)
	}

	val, err := v.Get("imported.key")
	if err != nil {
		t.Fatal(err)
	}
	if string(val) != "imported-value" {
		t.Fatalf("expected imported-value, got %s", val)
	}
}

func TestExportImportRoundtrip(t *testing.T) {
	dir := t.TempDir()

	v1 := NewVault(filepath.Join(dir, "source.vault"), "pass")
	v1.Set("db.host", []byte("localhost"), "prod")
	v1.Set("db.port", []byte("5432"), "prod")
	v1.Set("api.key", []byte("abc123"))

	exportData, err := v1.ExportJSON()
	if err != nil {
		t.Fatal(err)
	}

	importPath := filepath.Join(dir, "export.json")
	if err := os.WriteFile(importPath, exportData, 0644); err != nil {
		t.Fatal(err)
	}

	v2 := NewVault(filepath.Join(dir, "dest.vault"), "pass")
	fileData, err := os.ReadFile(importPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := v2.ImportJSON(fileData); err != nil {
		t.Fatal(err)
	}

	if v2.Count() != 3 {
		t.Fatalf("expected 3 entries, got %d", v2.Count())
	}

	val, _ := v2.Get("db.host")
	if string(val) != "localhost" {
		t.Fatalf("expected localhost, got %s", val)
	}

	entry, _ := v2.GetEntry("db.host")
	if len(entry.Tags) != 1 || entry.Tags[0] != "prod" {
		t.Fatalf("tags not preserved: %v", entry.Tags)
	}
}

func TestBackup(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "vault.db")
	backupPath := filepath.Join(dir, "vault.bak")

	v := NewVault(path, "pass")
	v.Set("key1", []byte("value1"))
	v.Set("key2", []byte("value2"), "tag1")
	v.Save()

	if err := v.Backup(backupPath); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(backupPath)
	if err != nil {
		t.Fatalf("backup file not created: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("backup file is empty")
	}

	restored := NewVault(filepath.Join(dir, "restored.vault"), "pass")
	encrypted, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatal(err)
	}
	data, err := restored.decrypt(encrypted)
	if err != nil {
		t.Fatal(err)
	}
	var store VaultStore
	if err := json.Unmarshal(data, &store); err != nil {
		t.Fatal(err)
	}
	if len(store.Entries) != 2 {
		t.Fatalf("expected 2 entries in backup, got %d", len(store.Entries))
	}
}

func TestBackupIsEncrypted(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "vault.db")
	backupPath := filepath.Join(dir, "vault.bak")

	v := NewVault(path, "pass")
	v.Set("secret", []byte("sensitive-data"))
	v.Save()

	if err := v.Backup(backupPath); err != nil {
		t.Fatal(err)
	}

	raw, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) == `{"entries":` {
		t.Fatal("backup is not encrypted")
	}

	wrongKey := NewVault(filepath.Join(dir, "other.vault"), "wrongpass")
	_, err = wrongKey.decrypt(raw)
	if err == nil {
		t.Fatal("backup should not decrypt with wrong password")
	}
}

func TestBackupWithDifferentPassword(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "vault.db")
	backupPath := filepath.Join(dir, "vault.bak")

	v := NewVault(path, "oldpass")
	v.Set("key1", []byte("val1"))
	v.Save()

	if err := v.Backup(backupPath); err != nil {
		t.Fatal(err)
	}

	wrongKey := NewVault(filepath.Join(dir, "other.vault"), "newpass")
	encrypted, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatal(err)
	}
	_, err = wrongKey.decrypt(encrypted)
	if err == nil {
		t.Fatal("backup should not decrypt with wrong password")
	}
}

func TestChangePassword(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "vault.db")

	v := NewVault(path, "oldpass")
	v.Set("key1", []byte("value1"))
	v.Save()

	v.ChangePassword("newpass")
	if err := v.Save(); err != nil {
		t.Fatal(err)
	}

	v2 := NewVault(path, "newpass")
	if err := v2.Load(); err != nil {
		t.Fatal(err)
	}
	val, _ := v2.Get("key1")
	if string(val) != "value1" {
		t.Fatalf("expected value1, got %s", val)
	}
}

func TestChangePasswordWrongFails(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "vault.db")

	v := NewVault(path, "correct")
	v.Set("key", []byte("val"))
	v.Save()

	v.ChangePassword("wrong")
	v.Save()

	v2 := NewVault(path, "correct")
	if err := v2.Load(); err == nil {
		t.Fatal("loading with old password should fail after change")
	}
}

func TestRenamePreservesCreatedAt(t *testing.T) {
	v := NewVault(filepath.Join(t.TempDir(), "test.vault"), "pass")
	v.Set("key", []byte("val"))
	entry, _ := v.GetEntry("key")
	createdAt := entry.CreatedAt

	v.Rename("key", "renamed")
	renamed, _ := v.GetEntry("renamed")

	if !renamed.CreatedAt.Equal(createdAt) {
		t.Fatal("CreatedAt should be preserved after rename")
	}
}

func TestSetAndGetLargeValue(t *testing.T) {
	v := NewVault(filepath.Join(t.TempDir(), "test.vault"), "pass")
	big := make([]byte, 1024*1024)
	for i := range big {
		big[i] = byte(i % 256)
	}

	v.Set("big", big)
	val, _ := v.Get("big")
	if len(val) != len(big) {
		t.Fatalf("expected %d bytes, got %d", len(big), len(val))
	}
}

func TestCount(t *testing.T) {
	v := NewVault(filepath.Join(t.TempDir(), "test.vault"), "pass")
	if v.Count() != 0 {
		t.Fatal("empty vault should have 0 entries")
	}
	v.Set("a", []byte("1"))
	v.Set("b", []byte("2"))
	if v.Count() != 2 {
		t.Fatalf("expected 2 entries, got %d", v.Count())
	}
}

func TestExists(t *testing.T) {
	v := NewVault(filepath.Join(t.TempDir(), "test.vault"), "pass")
	v.Set("key", []byte("val"))
	if !v.Exists("key") {
		t.Fatal("expected Exists to return true")
	}
	if v.Exists("missing") {
		t.Fatal("expected Exists to return false")
	}
}
