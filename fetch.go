package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/dustin/go-humanize"
	"github.com/urfave/cli"
)

const FiveGB = 5000000000
const TenMB = 10000000

const downloadPartLimitInBytes = TenMB

var wg sync.WaitGroup

func FetchFile() func(c *cli.Context) error {
	return func(c *cli.Context) error {

		path, errWd := os.Getwd()
		if errWd != nil {
			ErrorLog("Could not get current working directory")
			return nil
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
					return nil
				}
			}
		}

		if etag, ok := resp.Header["Etag"]; ok {
			if len(etag) != 0 {
				val := etag[0]
				if val != "" {
					WarningLog("Etag: %s\n\n", strings.Replace(val, "\"", "", -1))
				}
			}
		}

		if lastModified, ok := resp.Header["Last-Modified"]; ok {
			if len(lastModified) != 0 {
				val := lastModified[0]
				if val != "" {
					WarningLog("Last Modified: %s\n\n", val)
				}
			}
		}

		var length int
		var errContentRange error

		if contentRange, ok := resp.Header["Content-Range"]; ok {
			if len(contentRange) != 0 {
				val := contentRange[0]
				if val != "" {
					lengthString := strings.Replace(val, "bytes 0-0/", "", 1)
					length, errContentRange = strconv.Atoi(lengthString)
					if errContentRange != nil {
						ErrorLog("Could not get file content range")
						return nil
					} else {
						WarningLog("Size: %s\n\n", humanize.Bytes(uint64(length)))
					}
				}
			}
		} else {
			ErrorLog("File does not exist\n")
			return nil
		}

		// downloadWithoutResume(length, urlInput, filename, path)

		downloadWithParty(length, urlInput, filename, path)

		return nil
	}
}

func downloadWithParty(length int, urlInput string, filename string, path string) {

	// compute parts
	parts := length / downloadPartLimitInBytes
	// diff := length % downloadPartLimitInBytes
	// lensub := downloadPartLimitInBytes

	// download it all at once no need to download in parts
	// in case file size is less than download part limit in bytes
	if length < downloadPartLimitInBytes {
		parts = 1
		// diff = 0
		// lensub = length
	}

	runtime.MemProfileRate = 0
	ctx, cancel := backgroundContext()
	defer cancel()

	cmd := &Cmd{
		Ctx: ctx,
		Out: os.Stdout,
		Err: os.Stderr,
	}

	arguments := make([]string, 0, 0)
	arguments = append(arguments, urlInput)
	arguments = append(arguments, fmt.Sprintf("-p=%d", parts)) // parts
	arguments = append(arguments, fmt.Sprintf("-r=%d", 3))     // max retries per part
	arguments = append(arguments, fmt.Sprintf("-t=%d", 15))    // timeout
	arguments = append(arguments, fmt.Sprintf("-o=%s", path))  // output path
	// arguments = append(arguments, fmt.Sprintf("-s=download.json")) // output path for download session

	if err := cmd.Run(arguments, "dev", "xxxxx"); err != nil {
		ErrorLog("Error: %s", err)
	}
}

func backgroundContext() (context.Context, func()) {
	ctx, cancel := context.WithCancel(context.Background())
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		defer signal.Stop(quit)
		<-quit
		cancel()
	}()

	return ctx, cancel
}

func downloadWithoutResume(length int, urlInput string, filename string, path string) error {

	// compute parts
	parts := length / downloadPartLimitInBytes
	diff := length % downloadPartLimitInBytes
	lensub := downloadPartLimitInBytes

	// download it all at once no need to download in parts
	// in case file size is less than download part limit in bytes
	if length < downloadPartLimitInBytes {
		parts = 1
		diff = 0
		lensub = length
	}

	SuccessLog("Will fetch file %s in %d parts of about %s and reminder of about %s ...\n\n", filename, parts, humanize.Bytes(uint64(lensub)), humanize.Bytes(uint64(diff)))

	// create empty file
	file, errFileCreate := os.Create(path)
	if errFileCreate != nil {
		ErrorLog("Could not create file: %s", errFileCreate)
		return nil
	}

	defer file.Close()

	wg.Add(parts)

	for i := 0; i < parts; i++ {

		min := lensub * i
		max := lensub * (i + 1)

		if i == parts-1 {
			max += diff
		}

		client := &http.Client{}
		req, _ := http.NewRequest("GET", urlInput, nil)
		range_header := "bytes=" + strconv.Itoa(min) + "-" + strconv.Itoa(max-1)
		req.Header.Add("Range", range_header)
		resp, _ := client.Do(req)
		defer resp.Body.Close()

		if cache, ok := resp.Header["X-Cache"]; ok {
			if len(cache) != 0 {
				val := cache[0]
				if strings.Contains(val, "Hit from cloudfront") {
					SimpleLog("[%d] Downloading chunk from edge cache %s → %s ...\n", i+1, humanize.Bytes(uint64(min)), humanize.Bytes(uint64(max-1)))
				} else {
					SimpleLog("[%d] Downloading chunk from origin %s → %s ...\n", i+1, humanize.Bytes(uint64(min)), humanize.Bytes(uint64(max-1)))
				}
			}
		} else {
			SimpleLog("[%d] Downloading chunk %s → %s ...\n", i+1, humanize.Bytes(uint64(min)), humanize.Bytes(uint64(max-1)))
		}

		writer := bufio.NewWriter(file)
		io.Copy(writer, resp.Body)
		writer.Flush()

		wg.Done()
	}

	wg.Wait()
	SuccessLog("\nFile downloaded successfully ...\n")

	return nil
}
