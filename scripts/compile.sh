if [ "$1" == "-c" ]
then
  go build ./cmd/client
elif [ "$1" == "-s" ]
then
  go build ./cmd/server
elif [ "$1" == "-a" ] || [ "$1" == "" ]
then
  go build ./cmd/server
  go build ./cmd/client
fi