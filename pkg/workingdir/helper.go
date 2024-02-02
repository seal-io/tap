package workingdir

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// extractChdirOption is a helper function to extract the -chdir,
// borrows from https://github.com/hashicorp/terraform/blob/ee58ac1851c8a433005df9863ed47796a9f6b5e7/main.go#L431-L475.
func extractChdirOption(args []string) (string, []string, error) {
	if len(args) == 0 {
		return "", args, nil
	}

	const (
		argName   = "-chdir"
		argPrefix = argName + "="
	)

	var (
		argPos   int
		argValue string
	)

	for i, arg := range args {
		if !strings.HasPrefix(arg, "-") {
			// Because the chdir option is a subcommand-agnostic one, we require
			// it to appear before any subcommand argument, so if we find a
			// non-option before we find -chdir then we are finished.
			break
		}

		if arg == argName || arg == argPrefix {
			return "", args, fmt.Errorf("must include an equals sign followed by a directory path, like -chdir=example")
		}

		if strings.HasPrefix(arg, argPrefix) {
			argPos = i
			argValue = arg[len(argPrefix):]
		}
	}

	// When we fall out here, we'll have populated argValue with a non-empty
	// string if the -chdir=... Option was present and valid, or left it
	// empty if it wasn't present.
	if argValue == "" {
		return "", args, nil
	}

	// If we did find the option then we'll need to produce a new args that
	// doesn't include it anymore.
	if argPos == 0 {
		// Easy case: we can just slice off the front.
		return argValue, args[1:], nil
	}
	// Otherwise we need to construct a new array and copy to it.
	newArgs := make([]string, len(args)-1)
	copy(newArgs, args[:argPos])
	copy(newArgs[argPos:], args[argPos+1:])

	return argValue, newArgs, nil
}

// cleanDir is a helper function to clean a directory,
// but keep the state files.
func cleanDir(tapDir string) error {
	if _, err := os.Lstat(tapDir); err != nil && os.IsNotExist(err) {
		return nil
	}

	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == tapDir {
			return nil
		}

		if info.IsDir() {
			if err = os.RemoveAll(path); err != nil {
				return err
			}

			// Skip walking the subdirectory.
			return filepath.SkipDir
		}

		// We don't remove any state files.
		switch filename := info.Name(); {
		default:
			return os.Remove(path)
		case strings.HasSuffix(filename, ".tfstate") ||
			strings.HasSuffix(filename, ".tfstate.backup"):
		}

		return nil
	}

	return filepath.Walk(tapDir, walkFn)
}

// copyDir is a helper function to copy a directory,
// borrows from https://github.com/hashicorp/terraform/blob/ee58ac1851c8a433005df9863ed47796a9f6b5e7/internal/copy/copy_dir.go#L38-L38.
//
// This function is modified to skip the .tap directory and .tf files in the root directory,
// and not to overwrite any state files.
func copyDir(src, dst string) error {
	if err := os.Mkdir(dst, 0o700); err != nil && !os.IsExist(err) {
		return fmt.Errorf("failed to create directory %q: %w", dst, err)
	}

	src, err := filepath.EvalSymlinks(src)
	if err != nil {
		return err
	}

	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == src {
			return nil
		}

		// [TAP] Skip rules.
		switch {
		case info.IsDir() && info.Name() == ".tap":
			// Skip .tap directory.
			return filepath.SkipDir
		case !info.IsDir() && filepath.Dir(path) == src && filepath.Ext(path) == ".tf":
			// Skip .tf files in the root directory.
			return nil
		}

		// The "path" has the src prefixed to it. We need to join our
		// destination with the path without the src on it.
		dstPath := filepath.Join(dst, path[len(src):])

		// We don't want to try and copy the same file over itself.
		if eq, err := isSameFile(path, dstPath); eq {
			return nil
		} else if err != nil {
			return err
		}

		// If we have a directory, make that subdirectory, then continue
		// the walk.
		if info.IsDir() {
			if path == filepath.Join(src, dst) {
				// Dst is in src; don't walk it.
				return nil
			}

			if err := os.MkdirAll(dstPath, 0o755); err != nil {
				return err
			}

			return nil
		}

		// If the current path is a symlink, recreate the symlink relative to
		// the dst directory.
		if info.Mode()&os.ModeSymlink == os.ModeSymlink {
			target, err := os.Readlink(path)
			if err != nil {
				return err
			}

			return os.Symlink(target, dstPath)
		}

		// [TAP] Skip rules.
		switch filename := info.Name(); {
		case strings.HasSuffix(filename, ".tfstate") ||
			strings.HasSuffix(filename, ".tfstate.backup"):
			// We don't overwrite any state files.
			if _, err := os.Lstat(dstPath); err == nil {
				return nil
			}
		default:
		}

		// If we have a file, copy the contents.
		srcF, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcF.Close()

		dstF, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		defer dstF.Close()

		if _, err := io.Copy(dstF, srcF); err != nil {
			return err
		}

		// Chmod it.
		return os.Chmod(dstPath, info.Mode())
	}

	return filepath.Walk(src, walkFn)
}

// isSameFile is a helper function to compare two files,
// borrows from https://github.com/hashicorp/terraform/blob/ee58ac1851c8a433005df9863ed47796a9f6b5e7/internal/copy/copy_dir.go#L127-L127.
func isSameFile(a, b string) (bool, error) {
	if a == b {
		return true, nil
	}

	aInfo, err := os.Lstat(a)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, err
	}

	bInfo, err := os.Lstat(b)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, err
	}

	return os.SameFile(aInfo, bInfo), nil
}
