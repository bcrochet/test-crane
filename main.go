package main

import (
	"archive/tar"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
)

func main() {
	// lazy check args for this test
	if len(os.Args) != 2 {
		fmt.Printf("expected one parameter - the image to extract. E.g.:\n\t %s <yourimage>\n", os.Args[0])
		os.Exit(1)
	}

	image := os.Args[1]
	log.Println("starting")
	options := make([]crane.Option, 0) // empty option list

	// optionally login... trying to use the Basic implementation, but there are others in the authn pkg.
	username := os.Getenv("REG_USERNAME")
	password := os.Getenv("REG_PASSWORD")
	if notEmpty([]string{username, password}...) {
		log.Println("authenticating using provided credentials")
		authOption := crane.WithAuth(&authn.Basic{
			Username: username,
			Password: password,
		})
		options = append(options, authOption)
	} else {
		log.Println("continuing with anonymous authentication to a registry")
		log.Println("if you intended to using authentication, set the REG_USERNAME and REG_PASSWORD environment variables")
	}

	// pull the image
	log.Println("pulling image", image)
	img, err := crane.Pull(image, options...)
	if err != nil {
		log.Fatal(err)
	}

	r, w := io.Pipe()

	// Extract the tarball
	go func() {
		log.Println("writing to output dir")
		err = crane.Export(img, w)
		if err != nil {
			log.Fatal(err)
		}
		w.Close()
	}()

	// extract the tarball, can probably do this programmatically in the future.
	log.Println("Extracting the tar from the export")
	os.MkdirAll("extracted", 0755)
	err = Untar("extracted", r)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("done")
}

// notEmpty returns true if all strings in s are non-zero length strings
func notEmpty(s ...string) bool {
	for _, v := range s {
		if len(v) <= 0 {
			return false
		}
	}

	return true
}

// Untar takes a destination path and a reader; a tar reader loops over the tarfile
// creating the file structure at 'dst' along the way, and writing any files
func Untar(dst string, r io.Reader) error {
	tr := tar.NewReader(r)

	for {
		header, err := tr.Next()

		switch {

		// if no more files are found return
		case err == io.EOF:
			return nil

		// return any other error
		case err != nil:
			return err

		// if the header is nil, just skip it (not sure how this happens)
		case header == nil:
			continue
		}

		// the target location where the dir/file should be created
		target := filepath.Join(dst, header.Name)

		// the following switch could also be done using fi.Mode(), not sure if there
		// a benefit of using one vs. the other.
		// fi := header.FileInfo()

		// check the file type
		switch header.Typeflag {

		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}

		// if it's a file create it
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			// copy over contents
			if _, err := io.Copy(f, tr); err != nil {
				return err
			}

			// manually close here after each file operation; defering would cause each file close
			// to wait until all operations have completed.
			f.Close()

			// if it's a link create it
		case tar.TypeSymlink:
			// head, _ := tar.FileInfoHeader(header.FileInfo(), "link")
			log.Println(fmt.Sprintf("Old: %s, New: %s", header.Linkname, header.Name))
			err := os.Symlink(header.Linkname, filepath.Join(dst, header.Name))
			if err != nil {
				log.Println(fmt.Sprintf("Error creating link: %s. Ignoring.", header.Name))
				continue
			}
		}
	}
}
