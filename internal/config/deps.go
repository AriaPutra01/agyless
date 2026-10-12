package config

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// Home returns user's home directory.
func Home() string {
	h, err := os.UserHomeDir()
	if err != nil {
		return os.Getenv("HOME")
	}
	return h
}

// FindBinary searches binary in PATH and in ~/.local/bin.
func FindBinary(name string) string {
	if p, err := exec.LookPath(name); err == nil {
		return p
	}
	localBin := filepath.Join(Home(), ".local", "bin", name)
	if _, err := os.Stat(localBin); err == nil {
		return localBin
	}
	for _, usrBin := range []string{"/usr/local/bin", "/opt/homebrew/bin"} {
		cand := filepath.Join(usrBin, name)
		if _, err := os.Stat(cand); err == nil {
			return cand
		}
	}
	return ""
}

// AgylessBin returns current agyless executable path or ~/.local/bin/agyless.
func AgylessBin() string {
	if exe, err := os.Executable(); err == nil && filepath.Base(exe) != "main" {
		if resolved, err := filepath.EvalSymlinks(exe); err == nil {
			return resolved
		}
		return exe
	}
	if p := FindBinary("agyless"); p != "" {
		return p
	}
	return filepath.Join(Home(), ".local", "bin", "agyless")
}

// EnsureDeps checks and installs rtk, context-mode, and codegraph if missing.
func EnsureDeps() error {
	// 1. Check RTK
	if FindBinary("rtk") == "" {
		fmt.Println("rtk not found. Downloading prebuilt rtk binary...")
		if err := installRtk(); err != nil {
			fmt.Printf("Warning: failed to auto-install rtk: %v. You can install manually: https://github.com/rtk-ai/rtk\n", err)
		} else {
			fmt.Println("rtk installed successfully in ~/.local/bin/rtk")
		}
	}

	// 2. Check context-mode and codegraph
	needNpm := false
	if FindBinary("context-mode") == "" {
		needNpm = true
	}
	if FindBinary("codegraph") == "" {
		needNpm = true
	}

	if needNpm {
		if FindBinary("npm") != "" {
			fmt.Println("Installing context-mode and @colbymchenry/codegraph via npm...")
			cmd := exec.Command("npm", "install", "-g", "context-mode@latest", "@colbymchenry/codegraph@latest")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Printf("Warning: npm install -g failed: %v. Ensure you have Node.js 22+ installed.\n", err)
			} else {
				fmt.Println("npm packages installed successfully.")
			}
		} else {
			fmt.Println("Warning: npm not found. Please install Node.js 22+ to use context-mode and codegraph.")
		}
	}

	return nil
}

func rtkAsset() string {
	arch := "x86_64"
	if runtime.GOARCH == "arm64" {
		arch = "aarch64"
	}
	switch runtime.GOOS {
	case "darwin":
		return fmt.Sprintf("rtk-%s-apple-darwin.tar.gz", arch)
	case "linux":
		if runtime.GOARCH == "arm64" {
			return "rtk-aarch64-unknown-linux-gnu.tar.gz"
		}
		return "rtk-x86_64-unknown-linux-musl.tar.gz"
	}
	return ""
}

func installRtk() error {
	asset := rtkAsset()
	if asset == "" {
		return fmt.Errorf("unsupported platform %s/%s for prebuilt rtk", runtime.GOOS, runtime.GOARCH)
	}

	url := "https://github.com/rtk-ai/rtk/releases/latest/download/" + asset
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("download %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s failed: status %s", url, resp.Status)
	}

	gzReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return fmt.Errorf("gzip reader: %w", err)
	}
	defer gzReader.Close()

	destDir := filepath.Join(Home(), ".local", "bin")
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("mkdir %s: %w", destDir, err)
	}

	tarReader := tar.NewReader(gzReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("tar next: %w", err)
		}

		if header.Typeflag == tar.TypeReg && (header.Name == "rtk" || filepath.Base(header.Name) == "rtk") {
			targetPath := filepath.Join(destDir, "rtk")
			f, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
			if err != nil {
				return fmt.Errorf("create %s: %w", targetPath, err)
			}
			if _, err := io.Copy(f, tarReader); err != nil {
				f.Close()
				return fmt.Errorf("write %s: %w", targetPath, err)
			}
			f.Close()
			return nil
		}
	}

	return fmt.Errorf("binary rtk not found in tar archive")
}
