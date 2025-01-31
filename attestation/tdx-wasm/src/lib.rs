use wasm_bindgen::prelude::*;
use dcap_rs::types::quotes::version_4::QuoteV4;
use dcap_rs::types::quotes::body::QuoteBody;
use serde::Serialize;
use sha2::{Sha256, Digest};
use std::net::Ipv4Addr;
use base64::{engine::general_purpose, Engine as _};

#[derive(Serialize)]
struct DNSReportData {
    ipv4_address: String,
    dns_challenge: String,
    public_key_fingerprint: String,
}

impl DNSReportData {
    fn from_report_data(report_data: &[u8; 64]) -> Self {
        // First 16 bytes: take last 4 for IPv4
        let ipv4_bytes: [u8; 4] = report_data[12..16].try_into().unwrap();
        let ipv4 = Ipv4Addr::from(ipv4_bytes);

        // Next 48 bytes for DNS challenge
        let challenge_input = &report_data[16..64];
        let mut hasher = Sha256::new();
        hasher.update(challenge_input);
        let hash_result = hasher.finalize();
        let dns_challenge = general_purpose::URL_SAFE_NO_PAD.encode(&hash_result);

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
    let QuoteBody::TD10QuoteBody(td10_report) = quote.quote_body else { todo!() };
    let report_data = td10_report.report_data;
    let dns_report_data = DNSReportData::from_report_data(&report_data);
    
    // Create attested DNS quote
    let attested_quote = AttestedDNSQuoteV4 {
        unpacked_report_data: dns_report_data,
    };
    
    // Convert to JsValue using serde
    let js_value = serde_wasm_bindgen::to_value(&attested_quote)
        .map_err(|e| JsError::new(&format!("Failed to serialize quote: {}", e)))?;
    
    Ok(js_value)
}
