package main

import (
	`fmt`
	`os`
	`path/filepath`
	`strings`
)

type set struct {
	dirs, files []string
	path, root  string
}

func makeSet(path, root string) (s *set, err error) {
	s = new(set)
	s.path, s.root = path, root
	if err = s.getFiles(); err != nil {
		return nil, err
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
				s.dirs = append(s.dirs, `./`)
			} else {
				s.dirs = append(s.dirs, `./`+
					strings.Replace(path, `\`, `/`, -1)+ `/`)
			}
		} else {
			path, _ = filepath.Rel(s.root, path)
			s.files = append(s.files, `/`+
				strings.Replace(path, `\`, `/`, -1))
		}
		return nil
	})
	return
}

func (s *set) createBackup() (err error) {
	var bakPath = filepath.Join(s.path, `.backup/`)
	for _, dir := range s.dirs {
		if err = os.MkdirAll(filepath.Join(bakPath, dir), 0666); err != nil {
			return
		}
	}
	for _, file := range s.files {
		if _, err = os.Stat(filepath.Join(s.path, file)); err == nil {
			if err = os.Rename(filepath.Join(s.path, file),
				filepath.Join(bakPath, file)); err != nil {
				return
			}
		}
	}
	return
}

func (s *set) patchFiles() (err error) {
	for _, file := range s.files {
		if err = copyFile(filepath.Join(s.root, file),
			filepath.Join(s.path, file)); err != nil {
			return
		}
	}
	return
}

func (s *set) createRecoverBat() (err error) {
	var batFile *os.File
	var _, batName = filepath.Split(os.Args[0])
	batName = `recover-` + strings.TrimSuffix(batName, filepath.Ext(batName))
	if batFile, err = os.Create(batName + `.bat`); err != nil {
		return
	}
	defer batFile.Close()
	
	fmt.Fprintln(batFile, "@echo off")
	
	fmt.Fprintln(batFile, "Recovering files . . . ")
	for _, file := range s.files {
		fmt.Fprintln(batFile, "\t"+file)
		fmt.Fprintf(batFile, "del /q \"%s\" >nul\n", file)
		fmt.Fprintf(batFile, "move \"%s\" \"%s\" >nul\n",
			`./`+filepath.Join(`./.backup/`, file), file)
	}
	
	fmt.Fprintln(batFile, "Cleaning empty directories . . . ")
	for _, dir := range s.dirs {
		fmt.Fprintln(batFile, "\t"+dir)
		fmt.Fprintf(batFile, "rd /q \"%s\" >nul 2>nul\n", dir)
		dir = `./` + filepath.Join(`./.backup/`, dir) + `/`
		fmt.Fprintln(batFile, "\t"+dir)
		fmt.Fprintf(batFile, "rd /q \"%s\" >nul 2>nul\n", dir)
	}
	return
}

func (s *set) recoverPatch() (err error) {
	for _, file := range s.files {
		var src = filepath.Join(s.path, `.backup/`, file)
		var dst = filepath.Join(s.path, file)
		if _, err = os.Stat(src); err == nil {
			if err = os.Remove(dst); err != nil {
				return
			}
			if err = os.Rename(src, dst); err != nil {
				return
			}
		}
	}
	for _, dir := range s.dirs {
		if err = os.RemoveAll(filepath.Join(s.path,
			`.backup/`, dir)); !os.IsNotExist(err) {
			return
		}
	}
	return
}
