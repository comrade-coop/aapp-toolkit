use crate::error::Result;
use crate::utils::generate_random_data;
use coco_provider::{
    coco::{CocoDeviceType, ReportRequest},
    get_coco_provider, CocoProvider,
};
use dcap_rs::types::quotes::version_4::QuoteV4;
use serde::Deserialize;

const IMDS_QUOTE_URL: &str = "http://169.254.169.254/acc/tdquote";

pub struct DeviceOptions {
    /// 64 bytes of data to use for the request
    /// Basically acts as a nonce.
    /// Only applicable when the device is configfs or legacy.
    /// For TPM, the the platform provides the data.
    /// Defaults to randomly generating 64 bytes if `None` provided.
    pub report_data: Option<[u8; 64]>,
}
pub struct Device {
    options: DeviceOptions,
    provider: CocoProvider,
}

#[derive(Clone, Debug, Deserialize)]
struct QuoteResponse {
    quote: String,
}

impl Device {
    pub fn default() -> Result<Self> {
        let options = DeviceOptions { report_data: None };
        let provider = get_coco_provider()?;
        Ok(Device { options, provider })
    }

    pub fn new(options: DeviceOptions) -> Result<Self> {
        let provider = get_coco_provider()?;
        Ok(Device { options, provider })
    }

    pub fn get_attestation_report(&mut self) -> Result<QuoteV4> {
        let raw_report = self.get_attestation_report_raw()?;
        Ok(QuoteV4::from_bytes(&raw_report))
    }

    pub fn get_attestation_report_raw(&mut self) -> Result<Vec<u8>> {
        let report_data = match self.provider.device_type {
            CocoDeviceType::Tpm => {
                if !self.options.report_data.is_none() {
                    return Err("report_data cannot be provided for TPM!".into());
                }
                None
            }
            _ => self.options.report_data.or_else(generate_random_data),
        };
        let req = ReportRequest {
            report_data,
            vmpl: None,
            get_certs: None,
        };
        let response = self.provider.device.get_report(&req)?;
        // If the provider is a TPM, we still need to turn the td_report into a signed quote.
        // This can be done by sending the td_report to the IMDS.
        if self.provider.device_type == CocoDeviceType::Tpm {
            // Get the quote from the IMDS.
            let td_report = response.report;
            let response: QuoteResponse = ureq::post(IMDS_QUOTE_URL)
                .send_json(ureq::json!({
                    "report": base64_url::encode(&td_report),
                }))?
                .into_json()?;
            let quote = base64_url::decode(&response.quote)?;
            return Ok(quote);
        }
        // Otherwise we can just return the quote from the TPM as it is.
        Ok(response.report)
    }
}
