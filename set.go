package main

import (
	`fmt`
	`os`
	`path/filepath`
	`strings`
)

type set struct {
	dirs, files, exFiles []string
	path, root           string
}

func makeSet(path, relRoot string) (s *set, err error) {
	fmt.Println("Preparing . . . ")
	s = new(set)
	s.path = path
	if s.root, err = s.getRoot(relRoot); err != nil {
		return nil, err
	}
	if err = s.getFiles(); err != nil {
		return nil, err
	}
	return
}

func (s *set) getRoot(relRoot string) (root string, err error) {
	var ex string
	if ex, err = os.Executable(); err != nil {
		return
	}
	root = filepath.Join(filepath.Dir(ex), relRoot)
	if _, err = os.Stat(root); err != nil && os.IsNotExist(err) {
		if root, err = os.Getwd(); err == nil {
			root = filepath.Join(root, relRoot)
			_, err = os.Stat(root)
		}
	}
	return
}

func (s *set) procErr(err error, msg string, needRecover bool, needPause bool) {
	fmt.Fprintf(os.Stderr, "Error occurred when %s: \n", msg)
	fmt.Fprintf(os.Stderr, "\t%s\n", err.Error())
	if needRecover {
		if err = s.recoverPatch(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to recover patch: \n")
			fmt.Fprintf(os.Stderr, "\t%s\n", err.Error())
		}
	}
	if needPause {
		pause()
		os.Exit(1)
	}
}

func (s *set) getFiles() (err error) {
	err = filepath.Walk(s.root, func(path string,
		info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			path, _ = filepath.Rel(s.root, path)
			if path == `.` {
				s.dirs = append(s.dirs, `.\`)
			} else {
				s.dirs = append(s.dirs, `.\`+path+`\`)
			}
		} else {
			path, _ = filepath.Rel(s.root, path)
			s.files = append(s.files, `.\`+path)
			if _, err = os.Stat(filepath.Join(s.path, path));
				err != nil && os.IsNotExist(err) {
				s.exFiles = append(s.exFiles, s.files[len(s.files)-1])
			}
		}
		return nil
	})
	return nil
}

func (s *set) createBackup() (err error) {
	var bakPath = filepath.Join(s.path, `.backup\`)
	for _, dir := range s.dirs {
		if err = os.Mkdir(filepath.Join(bakPath, dir),
			0666); err != nil && !os.IsExist(err) {
			return
		}
	}
	for _, file := range s.files {
		fmt.Printf("Backuping %s . . . \n", file)
		if _, err = os.Stat(filepath.Join(s.path, file)); err == nil {
			if err = os.Rename(filepath.Join(s.path, file),
				filepath.Join(bakPath, file)); err != nil {
				return
			}
		}
	}
	return nil
}

func (s *set) patchFiles() (err error) {
	for i, dir := range s.dirs {
		fmt.Printf("Patching %4d/%4d %s . . . \n", i+1, len(s.files), dir)
		os.Mkdir(filepath.Join(s.path, dir), 0666)
	}
	for i, file := range s.files {
		fmt.Printf("Patching %4d/%4d %s . . . \n", i+1, len(s.files), file)
		if err = copyFile(filepath.Join(s.root, file),
			filepath.Join(s.path, file)); err != nil {
			return
		}
	}
	return nil
}

func (s *set) createRecoverBat() (err error) {
	fmt.Println("Creating recover batch file . . . ")
	var batFile *os.File
	var _, batPath = filepath.Split(os.Args[0])
	batPath = filepath.Join(s.path, `recover-`+
		strings.TrimSuffix(batPath, filepath.Ext(batPath))+ `.bat`)
	if batFile, err = os.Create(batPath); err != nil {
		return
	}
	defer batFile.Close()
	
	fmt.Fprintln(batFile, "@echo off")
	
	fmt.Fprintln(batFile, "echo Recovering files . . . ")
	for _, file := range s.files {
		fmt.Fprintln(batFile, "echo     "+file)
		fmt.Fprintf(batFile, "del /q \"%s\" >nul\n", file)
		if _, err = os.Stat(filepath.Join(
			s.path, `.backup\`, file)); err == nil {
			fmt.Fprintf(batFile, "move \"%s\" \"%s\" >nul\n",
				`.\`+filepath.Join(`.\.backup\`, file), file)
		}
	}
	
	fmt.Fprintln(batFile, "echo Cleaning empty directories . . . ")
	var dirsRev = make([]string, len(s.dirs))
	copy(dirsRev, s.dirs)
	for _, dir := range getRevSortedStrSlice(s.dirs) {
		fmt.Fprintln(batFile, "echo     "+dir)
		fmt.Fprintf(batFile, "rd /q \"%s\" >nul 2>nul\n", dir)
		dir = `.\` + filepath.Join(`.\.backup\`, dir) + `\`
		fmt.Fprintln(batFile, "echo     "+dir)
		fmt.Fprintf(batFile, "rd /q \"%s\" >nul 2>nul\n", dir)
	}
	fmt.Fprintln(batFile, "pause")
	fmt.Fprintln(batFile, "del /q \"%~f0\" >nul 2>nul")
	return batFile.Sync()
}

func (s *set) recoverPatch() (err error) {
	var i = 1
	for _, file := range s.files {
		var src = filepath.Join(s.path, `.backup\`, file)
		var dst = filepath.Join(s.path, file)
		if _, err = os.Stat(src); err == nil {
			fmt.Printf("Recovering %4d/%4d `%s`\n", i,
				len(s.files)-len(s.exFiles), file)
			if err = os.Remove(dst); err != nil && !os.IsNotExist(err) {
				return
			}
			if err = os.Rename(src, dst); err != nil {
				return
			}
			i++
		}
	}
	for i, file := range s.exFiles {
		fmt.Printf("Deleting %4d/%4d `%s`\n", i+1, len(s.exFiles), file)
		if err = os.Remove(filepath.Join(s.path,
			file)); err != nil && !os.IsNotExist(err) {
			return
		}
	}
	for _, dir := range getRevSortedStrSlice(s.dirs) {
		var src = filepath.Join(s.path, `.backup\`, dir)
		var dst = filepath.Join(s.path, dir)
		var isEmpty bool
		if _, err = os.Stat(src); err == nil {
			if isEmpty, err = isDirEmpty(src); err != nil {
				return
			} else if err = os.Remove(src); isEmpty && err != nil {
				return
			}
		}
		if _, err = os.Stat(dst); err == nil {
			if isEmpty, err = isDirEmpty(dst); err != nil {
				return
			} else if err = os.Remove(dst); isEmpty && err != nil {
				return
			}
		}
	}
	return nil
}
