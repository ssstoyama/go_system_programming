.PHONY: serve
serve:
	go run tcp/server/main.go

.PHONY: req
req:
	go run tcp/client/main.go
