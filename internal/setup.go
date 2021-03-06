package internal

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Status struct {
	Name     string
	Message  string
	Progress string
	Size     string
}

var (
	toolPath          = filepath.ToSlash(os.Getenv("APPDATA") + "/Easy2Burst/")
	downloadCachePath = toolPath + "downloadCache/"
	burstCmdPath      = toolPath + "BurstWallet"
	relevantFileNames = []string{"AppInfo.xml", "BurstWallet", "MariaDB"}
	downloadUrl       = "https://download.cryptoguru.org/burst/qbundle/Easy2Burst/"
	stat              = Status{
		Name:     "starting",
		Message:  "Starting Setup...",
		Progress: "0%",
		Size:     "",
	}
)

func CheckTools(statusCh chan Status, commandCh chan string) {
	//set logs
	os.Remove(toolPath + "startup.log")
	logFile, _ := os.OpenFile(toolPath+"startup.log", os.O_WRONLY|os.O_CREATE, 0644)
	defer logFile.Close()
	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags | log.Llongfile)
	log.Print("Started Easy2Burst Setup")
	statusCh <- stat
	listOfMissingFiles := checkFileExistences()
	if len(listOfMissingFiles) > 0 {
		stat.Name = "startSetup"
		statusCh <- stat
		log.Printf("Files %s missing.", listOfMissingFiles)
		processFiles(listOfMissingFiles, statusCh)
	}
	stat.Name = "checkBurstDB"
	statusCh <- stat
	CheckBurstDB()
	stat.Name = "setupFinished"
	statusCh <- stat
	CheckForUpdates(statusCh)
	stat.Name = "updaterFinished"
	statusCh <- stat
	StartWallet(statusCh, commandCh)
	stat.Name = "walletStarted"
	statusCh <- stat
}

func checkFileExistences() []string {
	missingFiles := make([]string, 0)
	for _, fileName := range relevantFileNames {
		if _, err := os.Stat(toolPath + fileName); os.IsNotExist(err) {
			missingFiles = append(missingFiles, fileName)
		}
	}
	return missingFiles
}

func processFiles(missingFiles []string, statusCh chan Status) {
	for _, fileName := range missingFiles {
		if fileName != "AppInfo.xml" {
			DownloadFile(downloadCachePath+fileName+".zip", downloadUrl+fileName+".zip", statusCh)
			unzip(downloadCachePath+fileName+".zip", toolPath+fileName, statusCh)
		} else {
			DownloadFile(downloadCachePath+fileName, downloadUrl+fileName, statusCh)
			CopyFile(downloadCachePath+fileName, toolPath+fileName)
		}
	}
	os.RemoveAll(downloadCachePath)
}

func CheckBurstDB() {
	//Write the path of the .sql into the BAT file
	file, err := ioutil.ReadFile(toolPath + "MariaDB/bin/setupDb.BAT")
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	newContents := strings.Replace(string(file), "{PATH_TO_SQL_SCRIPT}", toolPath+"BurstWallet/init-mysql.sql", -1)
	err = ioutil.WriteFile(toolPath+"MariaDB/bin/setupDb.BAT", []byte(newContents), 0)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	cmd := exec.Command(toolPath + "MariaDB/bin/setupDb.BAT")
	cmd.Env = append(os.Environ())
	out, err := cmd.CombinedOutput()
	log.Print(string(out))
	if err != nil {
		log.Fatal(err)
	}
}
