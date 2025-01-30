use wasm_bindgen::prelude::*;
use tdx::QuoteV4;
use serde::Serialize;
use sha2::{Sha256, Digest};
use std::net::Ipv4Addr;

#[derive(Serialize)]
struct DNSReportData {
    ipv4_address: String,
    dns_challenge: String,
    public_key_fingerprint: String,
}

impl DNSReportData {
    fn from_report_data(report_data: &[u8]) -> Self {
        // First 16 bytes: take last 4 for IPv4
        let ipv4_bytes: [u8; 4] = report_data[12..16].try_into().unwrap();
        let ipv4 = Ipv4Addr::from(ipv4_bytes);

        // Next 48 bytes for DNS challenge
        let challenge_input = &report_data[16..64];
        let mut hasher = Sha256::new();
        hasher.update(challenge_input);
        let hash_result = hasher.finalize();
        let dns_challenge = base64::encode_config(&hash_result, base64::URL_SAFE_NO_PAD);

        // Last 32 bytes for public key fingerprint
        let fingerprint = hex::encode(&report_data[32..64]);

        DNSReportData {
            ipv4_address: ipv4.to_string(),
            dns_challenge,
            public_key_fingerprint: fingerprint,
        }
    }
}

#[derive(Serialize)]
struct AttestedDNSQuoteV4 {
    #[serde(flatten)]
    quote: QuoteV4,
    unpacked_report_data: DNSReportData,
}

#[wasm_bindgen]
pub fn decode_quote_v4(hex_quote: &str) -> Result<JsValue, JsError> {
    // Convert hex string to bytes
    let quote_bytes = hex::decode(hex_quote)
        .map_err(|e| JsError::new(&format!("Failed to decode hex: {}", e)))?;
    
    // Parse quote
    let quote = QuoteV4::from_bytes(&quote_bytes);
    
    // Get report data and create unpacked version
    let report_data = quote.report_body.report_data.as_bytes();
    let dns_report_data = DNSReportData::from_report_data(report_data);
    
    // Create attested DNS quote
    let attested_quote = AttestedDNSQuoteV4 {
        quote,
        unpacked_report_data: dns_report_data,
    };
    
    // Convert to JsValue using serde
    let js_value = serde_wasm_bindgen::to_value(&attested_quote)
        .map_err(|e| JsError::new(&format!("Failed to serialize quote: {}", e)))?;
    
    Ok(js_value)
}
