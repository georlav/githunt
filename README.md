![Tests](https://github.com/georlav/githunt/workflows/Tests/badge.svg)
![Linter](https://github.com/georlav/githunt/workflows/Golang-CI/badge.svg)

# GitHunt
A simple command line tool to mass check targets for exposed .git directories.

## Features
 * Check single target for exposed git directory
 * Check multiple targets for exposed git directory
## TODO
 * Add option to dump exposed git directories

## Usage
```text
  _   o  _|_  |_        ._   _|_ 
 (_|  |   |_  | |  |_|  | |   |_   
  _|
Usage: githunt [options...] 

Usage Examples:
  githunt -url example.com
  githunt -urls urls.txt -workers 100 -timeout 30s -output out.txt

Options:
  Target:
    -url         check single url
    -urls        file containing multiple urls (one per line)

  Request:
    -workers     sets the desirable number of http workers (default: 50)
    -cpus        sets the maximum number of CPUs that can be utilized (default: available-1
    -timeout     sets a time limit for requests, valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h". (default: 15s)
  
  General:
    -output      save vulnerable targets to a file
```

## Installation
To install the command line program, use the following:
```bash
go get -v github.com/georlav/githunt/...
```

## Build
To build a local version, use the following:
```bash
go build -o githunt main.go
```

## Credits
 * [georlav](https://github.com/georlav)

## License
The MIT License (MIT) - see [`LICENSE.md`](https://github.com/georlav/githunt/blob/master/LICENSE.md) for more details
