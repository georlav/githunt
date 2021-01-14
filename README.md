# GitHunt
A simple tool to mass check targets for exposed .git directories.

```text

  _   o  _|_  |_        ._   _|_ 
 (_|  |   |_  | |  |_|  | |   |_ 
  _|
Usage: githunt [options...] 

Usage Examples:
  githunt -target example.com
  githunt -targets urls.txt -workers 100 -timeout 5 -output out.txt

Options:
  -url         url of target
  -urls        file containing multiple urls (one url per line)
  -rate-limit  sets requests per second limit (default:500)
  -workers     sets the desirable number of http workers (default: 10)
  -cpus        sets the maximum number of CPUs that can be utilized (default: 7)
  -timeout     set a time limit for requests in seconds (default: 15)
  -output      save vulnerable targets in a file
  -debug       enable debug messages (default: disabled)

```