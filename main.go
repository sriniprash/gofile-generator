package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var match = regexp.MustCompile("import\\s+?\\(([^\\)]+)\\)")
var inQuotes = regexp.MustCompile("\"([^\"]+)\"")

func appendIfMissing(slice []string, s string) []string {

	spl := strings.Split(s, "/")

	toKeep := []string{}
	for i := 0; i < 3; i++ {
		if len(spl) >= (i + 1) {
			toKeep = append(toKeep, spl[i])
			// toKeep[i] = spl[i]
		} else {
			break
		}
	}

	repo := strings.Join(toKeep, "/")

	for _, el := range slice {
		if el == repo {
			return slice
		}
	}

	return append(slice, repo)
}
func getImportsInDirectory(dir string) ([]string, error) {
	final := []string{}

	d, err := os.Open(dir)
	if err != nil {
		return []string{}, err
	}

	defer d.Close()

	read, err := d.Readdir(-1)
	if err != nil {
		return []string{}, err
	}

	for _, file := range read {
		if file.Mode().IsDir() {
			imports, err := getImportsInDirectory(dir + "/" + file.Name())
			if err != nil {
				return imports, err
			}

			for _, i := range imports {
				final = appendIfMissing(final, i)
			}

		} else if file.Mode().IsRegular() && strings.HasSuffix(file.Name(), ".go") {
			f, err := os.Open(dir + "/" + file.Name())
			if err != nil {
				return []string{}, err
			}
			defer f.Close()

			buf := &bytes.Buffer{}
			_, err = buf.ReadFrom(f)
			if err != nil {
				return []string{}, err
			}

			m := match.Find(buf.Bytes())
			if len(m) > 0 {
				spl := strings.Split(string(m), "\n")
				for _, s := range spl {
					if strings.Contains(s, ".") {
						quoteless := inQuotes.FindString(s)
						if len(quoteless) > 0 {
							final = appendIfMissing(final, strings.Replace(quoteless, "\"", "", -1))
						}
					}
				}
			}
		}
	}

	return final, nil
}

func main() {
	gopath := os.Getenv("GOPATH")
	if len(gopath) == 0 {
		log.Fatal("$GOPATH not defined!")
	}

	gofile, err := os.OpenFile("Gofile", os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0660)
	if err != nil {
		log.Fatal(err)
	}

	defer gofile.Close()

	fmt.Println("Scanning directories...\n")

	// Run from current working directory

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	imports, err := getImportsInDirectory(wd)
	if err != nil {
		log.Fatal(err)
	}

	for _, i := range imports {
		if wd == gopath+"/src/"+i {
			continue
		}
		err = os.Chdir(gopath + "/src/" + i)
		if err != nil {
			log.Fatal(err)
		}

		// Make sure it's the .git root (has a .git folder)
		_, err = os.Stat(gopath + "/src/" + i + "/.git")
		if err != nil {
			if _, ok := err.(*os.PathError); ok {
				continue
			}
			log.Fatal(err)
		}

		out, err := exec.Command("git", "rev-parse", "HEAD").Output()
		if err != nil {

			log.Fatal(err)
		}

		fmt.Print(i + " " + string(out))
		gofile.Write([]byte(i + " " + string(out)))
	}

	fmt.Println("\nDone!")

}
