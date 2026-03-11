package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"teamspeaker/ts3portable/internal/payload"
)

const (
	appName          = "TS3 Portable Launcher"
	runtimeDirName   = "runtime"
	clientDirName    = "client"
	dataDirName      = "data"
	manifestFileName = ".payload.sha256"
	logFileName      = "launcher.log"
)

var clientNames = []string{
	"ts3client_win64.exe",
	"ts3client_win32.exe",
}

func main() {
	if err := run(); err != nil {
		reportFatalError(err)
		os.Exit(1)
	}
}

func run() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("%s: resolve executable path: %w", appName, err)
	}

	appDir := filepath.Dir(exePath)
	payloadBytes, err := payload.Load(appDir)
	if err != nil {
		return fmt.Errorf("%s: load payload: %w", appName, err)
	}

	sum := sha256.Sum256(payloadBytes)
	payloadHash := hex.EncodeToString(sum[:])

	runtimeRoot := filepath.Join(appDir, runtimeDirName)
	clientRoot := filepath.Join(runtimeRoot, clientDirName)
	manifestPath := filepath.Join(runtimeRoot, manifestFileName)

	if err := ensureExtracted(payloadBytes, payloadHash, runtimeRoot, clientRoot, manifestPath); err != nil {
		return fmt.Errorf("%s: prepare runtime: %w", appName, err)
	}

	clientExe, err := findClientExecutable(clientRoot)
	if err != nil {
		return fmt.Errorf("%s: locate client executable: %w", appName, err)
	}

	profile, err := ensureProfileDirs(appDir)
	if err != nil {
		return fmt.Errorf("%s: prepare profile: %w", appName, err)
	}

	if err := launchClient(clientExe, os.Args[1:], filepath.Dir(clientExe), mergeEnv(os.Environ(), profile.env()...)); err != nil {
		return fmt.Errorf("%s: start TeamSpeak: %w", appName, err)
	}

	return nil
}

func ensureExtracted(payloadBytes []byte, payloadHash, runtimeRoot, clientRoot, manifestPath string) error {
	currentHash, err := os.ReadFile(manifestPath)
	if err == nil && strings.TrimSpace(string(currentHash)) == payloadHash {
		if _, statErr := os.Stat(clientRoot); statErr == nil {
			if _, findErr := findClientExecutable(clientRoot); findErr == nil {
				return nil
			}
		}
	}

	if err := os.MkdirAll(runtimeRoot, 0o755); err != nil {
		return err
	}

	stagingRoot := clientRoot + ".staging"
	backupRoot := clientRoot + ".backup"

	if err := os.RemoveAll(stagingRoot); err != nil {
		return err
	}
	if err := os.RemoveAll(backupRoot); err != nil {
		return err
	}
	if err := os.MkdirAll(stagingRoot, 0o755); err != nil {
		return err
	}

	if err := unzip(payloadBytes, stagingRoot); err != nil {
		return err
	}
	if _, err := findClientExecutable(stagingRoot); err != nil {
		return fmt.Errorf("extracted payload is missing TeamSpeak client executable: %w", err)
	}

	if _, err := os.Stat(clientRoot); err == nil {
		if err := os.Rename(clientRoot, backupRoot); err != nil {
			return err
		}
	}

	if err := os.Rename(stagingRoot, clientRoot); err != nil {
		_ = os.RemoveAll(stagingRoot)
		if _, backupErr := os.Stat(backupRoot); backupErr == nil {
			_ = os.Rename(backupRoot, clientRoot)
		}
		return err
	}

	_ = os.RemoveAll(backupRoot)

	if err := os.WriteFile(manifestPath, []byte(payloadHash), 0o644); err != nil {
		return err
	}

	return nil
}

func unzip(payloadBytes []byte, dest string) error {
	reader, err := zip.NewReader(bytes.NewReader(payloadBytes), int64(len(payloadBytes)))
	if err != nil {
		return fmt.Errorf("open zip payload: %w", err)
	}

	for _, file := range reader.File {
		if err := extractZipEntry(file, dest); err != nil {
			return err
		}
	}

	return nil
}

func extractZipEntry(file *zip.File, dest string) error {
	targetPath := filepath.Join(dest, file.Name)
	cleanDest := filepath.Clean(dest) + string(os.PathSeparator)
	cleanTarget := filepath.Clean(targetPath)
	if !strings.HasPrefix(cleanTarget, cleanDest) && cleanTarget != filepath.Clean(dest) {
		return fmt.Errorf("zip entry escapes destination: %s", file.Name)
	}

	info := file.FileInfo()
	if info.IsDir() {
		return os.MkdirAll(cleanTarget, 0o755)
	}

	if err := os.MkdirAll(filepath.Dir(cleanTarget), 0o755); err != nil {
		return err
	}

	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.OpenFile(cleanTarget, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode().Perm())
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return err
	}

	return nil
}

func findClientExecutable(root string) (string, error) {
	var found string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}

		name := strings.ToLower(d.Name())
		for _, candidate := range clientNames {
			if name == candidate {
				found = path
				return io.EOF
			}
		}

		return nil
	})

	if errors.Is(err, io.EOF) && found != "" {
		return found, nil
	}
	if err != nil {
		return "", err
	}
	return "", fmt.Errorf("none of %v found under %s", clientNames, root)
}

type profileDirs struct {
	base     string
	roaming  string
	local    string
	home     string
	download string
}

func ensureProfileDirs(appDir string) (profileDirs, error) {
	base := filepath.Join(appDir, dataDirName)
	p := profileDirs{
		base:     base,
		roaming:  filepath.Join(base, "Roaming"),
		local:    filepath.Join(base, "Local"),
		home:     filepath.Join(base, "Home"),
		download: filepath.Join(base, "Downloads"),
	}

	for _, dir := range []string{p.roaming, p.local, p.home, p.download, filepath.Join(p.local, "Temp")} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return profileDirs{}, err
		}
	}

	return p, nil
}

func (p profileDirs) env() []string {
	return []string{
		"APPDATA=" + p.roaming,
		"LOCALAPPDATA=" + p.local,
		"USERPROFILE=" + p.home,
		"HOME=" + p.home,
		"HOMEDRIVE=" + filepath.VolumeName(p.home),
		"HOMEPATH=" + trimVolume(p.home),
		"TMP=" + filepath.Join(p.local, "Temp"),
		"TEMP=" + filepath.Join(p.local, "Temp"),
	}
}

func mergeEnv(base []string, overrides ...string) []string {
	envMap := make(map[string]string, len(base)+len(overrides))
	order := make([]string, 0, len(base)+len(overrides))

	add := func(entry string) {
		parts := strings.SplitN(entry, "=", 2)
		key := parts[0]
		value := ""
		if len(parts) == 2 {
			value = parts[1]
		}
		if _, exists := envMap[key]; !exists {
			order = append(order, key)
		}
		envMap[key] = value
	}

	for _, entry := range base {
		add(entry)
	}
	for _, entry := range overrides {
		add(entry)
	}

	out := make([]string, 0, len(order))
	for _, key := range order {
		out = append(out, key+"="+envMap[key])
	}
	return out
}

func trimVolume(path string) string {
	volume := filepath.VolumeName(path)
	rest := strings.TrimPrefix(path, volume)
	if rest == "" {
		return string(os.PathSeparator)
	}
	return rest
}

func reportFatalError(err error) {
	message := err.Error()
	fmt.Fprintln(os.Stderr, message)

	if logErr := writeLauncherLog(message); logErr == nil {
		message += "\n\nDetails were also written to launcher.log next to the executable."
	}

	showErrorDialog(appName, message)
}

func writeLauncherLog(message string) error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	logPath := filepath.Join(filepath.Dir(exePath), logFileName)
	return os.WriteFile(logPath, []byte(message+"\n"), 0o644)
}
