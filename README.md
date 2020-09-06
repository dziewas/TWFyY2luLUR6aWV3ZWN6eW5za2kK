# Web Crawler

## Run

To tidy up, test, build and start service:
```shell script
./build.sh
./bin/crawler
```

Or to only start service:
```shell script
go run cmd/service/main.go
```

## Notes
1) I used in-memory storage, but architecture is ready for proper DB (i.e. redis).
2) It would be good to refactor internals to use []byte for responses.
3) Some tests were added, but it would be desirable to add some integration tests because worker is not covered by tests yet.
4) Having HTTP handlers and workers in the same file is not very readable. It could be refactored.

