#[cfg(feature = "configfs")]
pub mod configfs;
pub mod tpm;

use crate::cpu::CpuVendor;
use crate::error::{CocoError, Result};
use crate::utils::generate_random_number;
use std::fmt::Debug;
use std::fs::{create_dir_all, remove_dir};
use std::path::Path;

const CONFIGFS_BASE_PATH: &str = "/sys/kernel/config/tsm/report";
const SEV_LEGACY_PATH: &str = "/dev/sev-guest";
const TPM_PATHS: [&str; 2] = ["/dev/tpm0", "/dev/tpmrm0"];

pub trait Device: Debug {
    fn get_report(&mut self, req: &ReportRequest) -> Result<ReportResponse>;
}

#[derive(Debug, PartialEq)]
pub enum CocoDeviceType {
    /// Coco device is available via configfs
    ConfigFs,
    /// Coco device is available via legacy interface
    /// Only applicable to AMD /dev/sev_guest
    Legacy,
    /// Coco device is exposed via Tpm chip.
    /// For now, the only provider that uses this is Azure.
    Tpm,
}

/// Struct to request information from a CoCo device
pub struct ReportRequest {
    /// Nonce to provide for the report
    pub report_data: Option<[u8; 64]>,
    /// Privilege level to use for the report.
    /// Only applies to SEV-SNP. Must be a value between 0-3.
    pub vmpl: Option<u32>,
    /// Determine whether to retrieve certificates from the provider.
    /// This attribute is only available for AMD.
    /// Always set to None for intel.
    pub get_certs: Option<bool>,
}

/// Struct to encapsulate response from a CoCo device
pub struct ReportResponse {
    /// Only applicable to AMD SEV-SNP. Returns the certs if requested and if there's any.
    pub certs: Option<Vec<u8>>,
    /// Raw attestation report as bytes.
    pub report: Vec<u8>,
}

/// Get the type of CoCo device available on the system.
/// ## Parameters
/// * `vendor` - CPU vendor
///
/// ## Returns
/// * `CocoDeviceType` - Type of CoCo device available
pub fn get_device_type(vendor: &CpuVendor) -> Result<CocoDeviceType> {
    if vendor == &CpuVendor::Arm {
        return Err(CocoError::Firmware(
            "Arm CoCo is not supported yet. Maybe next time.".to_string(),
        ));
    }
    // Prefer configfs if it exists, as it's the unified standard
    // that all providers are moving to.
    if Path::new(CONFIGFS_BASE_PATH).exists() {
        if try_create_configfs_report_folder() {
            return Ok(CocoDeviceType::ConfigFs);
        }
    }
    // Check for legacy device only if it's AMD.
    if vendor == &CpuVendor::Amd && Path::new(SEV_LEGACY_PATH).exists() {
        return Ok(CocoDeviceType::Legacy);
    }
    // Else check if TPM is available.
    for path in TPM_PATHS.iter() {
        if Path::new(path).exists() {
            return Ok(CocoDeviceType::Tpm);
        }
    }
    Err(CocoError::Firmware("No CoCo device found".to_string()))
}

/// Initial test to see if we can create a folder in configfs
/// If we are unable to, it means the device does not exist,
/// or we do not have the right permissions.
fn try_create_configfs_report_folder() -> bool {
    let rand_num = generate_random_number();
    let device_path = format!("{}/report-{}", CONFIGFS_BASE_PATH, rand_num);
    if create_dir_all(&device_path).is_ok() {
        if remove_dir(device_path).is_ok() {
            return true;
        }
    }
    false
}
