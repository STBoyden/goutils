fmt:
	find -name *.go -exec go run mvdan.cc/gofumpt@latest -w -extra {} \;
