if [ "$1" == "-c" ]
then
  go run ./cmd/client
elif [ "$1" == "-s" ]
  then
    go run ./cmd/server
fi