.PHONY: all wasm lvlgen

all: wasm lvlgen

wasm:
	cd 'cmd/softgfx-wasm/' && GOOS=js GOARCH=wasm go build -o '../../web/main.wasm' './'

lvlgen:
	cd 'cmd/softgfx-lvlgen/' && go install
