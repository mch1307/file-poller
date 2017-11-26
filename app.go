package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kardianos/service"
	"github.com/mch1307/file-poller/config"
	"github.com/mch1307/file-poller/log"
)

var logger service.Logger

var srcLevel int
var conf string
var dirScanEntries []dirScanEntry
var ticker = time.NewTicker(10 * time.Second)
var quit chan struct{}

type dirScanEntry struct {
	srcFile  string
	destFile string
	fileSize int64
	fileTS   time.Time
}

func main() {
	fmt.Println("Starting app...")
	daemon()
}

func makeDirScanEntries() []dirScanEntry {
	var res []dirScanEntry
	return res
}

func init() {
	flag.StringVar(&conf, "conf", "", "Config file name (full path to TOML file)")
	flag.Parse()
	config.Initialize(conf)
	log.Initialize()
	log.Info("using config: ", config.Conf)
	srcLevel = dirLevel(config.Conf.SourceDir)
	quit = make(chan struct{})
}

func dirLevel(p string) int {
	cnt := strings.Count(p, string(filepath.Separator))
	return cnt
}

// isFileOK checks if file can be read and was not modified in the last 30 seconds
func isFileOK(p string) bool {
	var checkOK bool
	f, err := os.OpenFile(p, os.O_EXCL, 0)
	if err != nil {
		log.Warn("file locked: ", p)
		return false
	}
	checkOK = true
	defer f.Close()

	fInfo, err := os.Stat(p)
	if err != nil {
		log.Error("file check error ", err)
		return false
	}

	if time.Since(fInfo.ModTime()) > time.Second*30 && checkOK {
		return true
	}
	return false
}

// prepareMove populates the dirScanEntries arrays that will be used for moving files
func prepareMove(path string, info os.FileInfo, err error) error {
	log.Debug("scanned: ", path)
	if !info.IsDir() {
		// Level 2 from root is an expected fax file
		if dirLevel(path)-srcLevel == 2 {
			log.Debug("considering file ", path)
			// get directory as it means the fax extension
			targetPrefix := filepath.Base(filepath.Dir(path))
			// target filename include fax extension for further routing
			targetFile := filepath.Join(config.Conf.DestDir, targetPrefix+"_"+info.Name())
			if isFileOK(path) {
				log.Debug("prepareMove file id ok: ", path)
				var entry dirScanEntry
				entry.srcFile = path
				entry.destFile = targetFile
				fInfo, err := os.Stat(path)
				if err != nil {
					fmt.Println("error ", err)
				}
				entry.fileTS = fInfo.ModTime()
				entry.fileSize = fInfo.Size()
				dirScanEntries = append(dirScanEntries, entry)
				log.Debug("adding to list of files to process: ", entry)
			}
		}
	}
	return nil
}

func rename(src, dest string) error {
	from, err := os.Open(src)
	if err != nil {
		log.Fatal(err)
	}
	to, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE, os.ModeExclusive)
	if err != nil {
		log.Fatal(err)
	}
	defer to.Close()

	_, err = io.Copy(to, from)
	if err != nil {
		log.Fatal(err)
	}
	from.Close()
	err = os.Remove(src)
	if err != nil {
		log.Fatal("error removing src: ", src, err)
	}
	return err
}

func daemon() {
	for {
		select {
		case <-ticker.C:
			dirScanEntries = makeDirScanEntries()
			_ = filepath.Walk(config.Conf.SourceDir, prepareMove)
			fmt.Println(dirScanEntries)
			for _, v := range dirScanEntries {
				rename(v.srcFile, v.destFile)
				log.Info("processing file ", v.srcFile, " to ", v.destFile, ", original timestamp: ", v.fileTS)
			}
		case <-quit:
			ticker.Stop()
			log.Info("stopping...")
			return
		}
	}
}
