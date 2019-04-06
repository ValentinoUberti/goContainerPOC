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
	rootfs            = "/home/vale/projects/gocontainer/rootfs/quux" //Change this according to your directory
	containerHostname = "container"                                   //Container host name, change if you want
	cgroupsDirectory  = "/sys/fs/cgroup"                              //Host cgroup directory
	pidsDirectory     = "vale"
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

func run() {
	log.Printf("Creating child process with kernel namespaces for running %v \n", os.Args[2:])

	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:   syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
		Unshareflags: syscall.CLONE_NEWNS,
	}

	must(cmd.Run())

}

func child() {
	log.Printf("Running %v in containerized child\n", os.Args[2:])

	cg()

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	must(syscall.Sethostname([]byte(containerHostname)))
	must(syscall.Chroot(rootfs))
	must(os.Chdir("/"))
	must(syscall.Mount("proc", "proc", "proc", 0, ""))

	must(cmd.Run())

	must(syscall.Unmount("proc", 0))

}

func cg() {

	cgroups := cgroupsDirectory
	pids := filepath.Join(cgroups, "pids")
	//Create pids directory on host
	os.Mkdir(filepath.Join(pids, pidsDirectory), 0755)
	//Set max number of process for this container
	must(ioutil.WriteFile(filepath.Join(pids, pidsDirectory+"/pids.max"), []byte("20"), 0700))

	// Removes the new cgroup in place after the container exits
	must(ioutil.WriteFile(filepath.Join(pids, pidsDirectory+"/notify_on_release"), []byte("1"), 0700))
	must(ioutil.WriteFile(filepath.Join(pids, pidsDirectory+"/cgroup.procs"), []byte(strconv.Itoa(os.Getpid())), 0700))
}

//Error checking, panic on error
func must(err error) {
	if err != nil {
		panic(err)
	}
}
