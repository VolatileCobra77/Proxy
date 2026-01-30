#!/bin/zsh

GO_BIN=$(command -v go)

if [ -z "$GO_BIN" ]; then
  echo "Go not found in PATH"
  exit 1
fi

BUILD_TARGET="$1"
OS_TARGETS="$2"

# Defaults
if [ -z "$BUILD_TARGET" ] || [ "$BUILD_TARGET" = "-a" ]; then
  BUILD_SERVER=true
  BUILD_CLIENT=true
elif [ "$BUILD_TARGET" = "-s" ]; then
  BUILD_SERVER=true
  BUILD_CLIENT=false
elif [ "$BUILD_TARGET" = "-c" ]; then
  BUILD_SERVER=false
  BUILD_CLIENT=true
else
  echo "Unknown build target: $BUILD_TARGET"
  exit 1
fi

if [ -z "$OS_TARGETS" ] || [ "$OS_TARGETS" = "-a" ]; then
  OS_TARGETS="-wlm"
fi

build() {
  NAME=$1
  BUILD_PATH=$2

  if [[ "$OS_TARGETS" == *"w"* ]]; then
    echo "Building $NAME (Windows)"
    GOOS=windows GOARCH=amd64 "$GO_BIN" build -o build/${NAME}-windows.exe "$BUILD_PATH"
  fi

  if [[ "$OS_TARGETS" == *"l"* ]]; then
    echo "Building $NAME (Linux)"
    GOOS=linux GOARCH=amd64 "$GO_BIN" build -o build/${NAME}-linux "$BUILD_PATH"
  fi

  if [[ "$OS_TARGETS" == *"m"* ]]; then
    echo "Building $NAME (macOS)"
    GOOS=darwin GOARCH=amd64 "$GO_BIN" build -o build/${NAME}-macos "$BUILD_PATH"
  fi
}

mkdir -p build

if [ "$BUILD_SERVER" = true ]; then
  build server ./cmd/server
fi

if [ "$BUILD_CLIENT" = true ]; then
  build client ./cmd/client
fi

echo "Build complete."
