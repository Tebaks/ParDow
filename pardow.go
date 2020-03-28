package pardow

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

var wg sync.WaitGroup

type info struct {
	url       string
	totalSize int64
	file      *os.File
}

// ParallelDownload is Downloading a single file with multiple goroutines in same time.
func ParallelDownload(url string, goroutineCount int64, path string, saveName string) {
	fileSize, err := getFileSize(url)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	log.Printf("File Size : %d ", fileSize)
	localPath := "./" + saveName + filepath.Ext(url)
	f, err := os.OpenFile(localPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer f.Close()

	var info = info{
		url:       url,
		totalSize: fileSize,
		file:      f,
	}

	var start, end int64

	var partSize = int64(fileSize / goroutineCount)
	fmt.Println("Downloading...")
	for num := int64(0); num < goroutineCount; num++ {
		if num == goroutineCount {
			end = fileSize
		} else {
			end = start + partSize
		}
		wg.Add(1)
		go info.partialDownload(num, start, end-1)
		start = end

	}
	wg.Wait()
	fmt.Println("Download completed.")

}

func (i *info) partialDownload(number, start, end int64) {
	var completed int64
	body, size, err := i.getBody(start, end)
	if err != nil {
		log.Fatalf("Partial Download Error : %s", err)
	}
	defer body.Close()
	defer wg.Done()
	buffer := make([]byte, 4*1024)
	for {
		read, readErr := body.Read(buffer)
		if read > 0 {
			write, writeErr := i.file.WriteAt(buffer[0:read], start)
			if writeErr != nil {
				log.Println(read)
				log.Fatalf("Error occurs when writing: %s", writeErr)
			}
			if read != write {
				log.Fatal("Read not equals to write")
			}

			start = int64(write) + start
			if write > 0 {
				completed += int64(write)
				fmt.Printf("%d downloaded by %d \n", completed, number)
			}
		}
		if readErr != nil {
			if readErr.Error() == "EOF" {
				//End Of File
				if size == completed {
					//Part Downloaded
				} else {
					log.Fatalf("Error occors when reading: %s", readErr)
				}
				break
			}
		}
	}
}
func (i *info) getBody(start, end int64) (io.ReadCloser, int64, error) {
	var client http.Client
	req, err := http.NewRequest("GET", i.url, nil)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	req.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", start, end))
	res, err := client.Do(req)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	size, err := strconv.ParseInt(res.Header["Content-Length"][0], 10, 64)
	return res.Body, size, err
}

func getFileSize(url string) (size int64, err error) {
	client := http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}
	res, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	header := res.Header
	size, err = strconv.ParseInt(header["Content-Length"][0], 10, 64)
	return size, err

}
