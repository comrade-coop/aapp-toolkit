pub mod device;
pub mod error;
mod utils;
use dcap_rs::types::quotes::version_4::QuoteV4;
use error::Result;

pub struct Tdx;

impl Tdx {
    pub fn new() -> Self {
        Tdx
    }

    pub fn get_attestation_report(&self) -> Result<QuoteV4> {
        let mut device = device::Device::default()?;
        device.get_attestation_report()
    }

    pub fn get_attestation_report_with_options(
        &self,
        options: device::DeviceOptions,
    ) -> Result<QuoteV4> {
        let mut device = device::Device::new(options)?;
        device.get_attestation_report()
    }

    pub fn get_attestation_report_raw(&self) -> Result<Vec<u8>> {
        let mut device = device::Device::default()?;
        device.get_attestation_report_raw()
    }

    pub fn get_attestation_report_raw_with_options(
        &self,
        options: device::DeviceOptions,
    ) -> Result<Vec<u8>> {
        let mut device = device::Device::new(options)?;
        device.get_attestation_report_raw()
    }
}

#[cfg(feature = "clib")]
pub mod c {
    use once_cell::sync::Lazy;
    use std::ptr::copy_nonoverlapping;
    use std::sync::Mutex;

    use super::device::DeviceOptions;
    use super::Tdx;

    static ATTESTATION_REPORT: Lazy<Mutex<Vec<u8>>> = Lazy::new(|| Mutex::new(Vec::new()));

    /// Use this function to generate the attestation report with default settings.
    /// Returns the size of the report, which you can use to malloc a buffer of suitable size
    /// before you call get_attestation_report_raw().
    #[no_mangle]
    pub extern "C" fn generate_attestation_report() -> usize {
        let tdx = Tdx::new();

        let bytes = tdx.get_attestation_report_raw().unwrap();
        let len = bytes.len();
        match ATTESTATION_REPORT.lock() {
            Ok(mut t) => {
                *t = bytes;
            }
            Err(e) => {
                panic!("Error: {:?}", e);
            }
        }
        len
    }

    /// Use this function to generate the attestation report with options.
    /// Returns the size of the report, which you can use to malloc a buffer of suitable size
    /// before you call get_attestation_report_raw().
    #[no_mangle]
    pub extern "C" fn generate_attestation_report_with_options(report_data: *mut u8) -> usize {
        let tdx = Tdx::new();
        let mut rust_report_data: [u8; 64] = [0; 64];
        unsafe {
            copy_nonoverlapping(report_data, rust_report_data.as_mut_ptr(), 64);
        }
        let device_options = DeviceOptions {
            report_data: Some(rust_report_data),
        };
        let bytes = tdx
            .get_attestation_report_raw_with_options(device_options)
            .unwrap();
        let len = bytes.len();
        match ATTESTATION_REPORT.lock() {
            Ok(mut t) => {
                *t = bytes;
            }
            Err(e) => {
                panic!("Error: {:?}", e);
            }
        }
        len
    }

    /// Ensure that generate_attestation_report() is called first to get the size of buf.
    /// Use this size to malloc enough space for the attestation report that will be transferred.
    #[no_mangle]
    pub extern "C" fn get_attestation_report_raw(buf: *mut u8) {
        let bytes = match ATTESTATION_REPORT.lock() {
            Ok(t) => t.clone(),
            Err(e) => {
                panic!("Error: {:?}", e);
            }
        };
        if bytes.len() == 0 {
            panic!("Error: No attestation report found! Please call generate_attestation_report() first.");
        }

        unsafe {
            copy_nonoverlapping(bytes.as_ptr(), buf, bytes.len());
        }
    }
}
