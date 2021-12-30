all:
	sass sass/:public/css/ --style compressed
	go build -o nilsu-docs
