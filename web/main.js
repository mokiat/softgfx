window.Plotter = class Plotter {
	constructor(canvasID) {
		const canvas = document.getElementById(canvasID)
		this.width = canvas.width;
		this.height = canvas.height;
		this.context = canvas.getContext("2d");
		this.imageData = this.context.createImageData(canvas.width, canvas.height);
		this.pixels = new Uint8Array(this.imageData.data.buffer);
	}

	flush() {
		this.context.putImageData(this.imageData, 0, 0);
	}
};

window.onload = () => {
	console.log('loading webassembly executable...');
	const go = new Go();
	WebAssembly.instantiateStreaming(fetch("web/main.wasm"), go.importObject).then((result) => {
		console.log('running webassembly executable...');
		go.run(result.instance);
		document.getElementById("loading").remove();
		document.getElementById("screen").style.display = "block";
	});
};