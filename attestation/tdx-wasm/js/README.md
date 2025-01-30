# TDX Quote Decoder

A JavaScript library for decoding Intel TDX attestation quotes.

## Usage

```html
<script type="module">
import TdxQuote from './tdx-quote.js';

async function decodeQuote() {
    const decoder = new TdxQuote();
    await decoder.initialize();
    
    const hexQuote = "..."; // Your hex-encoded quote string
    const decodedQuote = decoder.decodeQuote(hexQuote);
    console.log(decodedQuote);
}
</script>
```

## Return Value Structure

The `decodeQuote()` method returns a JSON object with the following structure:

```json
{
    "quote_header": {
        "version": 4,
        "attestation_key_type": "ECDSA-P384",
        "tee_type": "TDX",
        "qe_svn": "...",
        "pce_svn": "...",
        "vendor_id": "...",
        "user_data": "..."
    },
    "report_body": {
        "mr_config_id": "...",
        "mr_owner": "...",
        "mr_owner_config": "...",
        "mr_td": "...",
        "td_attributes": "...",
        "xfam": "...",
        "mr_servicetd": "...",
        "report_data": "..."
    },
    "signature": "...",
    "dns_report_data": {
        "ipv4_address": "192.168.1.100",
        "dns_challenge": "example-challenge",
        "public_key_fingerprint": "abcdef..."
    }
}
```

### Fields Description

- `quote_header`: Contains metadata about the quote format and type
- `report_body`: Contains the core attestation measurements and data
- `signature`: The cryptographic signature of the quote
- `dns_report_data`: Additional DNS-related data embedded in the quote
  - `ipv4_address`: The IPv4 address associated with the attestation
  - `dns_challenge`: A challenge string for DNS verification
  - `public_key_fingerprint`: Fingerprint of the associated public key

## Error Handling

The decoder will throw errors if:
- WASM module is not initialized (call initialize() first)
- Invalid hex string is provided
- Quote data is malformed or cannot be parsed
