if [ "$1" == "-c" ]
then
  go run ./cmd/client/core
elif [ "$1" == "-s" ]
  then
    go run ./cmd/server
elif [ "$1" == "-cg" ]
  then
    go run ./cmd/client/gui
fi