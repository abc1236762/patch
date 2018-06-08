package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
)

func main() {
	var err error
	var path string
	var cpySet *set
	if path, err = getPath(); err != nil {
		cpySet.procErr(err, "opening path "+"`"+path+"`", false, true)
	}
	if cpySet, err = makeSet(path, `.\copy_files\`); err != nil {
		cpySet.procErr(err, "getting file list of .\\copy_files\\", false, true)
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
	pause()
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

func isDirEmpty(path string) (isEmpty bool, err error) {
	var fileInfo os.FileInfo
	if fileInfo, err = os.Stat(path); err != nil || !fileInfo.IsDir() {
		return
	}
	var file *os.File
	if file, err = os.Open(path); err != nil {
		return
	}
	defer file.Close()
	if _, err = file.Readdirnames(1); err == io.EOF {
		return true, nil
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

func getRevSortedStrSlice(src []string) (dst []string) {
	dst = make([]string, len(src))
	copy(dst, src)
	sort.Sort(sort.Reverse(sort.StringSlice(dst)))
	return
}