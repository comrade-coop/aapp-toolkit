pub mod coco;
pub mod cpu;
pub mod error;
pub mod hypervisor;
pub mod utils;

use users::get_current_uid;

use coco::{get_device_type, CocoDeviceType, Device};
use cpu::{CpuArchitecture, CpuVendor};
use error::Result;
use hypervisor::Hypervisor;

#[cfg(test)]
#[path = "./tests/lib-test.rs"]
mod lib_tests;

/// Struct to represent all information we need about the
/// Confidential Computing (coco) Provider.
#[derive(Debug)]
pub struct CocoProvider {
    /// VM architecture
    pub arch: CpuArchitecture,
    /// VM hypervisor
    pub hypervisor: Hypervisor,
    /// Cpu Vendor
    pub cpu_vendor: CpuVendor,
    /// CPU Model
    pub cpu_model: String,
    /// Type of hardware device exposed to the VM for coco
    pub device_type: CocoDeviceType,
    /// Handle to the device
    pub device: Box<dyn Device>,
}

/// Retrieve the coco provider information for your system.
pub fn get_coco_provider() -> Result<CocoProvider> {
    if get_current_uid() != 0 {
        return Err(error::CocoError::Permission(
            "Please run this program as root to access CoCo device!".to_string(),
        ));
    }
    let arch = cpu::get_architecture();
    let cpu_vendor = cpu::get_vendor();
    let cpu_model = cpu::get_model();
    let hypervisor = hypervisor::get_hypervisor()?;
    // Perform additional checks before creating device.
    match cpu_vendor {
        CpuVendor::Intel => {
            if hypervisor == Hypervisor::HyperV {
                hypervisor::hyperv_extra_isolation_checks(&hypervisor::HypervIsolationType::Tdx)?;
            } else {
                cpu::check_tdx_enabled()?;
            }
        }
        CpuVendor::Amd => {
            if hypervisor == Hypervisor::HyperV {
                hypervisor::hyperv_extra_isolation_checks(&hypervisor::HypervIsolationType::Snp)?;
            }
        }
        _ => {}
    }
    let device_type = get_device_type(&cpu_vendor)?;
    construct_coco_provider(device_type, arch, hypervisor, cpu_vendor, cpu_model)
}

fn construct_coco_provider(
    device_type: CocoDeviceType,
    arch: CpuArchitecture,
    hypervisor: Hypervisor,
    cpu_vendor: CpuVendor,
    cpu_model: String,
) -> Result<CocoProvider> {
    match device_type {
        CocoDeviceType::ConfigFs => {
            #[cfg(feature = "configfs")]
            {
                let device = coco::configfs::ConfigFs::new()?;
                return Ok(CocoProvider {
                    arch,
                    hypervisor,
                    cpu_vendor,
                    cpu_model,
                    device_type,
                    device: Box::new(device),
                });
            }
            #[cfg(not(feature = "configfs"))]
            {
                return Err(error::CocoError::Firmware(
                    "ConfigFS feature not enabled!!!".to_string(),
                ));
            }
        }
        CocoDeviceType::Tpm => {
            #[cfg(feature = "tpm")]
            {
                let device = coco::tpm::Tpm::new(cpu_vendor, hypervisor)?;
                return Ok(CocoProvider {
                    arch,
                    hypervisor,
                    cpu_vendor,
                    cpu_model,
                    device_type,
                    device: Box::new(device),
                });
            }
            #[cfg(not(feature = "tpm"))]
            {
                return Err(error::CocoError::Firmware(
                    "TPM feature not enabled!!!".to_string(),
                ));
            }
        }
        _ => {
            return Err(error::CocoError::Firmware(
                "Device type not supported".to_string(),
            ));
        }
    }
}
