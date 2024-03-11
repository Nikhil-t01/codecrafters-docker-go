package main

import (
	"os"
	"os/exec"
	"path"
	"syscall"

	"github.com/codecrafters-io/docker-starter-go/app/docker"
	"github.com/codecrafters-io/docker-starter-go/app/util"
)

// Usage: your_docker.sh run <image> <command> <arg1> <arg2> ...
func main() {
	imageName := os.Args[2]
	command := os.Args[3]
	args := os.Args[4:len(os.Args)]

	isolateFileSystem(imageName)

	cmd := exec.Command(command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWPID,
	}

	err := cmd.Run()
	exitCode := cmd.ProcessState.ExitCode()
	util.ExitOnError(err, "Err", exitCode)
}

func isolateFileSystem(imageName string) {
	dir, err := os.MkdirTemp("", "tmp_my_docker_*")
	util.ExitOnError(err, "Error in creating temp directory", 1)

	err = os.Chmod(dir, 0777)
	util.ExitOnError(err, "Error in chmod of temp directory", 1)

	image := docker.NewImage(imageName)
	image.PullImage(dir)

	err = os.MkdirAll(path.Join(dir, "/usr/local/bin"), 0755)
	util.ExitOnError(err, "Error in creating bin in the temp directory", 1)

	err = os.Link("/usr/local/bin/docker-explorer", path.Join(dir, "usr/local/bin/docker-explorer"))
	util.ExitOnError(err, "Error in copying exectable docker-explorer", 1)

	err = syscall.Chroot(dir)
	util.ExitOnError(err, "Error in chroot into temp dir", 1)

	err = os.Chdir("/")
	util.ExitOnError(err, "Error in chdir into root dir", 1)
}
