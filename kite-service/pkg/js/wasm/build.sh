#!/bin/sh
# Builds ../kite.wasm from QuickJS-NG and kite.c with wasi-sdk
# (https://github.com/WebAssembly/wasi-sdk), installed in WASI_SDK.
set -eu
cd "$(dirname "$0")"

QUICKJS_VERSION=v0.18.0
WASI_SDK=${WASI_SDK:-/opt/wasi-sdk}

src=$(mktemp -d)
trap 'rm -rf "$src"' EXIT
git clone --quiet --depth 1 --branch "$QUICKJS_VERSION" https://github.com/quickjs-ng/quickjs.git "$src/quickjs"

"$WASI_SDK/bin/clang" --target=wasm32-wasi -O2 \
  -D_GNU_SOURCE -D_WASI_EMULATED_PROCESS_CLOCKS -D_WASI_EMULATED_SIGNAL \
  -I"$src/quickjs" \
  "$src/quickjs/quickjs.c" "$src/quickjs/libregexp.c" "$src/quickjs/libunicode.c" "$src/quickjs/dtoa.c" \
  kite.c \
  -mexec-model=reactor \
  -Wl,-z,stack-size=1048576 \
  -lwasi-emulated-process-clocks -lwasi-emulated-signal \
  -o ../kite.wasm
