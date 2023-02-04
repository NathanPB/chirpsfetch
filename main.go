/*
 * Copyright 2023 Nathan P. Bombana
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 *
 */

package main

import (
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

var dateFlag = flag.String("date", "", "The date or date range to be fetched (e.g. 2022-01-01 or 2022-01-01..2022-01-31)")
var pollSizeFlag = flag.Int("poll-size", 128, "The number of records to be fetched and insert at once. Must be greater than 0, only affects date ranges")
var saveFlag = flag.String("save", "", "Save the downloaded files. Takes a relative path. If not specified, prints to stdout")
var attemptsFlag = flag.Int("attemptsFlag", 3, "The number of attempts to be made to fetch the data")
var noGunzipFlag = flag.Bool("no-gunzip", false, "Do not gunzip the downloaded files")
var silentFlag = flag.Bool("silent", false, "Do not print the output of the command. Only works if not using --save and --date with range simultaneously, otherwise the program is silent by default")
var precisionFlag = flag.String("precision", "p05", "The precision of the data. Can be either p05 or p25")

var regexDate = regexp.MustCompile("^\\d{4}-(0[1-9]|1[0-2])-([0-2][1-9]|[1-3]0|3[01])$")
var regexDateRange = regexp.MustCompile("^\\d{4}-(0[1-9]|1[0-2])-([0-2][1-9]|[1-3]0|3[01])\\.\\.\\d{4}-(0[1-9]|1[0-2])-([0-2][1-9]|[1-3]0|3[01])$")

type closingReader struct {
	io.Reader
	closer io.Closer
}

func (r *closingReader) Close() error {
	return r.closer.Close()
}

func main() {
	flag.Parse()

	if *precisionFlag != "p05" && *precisionFlag != "p25" {
		panic(fmt.Errorf("invalid precision: %s", *precisionFlag))
	}

	if *dateFlag == "" {
		panic("No --date defined")
	}

	if regexDate.MatchString(*dateFlag) {
		date, _ := time.Parse(time.DateOnly, *dateFlag)
		handleOne(date)
	} else if regexDateRange.MatchString(*dateFlag) {
		if *pollSizeFlag <= 0 {
			panic("Invalid --poll-size")
		}

		datesString := strings.Split(*dateFlag, "..")
		start, _ := time.Parse(time.DateOnly, datesString[0])
		end, _ := time.Parse(time.DateOnly, datesString[1])
		if start.After(end) {
			panic("The start date is after the end date")
		}

		dates := append(make([]time.Time, 0), start)

		for {
			current := dates[len(dates)-1]
			if current.Equal(end) || current.After(end) {
				break
			}
			dates = append(dates, current.AddDate(0, 0, 1))
		}

		handleMany(dates)
	} else {
		panic("Invalid date format")
	}
}

func makeUrl(date time.Time) string {
	return fmt.Sprintf(
		"https://data.chc.ucsb.edu/products/CHIRPS-2.0/global_daily/tifs/%s/%04d/chirps-v2.0.%04d.%02d.%02d.tif.gz",
		*precisionFlag,
		date.Year(),
		date.Year(),
		date.Month(),
		date.Day(),
	)
}

func downloadAndUnzipIfNeeded(url string, attempt int) (io.Reader, error) {
	if attempt >= *attemptsFlag {
		return nil, fmt.Errorf("too many attempts")
	}

	req, reqErr := http.NewRequest("GET", url, nil)
	if reqErr != nil {
		return downloadAndUnzipIfNeeded(url, attempt+1)
	}

	resp, respErr := http.DefaultClient.Do(req)
	if respErr != nil {
		return downloadAndUnzipIfNeeded(url, attempt+1)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("response status is not 2xx: %d", resp.StatusCode)
	}

	if *noGunzipFlag {
		return resp.Body, nil
	} else {
		reader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
		return &closingReader{reader, resp.Body}, nil
	}
}

func handleOne(date time.Time) {
	stream, err := downloadAndUnzipIfNeeded(makeUrl(date), 0)
	if err != nil {
		if err.Error() == "response status is not 2xx: 404" {
			_, err := fmt.Fprintln(os.Stderr, "No data for", date.Format(time.DateOnly))
			if err != nil {
				panic(err)
			}
			return
		}

		panic(err)
	}

	if *saveFlag != "" {
		fileName := fmt.Sprintf("%s.tif", date.Format(time.DateOnly))
		if *noGunzipFlag {
			fileName += ".gz"
		}

		err := os.MkdirAll(*saveFlag, 0755)
		if err != nil {
			panic(err)
		}
		file, err := os.Create(filepath.Join(*saveFlag, fileName))
		if err != nil {
			panic(err)
		}

		_, err = io.Copy(file, stream)
		if err != nil {
			panic(err)
		}

		err = file.Close()
		if err != nil {
			panic(err)
		}
	} else {
		_, err := io.Copy(os.Stdout, stream)
		if err != nil {
			panic(err)
		}
	}
}

func handleMany(dates []time.Time) {
	startedAt := time.Now()

	var wg sync.WaitGroup
	workPool := make(chan struct{}, *pollSizeFlag)
	done := 0

	for _, date := range dates {
		workPool <- struct{}{}
		wg.Add(1)

		date := date
		go func() {
			defer func() {
				<-workPool
				wg.Done()
			}()
			handleOne(date)
			done++

			if !*silentFlag {
				fmt.Println(
					fmt.Sprintf(
						"%d of %d files downloaded (%.2f%%). ETA of roughly %d more minutes",
						done,
						len(dates),
						float64(done)/float64(len(dates))*100,
						int(time.Since(startedAt).Seconds()/float64(done)*float64(len(dates)-done)/60),
					),
				)
			}

		}()
	}

	wg.Wait()
}
