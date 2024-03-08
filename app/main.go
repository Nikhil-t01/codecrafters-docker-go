package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"syscall"
)

// Usage: your_docker.sh run <image> <command> <arg1> <arg2> ...
func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	// fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage!
	//
	command := os.Args[3]
	args := os.Args[4:len(os.Args)]

	isolateFileSystem()

	cmd := exec.Command(command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWPID,
	}

	err := cmd.Run()
	exitCode := cmd.ProcessState.ExitCode()
	if err != nil {
		fmt.Printf("Err: %v", err)
	}
	os.Exit(exitCode)
}

func isolateFileSystem() {
	dir, err := os.MkdirTemp("", "tmp_my_docker_*")
	processInternalError(err, "Error in creating temp directory")

	err = os.Chmod(dir, 0755)
	processInternalError(err, "Error in chmod of temp directory")

	err = os.MkdirAll(path.Join(dir, "/usr/local/bin"), 0755)
	processInternalError(err, "Error in creating bin in the temp directory")

	err = os.Link("/usr/local/bin/docker-explorer", path.Join(dir, "usr/local/bin/docker-explorer"))
	processInternalError(err, "Error in copying exectable docker-explorer")

	err = syscall.Chroot(dir)
	processInternalError(err, "Error in chroot into temp dir")

	err = os.Chdir("/")
	processInternalError(err, "Error in chdir into root dir")
}

func processInternalError(err error, errMsg string) {
	if err != nil {
		fmt.Printf("%s: %v\n", errMsg, err)
		os.Exit(1)
	}
}
