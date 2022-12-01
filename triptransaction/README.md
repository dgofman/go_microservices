### Update dependencies
go mod download

### Clear cache
go clean -cache -modcache -i -r

### Run
go run main.go
or
go run main.go --targetEnv production

### The GOTRACEBACK variable controls the amount of output generated
GOTRACEBACK=system
(is like "all" but adds stack frames for run-time functions and shows goroutines created internally by the run-time.)


### Run tests
go test -v ./packages/...

### Run Code cover
go test ./packages/... -cover

go test ./packages/ -coverprofile=coverage.out && go tool cover -func=coverage.out
