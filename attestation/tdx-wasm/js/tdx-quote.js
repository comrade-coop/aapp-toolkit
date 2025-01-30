class TdxQuote {
    constructor() {
        this.wasm = null;
    }

    async initialize() {
        if (!this.wasm) {
            this.wasm = await import('../pkg/tdx_wasm.js');
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
