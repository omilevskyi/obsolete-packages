// Trigger build
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"

	ut "github.com/omilevskyi/go/pkg/utils"
)

const appName = "obsolete-packages"

var (
	version, gitCommit string // -ldflags -X main.version=v0.0.0 -X main.gitCommit=[[:xdigit:]] -X main.makeBin=/usr/bin/make

	makeBin = "make"
)

// readStdout runs the specified command with arguments and captures its standard output.
// It returns the combined output as a single string (without newlines) and any error encountered.
func readStdout(cmdPath string, args ...string) (string, error) {
	var output bytes.Buffer

	command := exec.Command(cmdPath, args...)

	stdout, err := command.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("error setting up stdout pipe: %w", err)
	}

	if err = command.Start(); err != nil {
		return "", fmt.Errorf("error running the command: %w", err)
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		output.WriteString(scanner.Text()) // result is concatenated strings without \n
	}

	if err = scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading command output: %w", err)
	}

	if err = command.Wait(); err != nil {
		return "", fmt.Errorf("error waiting for command to finish: %w", err)
	}

	return output.String(), nil
}

func main() {
	var helpFlag, verboseFlag, versionFlag, deleteFlag bool

	flag.BoolVar(&helpFlag, "help", false, "Display help message")
	flag.BoolVar(&versionFlag, "version", false, "Show version information")
	flag.BoolVar(&verboseFlag, "verbose", false, "Enable verbose output")
	flag.BoolVar(&deleteFlag, "delete", false, "Delete obsolete packages")
	flag.Parse()

	if helpFlag {
		fmt.Fprintln(os.Stderr, "Usage: "+appName+" [-help] [-version] [-verbose] [-delete] [packages_directories]")
		os.Exit(0)
	}

	if versionFlag {
		fmt.Fprintf(os.Stderr, "Version: %s, Commit: %s\n", version, gitCommit)
		os.Exit(0)
	}

	args, data, err := flag.Args(), map[string]*[]VersionType{}, error(nil)
	if len(args) < 1 {
		rootDir, err := ut.RootDirectory()
		ut.IsErr(err, 201, "ut.RootDirectory()")

		portsDir, err := readStdout(makeBin, "-C", rootDir, "-V", "PORTSDIR")
		ut.IsErr(err, 202, "readStdout()")

		packagesDir, err := readStdout(makeBin, "-C", filepath.Join(portsDir, "ports-mgmt", "pkg"), "-V", "PKGREPOSITORY")
		ut.IsErr(err, 203, "readStdout()")

		args = []string{packagesDir}
	}

	// Recursively walks through each directory provided in args
	for i := 0; i < len(args); i++ {
		err = filepath.Walk(args[i], func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.Mode().IsRegular() {
				if k, ver := keyAndVersion(path); k != "" {
					if versions, ok := data[k]; ok {
						if !versionsContain(*versions, ver.Path) {
							*versions = append(*versions, *ver)
						}
					} else {
						data[k] = &[]VersionType{*ver}
					}
				} else if verboseFlag {
					fmt.Fprintln(os.Stderr, "Empty key:", path)
				}
			}
			return nil
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, "ERROR:", err)
		}
	}

	// Iterates over all version groups sorted by key, and does the job aimed for.
	for _, k := range ut.Arrange(ut.Keys(data)) {
		versions := *data[k]
		if vlen := len(versions); vlen > 1 {
			slices.SortFunc(versions, compareVersionDesc)
			for i := 1; i < vlen; i++ {
				if path := versions[i].Path; deleteFlag {
					if err = os.Remove(path); err != nil {
						fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
					} else if verboseFlag {
						fmt.Println(path)
					}
				} else {
					fmt.Println(path)
				}
			}
		}
	}
}
