package main

import (
	"os"
	"fmt"
	"io"
    "bufio"
    "bytes"
	"strings"
	"path/filepath"
	"github.com/andreaskoch/go-fswatch"
)

// Constants
const (
    checkIntervalInSeconds = 2
    templateToken = "#>>"
    buildFolder = "build/"
)

// Start watcher
func Start(templateFolder string) {

    recurse := true // include all sub directories

    skipDotFilesAndFolders := func(path string) bool {
        return strings.HasPrefix(filepath.Base(path), ".")
    }


    folderWatcher := fswatch.NewFolderWatcher(
        templateFolder,
        recurse,
        skipDotFilesAndFolders,
        checkIntervalInSeconds)

    folderWatcher.Start()

    for folderWatcher.IsRunning() {

        select {

        case changes := <-folderWatcher.ChangeDetails():

        	switch {
        		case len(changes.New()) > 0: {
        			buildTemplates(templateFolder)
				}

        		case len(changes.Modified()) > 0 : {
		            buildTemplates(templateFolder)
		        }

        	} 
        }
    }

}

// Execute main
func executeFile(path string, folder string) string {

	fmt.Println("Executing: " + path)
 	file, _ := os.Open(path)

    var buffer bytes.Buffer

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
    	line := scanner.Text()

    	if strings.HasPrefix(line, templateToken) {
    		buffer.WriteString(executePartial(strings.Replace(line, templateToken, "", -1), folder) + "\n")
    	} else {
    		buffer.WriteString(line + "\n")
    	}
    }

  file.Close()
  return buffer.String()
}

// Execute Partial file
func executePartial(path string, folder string) string {
 	file, _ := os.Open(folder + "/" + path)

    var buffer bytes.Buffer

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
    	line := scanner.Text()

    	if strings.HasPrefix(line, templateToken) {
    		buffer.WriteString(executePartial(strings.Replace(line, templateToken, "", -1), folder) + "\n")
    	} else {
    		buffer.WriteString(line + "\n")
    	}
    }

  file.Close()
  return buffer.String()
}

func isFile(path string) bool {
    f, err := os.Open(path)

    if err != nil {
        fmt.Println(err)
        return false
    }
    defer f.Close()

    fi, err := f.Stat()

    if err != nil {
        fmt.Println(err)
        return false
    }
    switch mode := fi.Mode(); {
	    case mode.IsDir():
	        return false
	    case mode.IsRegular():
	        return true
    }

    return false
}


func buildTemplates(templateFolder string) {

	// Go throuch each template
	files, _ := filepath.Glob(templateFolder + "/*")
	fmt.Println(files)
	
	// Create output folder
	os.MkdirAll(buildFolder + templateFolder, 0777)

	for i := 0; i < len(files); i++ {
		if isFile(files[i]) {
			content := executeFile(files[i], templateFolder)

			f, err := os.Create(buildFolder + files[i])

			if err != nil {
			 	fmt.Println(err)
			}
	  		
	  		n, err := io.WriteString(f, content)

		  	if err != nil {
		   		fmt.Println(n, err)
		  	}
		  	f.Close()

			fmt.Println("---------------------")
		}
	}




}

func main() {

	Start(os.Args[1])

}