package modupdater

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseModConfig(t *testing.T) {
	// Create a temporary test config
	testXML := `<?xml version="1.0" encoding="utf-8"?>
<ModConfig xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <Core Enabled="true">
    <Path />
  </Core>
  <Local Enabled="true">
    <Path Value="/app/mods/Workshop_3576112002" />
  </Local>
  <Local Enabled="false">
    <Path Value="/app/mods/Workshop_3575689739" />
  </Local>
</ModConfig>`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "modconfig.xml")

	if err := os.WriteFile(configPath, []byte(testXML), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	config, err := ParseModConfig(configPath)
	if err != nil {
		t.Fatalf("ParseModConfig failed: %v", err)
	}

	if len(config.Locals) != 2 {
		t.Errorf("Expected 2 local mods, got %d", len(config.Locals))
	}

	if config.Locals[0].Enabled != true {
		t.Error("First local mod should be enabled")
	}

	if config.Locals[1].Enabled != false {
		t.Error("Second local mod should be disabled")
	}
}

func TestWorkshopID(t *testing.T) {
	tests := []struct {
		value      string
		expectedID uint64
		shouldWork bool
	}{
		{"/app/mods/Workshop_3576112002", 3576112002, true},
		{"/app/mods/workshop_123456", 123456, true},
		{"/app/mods/WORKSHOP_999", 999, true},
		{"/app/mods/invalid", 0, false},
		{"", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			item := PathItem{Value: tt.value}
			id, ok := item.WorkshopID()

			if ok != tt.shouldWork {
				t.Errorf("Expected ok=%v, got %v", tt.shouldWork, ok)
			}

			if tt.shouldWork && id != tt.expectedID {
				t.Errorf("Expected ID %d, got %d", tt.expectedID, id)
			}
		})
	}
}

func TestGetModsToUpdate(t *testing.T) {
	testXML := `<?xml version="1.0" encoding="utf-8"?>
<ModConfig xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <Core Enabled="true">
    <Path />
  </Core>
  <Local Enabled="true">
    <Path Value="/app/mods/Workshop_100" />
  </Local>
  <Local Enabled="false">
    <Path Value="/app/mods/Workshop_200" />
  </Local>
  <Local Enabled="true">
    <Path Value="/app/mods/Workshop_300" />
  </Local>
</ModConfig>`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "modconfig.xml")

	if err := os.WriteFile(configPath, []byte(testXML), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	config, err := ParseModConfig(configPath)
	if err != nil {
		t.Fatalf("ParseModConfig failed: %v", err)
	}

	// Test ignoring disabled
	mods, err := config.GetModsToUpdate(configPath, true, nil)
	if err != nil {
		t.Fatalf("GetModsToUpdate failed: %v", err)
	}

	if len(mods) != 2 {
		t.Errorf("Expected 2 enabled mods, got %d", len(mods))
	}

	// Test including disabled
	mods, err = config.GetModsToUpdate(configPath, false, nil)
	if err != nil {
		t.Fatalf("GetModsToUpdate failed: %v", err)
	}

	if len(mods) != 3 {
		t.Errorf("Expected 3 total mods, got %d", len(mods))
	}

	// Test specific mod IDs
	mods, err = config.GetModsToUpdate(configPath, false, []uint64{100, 300})
	if err != nil {
		t.Fatalf("GetModsToUpdate failed: %v", err)
	}

	if len(mods) != 2 {
		t.Errorf("Expected 2 specific mods, got %d", len(mods))
	}
}

func TestCopyDirContent(t *testing.T) {
	// Create a temporary source directory
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "src")
	dstDir := filepath.Join(tmpDir, "dst")

	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}

	// Create some test files
	testFile1 := filepath.Join(srcDir, "file1.txt")
	testFile2 := filepath.Join(srcDir, "file2.txt")

	if err := os.WriteFile(testFile1, []byte("content1"), 0644); err != nil {
		t.Fatalf("Failed to write test file 1: %v", err)
	}

	if err := os.WriteFile(testFile2, []byte("content2"), 0644); err != nil {
		t.Fatalf("Failed to write test file 2: %v", err)
	}

	// Create a subdirectory
	subDir := filepath.Join(srcDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	testFile3 := filepath.Join(subDir, "file3.txt")
	if err := os.WriteFile(testFile3, []byte("content3"), 0644); err != nil {
		t.Fatalf("Failed to write test file 3: %v", err)
	}

	// Copy the directory
	if err := CopyDirContent(srcDir, dstDir); err != nil {
		t.Fatalf("CopyDirContent failed: %v", err)
	}

	// Verify files were copied
	dstFile1 := filepath.Join(dstDir, "file1.txt")
	dstFile2 := filepath.Join(dstDir, "file2.txt")
	dstFile3 := filepath.Join(dstDir, "subdir", "file3.txt")

	for _, path := range []string{dstFile1, dstFile2, dstFile3} {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Expected file %s to exist", path)
		}
	}

	// Verify content
	content, err := os.ReadFile(dstFile1)
	if err != nil {
		t.Fatalf("Failed to read copied file: %v", err)
	}

	if string(content) != "content1" {
		t.Errorf("Expected content 'content1', got '%s'", string(content))
	}
}

func TestParseDownloadedMods(t *testing.T) {
	output := `Some text before
Downloaded item 123456 to "/path/to/mod1"
More text
Downloaded item 789012 to "/another/path/mod2"
Some text after`

	results := parseDownloadedMods(output)

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	if results[0].ModID != 123456 {
		t.Errorf("Expected ModID 123456, got %d", results[0].ModID)
	}

	if results[0].Path != "/path/to/mod1" {
		t.Errorf("Expected path '/path/to/mod1', got '%s'", results[0].Path)
	}

	if results[1].ModID != 789012 {
		t.Errorf("Expected ModID 789012, got %d", results[1].ModID)
	}
}
