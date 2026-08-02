package core

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTestManifest(t *testing.T, dir, appName, distributor string) {
	t.Helper()
	appDir := filepath.Join(dir, appName)
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", appDir, err)
	}
	content := "appname         = " + appName + "\n" +
		"version         = 1.0.0\n" +
		"display_name    = Test App\n" +
		"distributor     = " + distributor + "\n"
	if err := os.WriteFile(filepath.Join(appDir, "manifest"), []byte(content), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
}

func TestScanInstalled_DefaultAllowsOnlyConversun(t *testing.T) {
	dir := t.TempDir()
	writeTestManifest(t, dir, "app-conversun", "conversun")
	writeTestManifest(t, dir, "app-soleo", "soleo")
	writeTestManifest(t, dir, "app-other", "someone-else")

	// Ensure the env var doesn't leak from the environment.
	t.Setenv(allowedDistributorsEnv, "")

	apps, err := ScanInstalled(dir)
	if err != nil {
		t.Fatalf("ScanInstalled: %v", err)
	}
	if len(apps) != 1 {
		t.Fatalf("expected exactly 1 installed app (conversun), got %d: %+v", len(apps), apps)
	}
	if apps[0].AppName != "app-conversun" {
		t.Fatalf("expected app-conversun, got %q", apps[0].AppName)
	}
}

func TestScanInstalled_AllowsCustomDistributors(t *testing.T) {
	dir := t.TempDir()
	writeTestManifest(t, dir, "app-conversun", "conversun")
	writeTestManifest(t, dir, "app-soleo", "soleo")

	t.Setenv(allowedDistributorsEnv, "conversun, soleo")

	apps, err := ScanInstalled(dir)
	if err != nil {
		t.Fatalf("ScanInstalled: %v", err)
	}
	if len(apps) != 2 {
		t.Fatalf("expected 2 installed apps (conversun + soleo), got %d: %+v", len(apps), apps)
	}
}

func TestScanInstalled_SkipsUnknownDistributorsWithEnv(t *testing.T) {
	dir := t.TempDir()
	writeTestManifest(t, dir, "app-conversun", "conversun")
	writeTestManifest(t, dir, "app-soleo", "soleo")
	writeTestManifest(t, dir, "app-other", "someone-else")

	// conversun is always allowed (upstream default); the env var appends.
	t.Setenv(allowedDistributorsEnv, "soleo")

	apps, err := ScanInstalled(dir)
	if err != nil {
		t.Fatalf("ScanInstalled: %v", err)
	}
	if len(apps) != 2 {
		t.Fatalf("expected 2 installed apps (conversun + soleo), got %d: %+v", len(apps), apps)
	}
	seen := map[string]bool{}
	for _, a := range apps {
		seen[a.AppName] = true
	}
	if !seen["app-conversun"] || !seen["app-soleo"] {
		t.Fatalf("expected app-conversun and app-soleo, got %+v", seen)
	}
	if seen["app-other"] {
		t.Fatalf("app-other must be excluded, got %+v", seen)
	}
}
