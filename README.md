# CHIRPS 2.0 Database Fetcher

Little script for fetching data from the CHIRPS 2.0 database.

As of now only [global/daily](https://data.chc.ucsb.edu/products/CHIRPS-2.0/global_daily/) tiffs are supported. Plans are to support other datasets and formats.

Contributions welcome.

## Installation
For now, ``go build`` it by yourself.

## Usage

```
$ chirpsfetch --help
Usage of chirpsfetch:
  -attemptsFlag int
        The number of attemptsFlag to be made to fetch the data (default 3)
  -date string
        The date or date range to be fetched
  -no-gunzip
        Do not gunzip the downloaded files
  -poll-size int
        The number of records to be fetched and insert at once. Must be greater than 0 (default 128)
  -precision string
        The precision of the data. Can be either 'p05' or 'p25' (default "p05")
  -save string
        Save the downloaded files. Takes a relative path. If not specified, prints to stdout
  -silent
        Do not print the output of the command. Only works if not using --save and --date with range simultaneously, otherwise the program is silent by default

```

## LICENSE

```
Copyright 2023 Nathan P. Bombana

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
```

Do whatever you want with my code just don't make it boring
