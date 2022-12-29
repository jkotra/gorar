package main

import (
	"fmt"
	"flag"
	"io"
	"os"
	"path/filepath"
	"github.com/nwaples/rardecode"
	"archive/zip"
	"runtime"
	"strings"
)

//VERSION 0.2.0
const VERSION = "0.2.0"

//RarExtractor ..
func RarExtractor(path string, destination string) error {

	rr, err := rardecode.OpenReader(path, "")

	if err != nil {
		return fmt.Errorf("read: failed to create reader: %v", err)
	}

	//sum := 1
	for {
		//sum += sum
		header, err := rr.Next()
		if err == io.EOF {
			break
		}

		if header.IsDir {
			err = mkdir(filepath.Join(destination, header.Name))
			if err != nil {
				return err
			}
			continue
		}
		err = mkdir(filepath.Dir(filepath.Join(destination, header.Name)))
		if err != nil {
			return err
		}

		err = writeNewFile(filepath.Join(destination, header.Name), rr, header.Mode())
		if err != nil {
			return err
		}

	}

	return nil
}

//ZipExtractor ..
func ZipExtractor(source string, destination string) error {

	r, err := zip.OpenReader(source)
	if err != nil {
		return err
	}
	defer r.Close()

	return unzipAll(&r.Reader, destination)
}

func unzipAll(r *zip.Reader, destination string) error {
	for _, zf := range r.File {
		if err := unzipFile(zf, destination); err != nil {
			return err
		}
	}

	return nil
}


func unzipFile(zf *zip.File, destination string) error {
	if strings.HasSuffix(zf.Name, "/") {
		return mkdir(filepath.Join(destination, zf.Name))
	}

	rc, err := zf.Open()
	if err != nil {
		return fmt.Errorf("%s: open compressed file: %v", zf.Name, err)
	}
	defer rc.Close()

	return writeNewFile(filepath.Join(destination, zf.Name), rc, zf.FileInfo().Mode())
}


func mkdir(dirPath string) error {
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		return fmt.Errorf("%s: making directory: %v", dirPath, err)
	}
	return nil
}


func writeNewFile(fpath string, in io.Reader, fm os.FileMode) error {
	err := os.MkdirAll(filepath.Dir(fpath), 0755)
	if err != nil {
		return fmt.Errorf("%s: making directory for file: %v", fpath, err)
	}

	out, err := os.Create(fpath)
	if err != nil {
		return fmt.Errorf("%s: creating new file: %v", fpath, err)
	}
	defer out.Close()

	err = out.Chmod(fm)
	if err != nil && runtime.GOOS != "windows" {
		return fmt.Errorf("%s: changing file mode: %v", fpath, err)
	}

	_, err = io.Copy(out, in)
	if err != nil {
		return fmt.Errorf("%s: writing file: %v", fpath, err)
	}
	return nil
}

func main() {
	i := flag.String("i", "", "path to rar file")
	o := flag.String("o", "", "destination path")
	flag.Parse()

	if *i == "" {
		fmt.Println("path is required")
		os.Exit(1)
	}

	if *o == "" {
		fmt.Println("destination is required")
		os.Exit(1)
	}

	err := RarExtractor(*i, *o)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	//display help if no arguments are passed
	if len(os.Args) == 1 {
		fmt.Println("Usage: gorar -i <path to file> -o <destination path>")
		os.Exit(1)
	}
}
