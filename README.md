# ParDow
A library for download single file with multiple goroutines concurrently.

```go
ParallelDownload(url string, goroutineCount int64, path string, saveName string)
```
- url: Url of file which you want to download
- goroutineCount: How many goroutine you want it to work 
- path: Where to file will download
- saveName: Name of the file