TARGETS = mountkernfs.sh hostname.sh mountdevsubfs.sh procps urandom hwclock.sh checkroot.sh mountall-bootclean.sh mountall.sh bootmisc.sh mountnfs.sh mountnfs-bootclean.sh checkfs.sh checkroot-bootclean.sh
INTERACTIVE = checkroot.sh checkfs.sh
mountdevsubfs.sh: mountkernfs.sh
procps: mountkernfs.sh
urandom: hwclock.sh
hwclock.sh: mountdevsubfs.sh
checkroot.sh: hwclock.sh mountdevsubfs.sh hostname.sh
mountall-bootclean.sh: mountall.sh
mountall.sh: checkfs.sh checkroot-bootclean.sh
bootmisc.sh: mountall-bootclean.sh mountnfs-bootclean.sh checkroot-bootclean.sh
mountnfs-bootclean.sh: mountnfs.sh
checkfs.sh: checkroot.sh
checkroot-bootclean.sh: checkroot.sh
