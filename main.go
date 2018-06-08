package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

func main() {
	var err error
	var path string
	var cpySet *set
	
	if path, err = getPath(); err != nil {
		cpySet.procErr(err, "opening path "+"`"+path+"`", false, true)
	}
	if cpySet, err = makeSet(path, `./copy_files/`); err != nil {
		cpySet.procErr(err, "getting file list of `./copy_files/`", false, true)
	}
	if err = cpySet.createBackup(); err != nil {
		cpySet.procErr(err, "creating backup", true, true)
	}
	if err = cpySet.patchFiles(); err != nil {
		cpySet.procErr(err, "patching files", true, true)
	}
	if err = cpySet.createRecoverBat(); err != nil {
		cpySet.procErr(err, "creating recover batch file", true, true)
	}
}

func pause() {
	fmt.Print("Press enter key to continue . . . ")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func getPath() (path string, err error) {
	if len(os.Args) <= 1 {
		_, err = os.Stat(``)
		path = ``
	} else {
		_, err = os.Stat(os.Args[1])
		path = os.Args[1]
	}
	return
}

func copyFile(src, dst string) (err error) {
	var srcFile, dstFile *os.File
	if srcFile, err = os.Open(src); err != nil {
		return
	}
	defer srcFile.Close()
	if dstFile, err = os.Create(dst); err != nil {
		return
	}
	defer dstFile.Close()
	if _, err = io.Copy(dstFile, srcFile); err != nil {
		return
	}
	return dstFile.Sync()
}
