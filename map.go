package main

import(
	"syscall"
	"unsafe"
)

func memfdCreate(path string) (uintptr, error){
     sc, err := syscall.BytePtrFromString(path)
     if err != nil {
     	return 0, err
     }

     fd, _, errno := syscall.Syscall(319, uintptr(unsafe.Pointer(sc)), 0, 0)

     if int(fd) == -1 {
     	return fd, errno
     }

     return fd, nil
}

func copyToMem(fd uintptr, buf []byte) error {
     _, err :=  syscall.Write(int(fd), buf)
     return err
}

func execveAt(fd uintptr) error {
     sc, err := syscall.BytePtrFromString("")
     if err != nil {
     	return err
     }
     ret, _, errno := syscall.Syscall6(322, fd, uintptr(unsafe.Pointer(sc)), 0, 0, 0x1000, 0)
     if int(ret) == -1 {
     	return errno
     }
     // unreachable
     return nil
}