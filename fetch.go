package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/dustin/go-humanize"
	"github.com/urfave/cli"
)

const FiveGB = 5000000000
const InMB = 100000000

const downloadPartLimitInBytes = InMB

var wg sync.WaitGroup

func FetchFile() func(c *cli.Context) error {
	return func(c *cli.Context) error {

		path, errWd := os.Getwd()
		if errWd != nil {
			ErrorLog("Could not get current working directory")
			return nil
		}

		fromPart := c.Int("p") // index part from where to start

		concurrency := c.Int("c")
		if concurrency == 0 {
			concurrency = 3
		}

		urlInput := c.String("u")
		if urlInput == "" {
			ErrorLog("Missing url (please provide a url from where to fetch a file via -u flag)")
			return nil
		}

		filename := c.String("f")
		if filename == "" {
			ErrorLog("Missing filename (please provide a filename via -f flag)")
			return nil
		}

		// if there is an output path, override the current working directory output path
		output := c.String("o")
		if output != "" {
			path = output
		}

		// check url validiy
		_, errParseUrl := url.ParseRequestURI(urlInput)
		if errParseUrl != nil {
			ErrorLog("Invalid input url")
			return nil
		}

		// var directory string = path
		path = filepath.Join(path, filename)

		InfoLog("\nDownloading %s ...\n\n", urlInput)

		client := &http.Client{}
		req, _ := http.NewRequest("GET", urlInput, nil)
		range_header := "bytes=0-0"
		req.Header.Add("Range", range_header)
		resp, _ := client.Do(req)

		// for k, v := range resp.Header {
		// 	log.Print(k)
		// 	log.Print(v)
		// }

		if server, ok := resp.Header["Server"]; ok {
			if len(server) != 0 {
				val := server[0]
				if val != "AmazonS3" {
					ErrorLog("Could not download file, invalid file server")
					return nil
				}
			}
		}

		if etag, ok := resp.Header["Etag"]; ok {
			if len(etag) != 0 {
				val := etag[0]
				if val != "" {
					SimpleLog("Etag: %s\n", strings.Replace(val, "\"", "", -1))
				}
			}
		}

		if lastModified, ok := resp.Header["Last-Modified"]; ok {
			if len(lastModified) != 0 {
				val := lastModified[0]
				if val != "" {
					SimpleLog("Last Modified: %s\n", val)
				}
			}
		}

		var length int
		var errContentRange error

		if contentRange, ok := resp.Header["Content-Range"]; ok {
			if len(contentRange) != 0 {
				log.Print(contentRange[0])
				val := contentRange[0]
				if val != "" {
					lengthString := strings.Replace(val, "bytes 0-0/", "", 1)
					length, errContentRange = strconv.Atoi(lengthString)
					if errContentRange != nil {
						ErrorLog("Could not get file content range")
						return nil
					} else {
						SimpleLog("Size: %s\n", humanize.Bytes(uint64(length)))
					}
				}
			}
		} else {
			ErrorLog("File does not exist\n")
			return nil
		}

		// compute parts
		parts := length / downloadPartLimitInBytes

		downloadWithCustomMultipart(fromPart, length, parts, urlInput, filename, path)

		// Create final file
		out, errTotalFile := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
		if errTotalFile != nil {
			ErrorLog("failed to create file %s: %s", path, errTotalFile)
			return nil
		}
		defer out.Close()

		for i := 0; i < parts; i++ {

			f := fmt.Sprintf("%s.%d", path, i)
			val, err := os.Open(f)
			if err != nil {
				ErrorLog("failed to read %s: %s", f, err.Error())
				break
			}
			defer val.Close()

			_, errMergeChunk := io.Copy(out, val)
			if errMergeChunk != nil {
				ErrorLog("failed to append chunk %s: %s", f, errMergeChunk.Error())
				break
			}

			errDeleteChunk := os.Remove(f)
			if errDeleteChunk != nil {
				ErrorLog("failed to delete chunk %s: %s", f, errDeleteChunk.Error())
				break
			}

		}

		return nil
	}
}

func downloadWithCustomMultipart(fromPart int, length int, parts int, urlInput string, filename string, path string) {

	diff := length % downloadPartLimitInBytes
	lensub := downloadPartLimitInBytes

	// download it all at once no need to download in parts
	// in case file size is less than download part limit in bytes
	// if length < downloadPartLimitInBytes {
	// 	parts = 1
	// 	diff = 0
	// 	lensub = length
	// }

	SimpleLog("Will fetch file %s in %d parts of about %s and reminder of about %s ...\n\n", filename, parts, humanize.Bytes(uint64(lensub)), humanize.Bytes(uint64(diff)))
	wg.Add(parts - fromPart)
	waitChan := make(chan struct{}, 3) // max concurrent part downloads
	for i := fromPart; i < parts; i++ {
		waitChan <- struct{}{}
		go func(count int) {
			defer wg.Done()
			downloadAndSavePart(urlInput, path, count, parts, lensub, diff)
			<-waitChan
		}(i)
	}
	wg.Wait()
	SuccessLog("\nFile downloaded successfully ...\n")
}

func downloadAndSavePart(urlInput string, path string, i int, parts int, lensub int, diff int) {

	min := lensub * i
	max := lensub * (i + 1)

	if i == parts-1 {
		max += diff
	}

	fileName := fmt.Sprintf("%s.%d", path, i)
	// create empty file
	file, errFileCreate := os.Create(fileName)
	if errFileCreate != nil {
		ErrorLog("Could not create file: %s", errFileCreate)
		return
	}

	client := &http.Client{}
	req, _ := http.NewRequest("GET", urlInput, nil)
	range_header := "bytes=" + strconv.Itoa(min) + "-" + strconv.Itoa(max)
	req.Header.Add("Range", range_header)
	resp, errGet := client.Do(req)

	if errGet != nil {
		ErrorLog("[ERROR] Could not download chunk %d - %s → %s \n", i, humanize.Bytes(uint64(min)), humanize.Bytes(uint64(max-1)))
	}
	defer resp.Body.Close()

	if errGet == nil {
		if cache, ok := resp.Header["X-Cache"]; ok {
			if len(cache) != 0 {
				val := cache[0]
				if strings.Contains(val, "Hit from cloudfront") {
					SimpleLog("[%d] Downloading chunk from edge cache %s → %s ...\n", i, humanize.Bytes(uint64(min)), humanize.Bytes(uint64(max-1)))
				} else {
					SimpleLog("[%d] Downloading chunk from origin %s → %s ...\n", i, humanize.Bytes(uint64(min)), humanize.Bytes(uint64(max-1)))
				}
			}
		} else {
			SimpleLog("[%d] Downloading chunk %s → %s ...\n", i, humanize.Bytes(uint64(min)), humanize.Bytes(uint64(max-1)))
		}

		writer := bufio.NewWriter(file)
		io.Copy(writer, resp.Body)
		writer.Flush()
	}

}
