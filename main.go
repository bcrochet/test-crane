package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

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

	// create output file to which to write the tarball
	outputFile := "image.tar"
	log.Println("creating output file", outputFile)
	file, err := os.Create(outputFile)
	if err != nil {
		log.Fatal(err)
	}

	// write the tarball
	log.Println("writing to output file", outputFile)
	err = crane.Export(img, file)
	if err != nil {
		log.Fatal(err)
	}

	// extract the tarball, can probably do this programmatically in the future.
	os.MkdirAll("extracted", 0755)
	tr := exec.Command("tar", []string{"xvf", outputFile, "--directory", "extracted", "--no-same-permissions"}...)
	log.Println("running tar with the following flags:", tr.Args)
	err = tr.Run()
	if err != nil {
		log.Println("stdout:", tr.Stdout)
		log.Println("stderr:", tr.Stderr)
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
