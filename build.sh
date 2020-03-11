if [ ! -d out ]; then
  mkdir out
fi

GOOS=$1 GOARCH=$2 go build -o out/cron-node-$1-$2 ./main.go