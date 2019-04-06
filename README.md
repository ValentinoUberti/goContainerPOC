# goContainerPOC
POC by Lix Rice @DockerCon2017

This code is made by Liz Rice during the DockerCon 2017 and slightly edited by Valentino Uberti.
It's a POC for creating a basic containerized application using golang.

Please watch the video first:
https://www.youtube.com/watch?v=MHv6cWjvQjM&t=1316s

Not mentioned is the use of mkosi utility:
https://github.com/systemd/mkosi

With mkosi you can create a so called rootfs with usual directory and utilities:
ex : /
     /bin
     /etc
     /home
     ....
     
This rootfs will be mounted in the child process (aka container)
If you want to build a rootfs with mkosi (im on ubuntu 18):
   sudo mkosi -d ubuntu -t directory -o quux -r xenial
   
   (mkosi gave me some errors using bionic as release)
  
If you have some issue using mkosi, use the rootfs provided under /rootfs/quux

After cloning this project, open main.go and update at least the rootfs const with your directory


Example:

#sudo go run main.go run /bin/bash

--- Calling parent
--- Creating child process with kernel namespaces for running [/bin/bash] 
--- Calling child
--- Running [/bin/bash] in containerized child
--- root@container:/# hostname
--- container <-- this hostname came from containerHostname const
--- root@container:/# 


     


