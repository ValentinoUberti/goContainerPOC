/*
POC by Liz Rice
Video tutorial:
https://www.youtube.com/watch?v=MHv6cWjvQjM&t=1316s

slightly edited by Valentino Uberti
*/

package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
)

/*
rootfs = root filesystem directory
if you want to build rootfs use mkosi tool:

-- build on ubuntu 18-04 with:
   sudo mkosi -d ubuntu -t directory -o quux -r xenial
These files are avaible under rootfs/quux if you have some issue with mkosi
Remeber to fix this path with yours!
These files will be mounted in the container root
*/

const (
	// TODO: manage the root directory with an environment variable or a flag
	rootfs            = "/home/gbsalinetti/go/src/github.com/giannisalinetti/goContainerPOC/rootfs/quux" //Change this according to your directory
	containerHostname = "demo"                                                                           //Container host name, change if you want
	cgroupsDirectory  = "/sys/fs/cgroup"                                                                 //Host cgroup directory
	pidsDirectory     = "container"
)

/*
 go run main.go run <cmd> <args>
 example:
	 sudo go run main.go run /bin/bash
*/

func main() {

	switch os.Args[1] {
	case "run":
		log.Println("Calling parent")
		run()
	case "child":
		log.Println("Calling child")
		child()

	default:
		panic("help")
	}
}

// run executed the main same process with the "child" argument
func run() {
	log.Printf("Creating child process with kernel namespaces for running %v \n", os.Args[2:])

	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
		// Setting the Unshareflags to CLONE_NEWNS here is a workaround to hide
		// the container mounts to the host. This happens since systemd forces
		// all mount namespaces to be shared. The only workaround was to remount
		// internally with MS_PRIVATE flag. The workaround was implemented in
		// Go 1.9 and remounts with MS_REC|MS_PRIVATE when the unshare flag
		// CLONE_NEWNS is set. For more details see the following thread:
		// https://go-review.googlesource.com/c/go/+/38471
		Unshareflags: syscall.CLONE_NEWNS,
	}

	errorHandler(cmd.Run())
	defer errorHandler(cgDestroy())

	// This log message shows exactly when the child process returns to the
	// parent
	log.Println("Container exited")

}

// child runs the containerized process in the previuosly isolated namespaces
func child() {
	log.Printf("Running %v in containerized child\n", os.Args[2:])

	errorHandler(cgInit())

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Println("Setting container hostname")
	errorHandler(syscall.Sethostname([]byte(containerHostname)))
	log.Println("Changing container root directory")
	errorHandler(syscall.Chroot(rootfs))
	errorHandler(os.Chdir("/"))

	// Mounting the proc filesystem is necessary to access the processes
	// data
	log.Println("Mounting container proc filesystem")
	errorHandler(syscall.Mount("proc", "proc", "proc", 0, ""))
	defer errorHandler(syscall.Unmount("proc", 0))

	errorHandler(cmd.Run())

}

func cgInit() error {
	cgroups := cgroupsDirectory
	pids := filepath.Join(cgroups, "pids")

	//Create pids directory on host
	err := os.Mkdir(filepath.Join(pids, pidsDirectory), 0755)
	if err != nil {
		return err
	}

	//Set max number of process for this container
	err = ioutil.WriteFile(filepath.Join(pids, pidsDirectory+"/pids.max"), []byte("20"), 0700)
	if err != nil {
		return err
	}

	// Removes the new cgroup in place after the container exits
	err = ioutil.WriteFile(filepath.Join(pids, pidsDirectory+"/notify_on_release"), []byte("1"), 0700)
	if err != nil {
		return err
	}

	// Populate the procs list for the namespace
	err = ioutil.WriteFile(filepath.Join(pids, pidsDirectory+"/cgroup.procs"), []byte(strconv.Itoa(os.Getpid())), 0700)
	if err != nil {
		return err
	}

	log.Println("Finished cgroups creation")

	return nil
}

func cgDestroy() error {
	cgroups := cgroupsDirectory
	pids := filepath.Join(cgroups, "pids")
	containerPids := filepath.Join(pids, pidsDirectory)
	_, err := os.Stat(containerPids)
	if err != nil {
		return err
	} else {
		err := os.RemoveAll(containerPids)
		if err != nil {
			return err
		}
	}
	return nil
}

//errorHandler manages errors by printing log.Fatal messages and exiting
func errorHandler(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
