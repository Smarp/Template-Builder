package main

import (
    "os"
    "fmt"
    "io"
    "time"
    "bufio"
    "bytes"
    "strings"
    "path/filepath"
    "github.com/andreaskoch/go-fswatch"
)

var buildFolder = "build/"

// Constants
const (
    checkIntervalInSeconds = 2
    templateToken = "#>>"
)

// Start watcher
func Start(templateFolder string) {

    // Force building at startup
    buildTemplates(os.Args[1])


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
                    fmt.Println(time.Now().Format(time.Stamp) + ": " + changes.New()[0] + "...")
                    buildTemplates(templateFolder)
                }

                case len(changes.Modified()) > 0 : {
                    fmt.Println(time.Now().Format(time.Stamp) + ": " + changes.Modified()[0] + "...")
                    buildTemplates(templateFolder)
                }

            } 
        }
    }

}

// Execute main
func executeFile(path string, folder string) string {

    file, _ := os.Open(path)

    var buffer bytes.Buffer

    scanner := bufio.NewScanner(file)

    for scanner.Scan() {
        line := scanner.Text()

        if strings.HasPrefix(strings.TrimSpace(line), templateToken) {
            buffer.WriteString(executePartial(strings.Replace(strings.TrimSpace(line), templateToken, "", -1), folder) + "\n")
            
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

    buffer.WriteString("\n<!-- " + path +" // -->\n")
    for scanner.Scan() {
        line := scanner.Text()

        if strings.HasPrefix(strings.TrimSpace(line), templateToken) {
            buffer.WriteString(executePartial(strings.Replace(strings.TrimSpace(line), templateToken, "", -1), folder) + "\n")
        } else {
            buffer.WriteString(line + "\n")
        }
    }
    buffer.WriteString("<!-- // " + path +" -->\n")

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
    files, _ := filepath.Glob(templateFolder + "*")
    
    // Create output folder
    os.Mkdir(buildFolder, 0777)

    for i := 0; i < len(files); i++ {
        if isFile(files[i]) {
            content := executeFile(files[i], templateFolder)

            f, err := os.Create(buildFolder + filepath.Base(files[i]))

            if err != nil {
                fmt.Println(err)
            }
            
            n, err := io.WriteString(f, content)

            if err != nil {
                fmt.Println(n, err)
            }
            f.Close()
        }
    }

    fmt.Println(time.Now().Format(time.Stamp) + ": Build complete!")
    fmt.Println("===============================")

}

func main() {

    if len(os.Args) == 1 {
        fmt.Println("Usage: go run template_builder.go <TEMPLATE_FOLDER> [BUILD_FOLDER]")
        os.Exit(0)
    }
    if len(os.Args) > 2 {
        buildFolder = os.Args[2]
        fmt.Println("Folder changed to: " + buildFolder)
    }

    Start(os.Args[1])

}