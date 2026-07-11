//go:build android

// Android entry point for OpenNox. Built with -buildmode=c-shared as
// libmain.so and loaded by SDLActivity, which calls the exported SDL_main.
package main

/*
#cgo pkg-config: sdl2

#include <SDL.h>
#include <SDL_system.h>
*/
import "C"

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	opennox "github.com/opennox/opennox/v1"
)

// Android discards an app's stdout/stderr; point them at a file next to the
// game data so engine logs (and Go panics) are recoverable.
func redirectStdio(dir string) {
	if dir == "" {
		return
	}
	f, err := os.OpenFile(filepath.Join(dir, "opennox.log"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return
	}
	syscall.Dup3(int(f.Fd()), 1, 0)
	syscall.Dup3(int(f.Fd()), 2, 0)
}

func findData(external string) string {
	if d := os.Getenv("NOX_DATA"); d != "" {
		return d
	}
	candidates := []string{
		"/storage/emulated/0/Nox",
		"/storage/emulated/0/Download/Nox",
		"/sdcard/Nox",
	}
	if external != "" {
		candidates = append(candidates, filepath.Join(external, "nox"))
	}
	for _, dir := range candidates {
		if _, err := os.Stat(filepath.Join(dir, "gamedata.bin")); err == nil {
			return dir
		}
	}
	return ""
}

// Shared storage (/storage, /sdcard) is FUSE-backed on Android 11+; the
// engine's many small reads through it turn seconds of loading into minutes.
func isFuseStorage(dir string) bool {
	return strings.HasPrefix(dir, "/storage/") || strings.HasPrefix(dir, "/sdcard")
}

// mirrorData copies the game data to internal (ext4) storage, skipping the
// large MOVIES dir, which is symlinked instead: movie playback streams
// sequentially, which FUSE handles fine. Copies are skipped when the
// destination file already exists with the same size.
func mirrorData(src, internal string) string {
	dst := filepath.Join(internal, "noxdata")
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return src
	}
	var files, copied int
	var copiedBytes int64
	err := filepath.Walk(src, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return nil // unreadable entry; not fatal
		}
		rel, err := filepath.Rel(src, path)
		if err != nil || rel == "." {
			return nil
		}
		base := strings.ToLower(fi.Name())
		if fi.IsDir() {
			if base == "movies" {
				link := filepath.Join(dst, fi.Name())
				if _, err := os.Lstat(link); os.IsNotExist(err) {
					os.Symlink(path, link)
				}
				return filepath.SkipDir
			}
			return os.MkdirAll(filepath.Join(dst, rel), 0o755)
		}
		switch filepath.Ext(base) {
		case ".exe", ".dll", ".lnk", ".log":
			return nil // Windows leftovers / our own log
		}
		files++
		target := filepath.Join(dst, rel)
		if tfi, err := os.Stat(target); err == nil && tfi.Size() == fi.Size() {
			return nil
		}
		in, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer in.Close()
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
		if err != nil {
			return nil
		}
		defer out.Close()
		n, _ := io.Copy(out, in)
		copied++
		copiedBytes += n
		return nil
	})
	if err != nil {
		return src
	}
	fmt.Printf("data mirror: %d files checked, %d copied (%.1f MB) -> %s\n",
		files, copied, float64(copiedBytes)/(1024*1024), dst)
	if _, err := os.Stat(filepath.Join(dst, "gamedata.bin")); err != nil {
		return src
	}
	return dst
}

//export SDL_main
func SDL_main(argc C.int, argv **C.char) C.int {
	internal := C.GoString(C.SDL_AndroidGetInternalStoragePath())
	external := C.GoString(C.SDL_AndroidGetExternalStoragePath())

	// App processes have no shell environment; point everything the engine
	// may touch (config, saves, temp) at app-owned storage.
	if internal != "" {
		os.Setenv("HOME", internal)
		os.Setenv("XDG_CONFIG_HOME", filepath.Join(internal, "config"))
		os.Setenv("XDG_DATA_HOME", filepath.Join(internal, "data"))
		os.MkdirAll(filepath.Join(internal, "config"), 0o755)
		os.MkdirAll(filepath.Join(internal, "data"), 0o755)
	}

	dataDir := findData(external)
	if dataDir != "" {
		redirectStdio(dataDir) // log stays user-reachable next to the data
	} else if external != "" {
		redirectStdio(external)
	}
	if dataDir != "" && internal != "" && isFuseStorage(dataDir) {
		dataDir = mirrorData(dataDir, internal)
	}
	os.Setenv("NOX_DATA", dataDir)
	if external != "" {
		os.Chdir(external)
	}

	args := []string{"opennox"}
	if internal != "" {
		args = append(args, "--config", filepath.Join(internal, "config", "opennox.yml"))
	}
	if err := opennox.RunArgs(args); err != nil && err != flag.ErrHelp {
		if code, ok := err.(opennox.ErrExit); ok {
			return C.int(code)
		}
		slog.Error("opennox exited with error", "err", err)
		return 1
	}
	return 0
}

func main() {}
