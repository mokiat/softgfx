.PHONY: all wasm lvlgen

all: wasm lvlgen

wasm:
	cp "$$(go env GOROOT)/misc/wasm/wasm_exec.js" './web/'
	cd 'cmd/softgfx-wasm/' && GOOS=js GOARCH=wasm go build -o '../../web/main.wasm' './'

lvlgen:
	cd 'cmd/softgfx-lvlgen/' && go install
