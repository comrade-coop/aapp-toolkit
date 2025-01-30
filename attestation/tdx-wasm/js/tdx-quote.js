class TdxQuote {
    constructor() {
        this.wasm = null;
    }

    async initialize() {
        if (!this.wasm) {
            const wasmUrl = new URL('../pkg/tdx_wasm_bg.wasm', import.meta.url);
            const wasmModule = await import('../pkg/tdx_wasm.js');
            await wasmModule.default(wasmUrl);
            this.wasm = wasmModule;
        }
    }

    decodeQuote(hexQuote) {
        if (!this.wasm) {
            throw new Error('WASM module not initialized. Call initialize() first');
        }
        return this.wasm.decode_quote_v4(hexQuote);
    }
}

export default TdxQuote;
