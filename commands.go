package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/urfave/cli"
)

const FiveGB = 5000000000
const InMB = 100000000

var downloadPartLimitInBytes = FiveGB

var wg sync.WaitGroup

func AppendFileChunks() func(c *cli.Context) error {
	return func(c *cli.Context) error {

		fromPart, parts, _, path, _, _, _ := computeParts(c)

		appendParts(fromPart, parts, path)

		return nil
	}
}

func FetchFile() func(c *cli.Context) error {
	return func(c *cli.Context) error {

		fromPart, parts, length, path, filename, urlInput, concurrency := computeParts(c)

		downloadMultipart(fromPart, length, parts, urlInput, filename, path, concurrency)

		appendParts(fromPart, parts, path)

		return nil
	}
}

func computeParts(c *cli.Context) (fromPart int, parts int, length int, path string, filename string, urlInput string, concurrency int) {

	var errWd error
	path, errWd = os.Getwd()
	if errWd != nil {
		ErrorLog("Could not get current working directory")
		return
	}

	fromPart = c.Int("p") // index part from where to start

	concurrency = c.Int("c")
	if concurrency == 0 {
		concurrency = 1
	}

	urlInput = c.String("u")
	if urlInput == "" {
		ErrorLog("Missing url (please provide a url from where to fetch a file via -u flag)")
		return
	}

	filename = c.String("f")
	if filename == "" {
		ErrorLog("Missing filename (please provide a filename via -f flag)")
		return
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
		return
	}

	// var directory string = path
	path = filepath.Join(path, filename)

	InfoLog("\nDownloading %s ...\n\n", urlInput)

	client := &http.Client{}
	req, _ := http.NewRequest("GET", urlInput, nil)
	range_header := "bytes=0-0"
	req.Header.Add("Range", range_header)
	resp, _ := client.Do(req)

	if server, ok := resp.Header["Server"]; ok {
		if len(server) != 0 {
			val := server[0]
			if val != "AmazonS3" {
				ErrorLog("Could not download file, invalid file server")
				return
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

	var errContentRange error

	if contentRange, ok := resp.Header["Content-Range"]; ok {
		if len(contentRange) != 0 {
			val := contentRange[0]
			if val != "" {
				lengthString := strings.Replace(val, "bytes 0-0/", "", 1)
				length, errContentRange = strconv.Atoi(lengthString)
				if errContentRange != nil {
					ErrorLog("Could not get file content range")
					return
				} else {
					SimpleLog("Size: %s\n", humanize.Bytes(uint64(length)))
				}
			}
		}
	} else {
		ErrorLog("File does not exist\n")
		return
	}

	// compute parts
	parts = length / downloadPartLimitInBytes

	// change download part limit in bytes
	// this is just for testing pruposes, to test files that
	// are enough big to big chunked but not too big to make the test too long
	// (e.g 1GB)
	if length < downloadPartLimitInBytes {
		downloadPartLimitInBytes = InMB
	}

	// re-compute parts
	parts = length / downloadPartLimitInBytes

	// download it all at once no need to download in parts
	// in case file size is less than download part limit in bytes
	if length < downloadPartLimitInBytes {
		parts = 1
	}

	if fromPart > parts {
		ErrorLog("Part from where to start cannot be higher than total parts amount\n")
		return
	}

	if concurrency > parts {
		concurrency = parts
	}

	if concurrency > 10 {
		concurrency = 10
	}

	return
}

func downloadMultipart(fromPart int, length int, parts int, urlInput string, filename string, path string, concurrency int) {

	started := time.Now()

	diff := length % downloadPartLimitInBytes
	lensub := downloadPartLimitInBytes

	if parts == 1 {
		diff = 0
		lensub = length
	}

	SimpleLog("Will fetch file %s from part %d to part %d of about %s and reminder of about %s ...\n", filename, fromPart, parts, humanize.Bytes(uint64(lensub)), humanize.Bytes(uint64(diff)))
	SimpleLog("Downloaded started at %s\n", started.Format(time.RFC3339))
	wg.Add(parts - fromPart)
	waitChan := make(chan struct{}, concurrency) // max concurrent part downloads
	for i := fromPart; i < parts; i++ {
		waitChan <- struct{}{}
		go func(count int) {
			defer wg.Done()
			downloadAndSavePart(urlInput, path, count, parts, lensub, diff)
			<-waitChan
		}(i)
	}
	wg.Wait()

	ended := time.Now()

	SimpleLog("\nDownloaded finished at %s\n", ended.Format(time.RFC3339))

	if ended.Sub(started).Seconds() < 60 {
		SimpleLog("Download took %f seconds to complete\n", ended.Sub(started).Seconds())
	} else if ended.Sub(started).Minutes() < 60 {
		SimpleLog("Download took %f minutes to complete\n", ended.Sub(started).Minutes())
	} else {
		SimpleLog("Download took %f hours to complete\n", ended.Sub(started).Hours())
	}

	SuccessLog("File downloaded successfully ...\n")
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
	range_header := "bytes=" + strconv.Itoa(min) + "-" + strconv.Itoa(max-1)
	req.Header.Add("Range", range_header)
	resp, errGet := client.Do(req)

	if errGet != nil {
		ErrorLog("[ERROR] Could not download chunk %d - %s ??? %s \n", i, humanize.Bytes(uint64(min)), humanize.Bytes(uint64(max-1)))
	}
	defer resp.Body.Close()

	if errGet == nil {
		if cache, ok := resp.Header["X-Cache"]; ok {
			if len(cache) != 0 {
				val := cache[0]
				if strings.Contains(val, "Hit from cloudfront") {
					SimpleLog("[%d] Downloading chunk from edge cache %s ??? %s ...\n", i, humanize.Bytes(uint64(min)), humanize.Bytes(uint64(max-1)))
				} else {
					SimpleLog("[%d] Downloading chunk from origin %s ??? %s ...\n", i, humanize.Bytes(uint64(min)), humanize.Bytes(uint64(max-1)))
				}
			}
		} else {
			SimpleLog("[%d] Downloading chunk %s ??? %s ...\n", i, humanize.Bytes(uint64(min)), humanize.Bytes(uint64(max-1)))
		}

		writer := bufio.NewWriter(file)
		io.Copy(writer, resp.Body)
		writer.Flush()
	}
}

func appendParts(fromPart int, parts int, path string) {

	WarningLog("[WARNING] Now merging chunks, this may take some time, do not exit this process otherwise all progress will be lost.")

	checkChunkExistence := make([]string, 0, 0)

	// quickly check if there are all parts
	for i := fromPart; i < parts; i++ {
		f := fmt.Sprintf("%s.%d", path, i)
		info, err := os.Stat(f)
		if err != nil {
			checkChunkExistence = append(checkChunkExistence, fmt.Sprintf("failed to read %s: %s", f, err.Error()))
			return
		}
		if errors.Is(err, os.ErrNotExist) {
			checkChunkExistence = append(checkChunkExistence, fmt.Sprintf("file does not exists %s", f))
			return
		}
		if info.IsDir() {
			checkChunkExistence = append(checkChunkExistence, fmt.Sprintf("file should not be directory %s", f))
			return
		}
	}

	if len(checkChunkExistence) > 0 {
		errorCheckExistence := strings.Join(checkChunkExistence, ", ")
		ErrorLog(errorCheckExistence)
		return
	}

	out, errTotalFile := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if errTotalFile != nil {
		ErrorLog("failed to create file %s: %s", path, errTotalFile)
		return
	}
	// defer out.Close()

	for i := fromPart; i < parts; i++ {

		f := fmt.Sprintf("%s.%d", path, i)

		val, err := os.Open(f)
		if err != nil {
			ErrorLog("failed to read %s: %s", f, err.Error())
			break
		}

		_, errMergeChunk := io.Copy(out, val)
		if errMergeChunk != nil {
			defer val.Close()
			ErrorLog("failed to append chunk %s: %s", f, errMergeChunk.Error())
			break
		}

		// close file then remove
		val.Close()
		errDeleteChunk := os.Remove(f)
		if errDeleteChunk != nil {
			ErrorLog("failed to delete chunk %s: %s", f, errDeleteChunk.Error())
			break
		}

		SimpleLog("Chunk %d/%d was merged\n", i+1, parts)
	}

	stat, errStat := out.Stat()
	if errStat != nil {
		ErrorLog("failed to get file information %s: %s", path, errStat)
		return
	}

	SimpleLog("File Size: %d\n", stat.Size())
	SimpleLog("Process has now finished\n")
}
