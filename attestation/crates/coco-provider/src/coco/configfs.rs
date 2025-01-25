use crate::coco::{Device, ReportRequest, ReportResponse, CONFIGFS_BASE_PATH};
use crate::error::{CocoError, Result};
use crate::utils::{generate_random_data, generate_random_number};
use log::error;
use std::fmt;
use std::fs;
use std::fs::{create_dir_all, remove_dir};

#[derive(Debug)]
pub enum TsmReportAttribute {
    /// Read only. Only for AMD. Holds certs.
    AuxBlob,
    /// Write only
    InBlob,
    /// Read only
    OutBlob,
    /// Read only
    Generation,
    /// Write only. Only for AMD.
    PrivLevel,
    /// Read only
    PrivLevelFloor,
    /// Read only
    Provider,
}

impl fmt::Display for TsmReportAttribute {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        let output = format!("{:?}", self);
        write!(f, "{}", output.to_lowercase())
    }
}

/// ConfigFs does two things.
/// 1. It wraps over a ConfigFsClient which does the real work.
/// 2. Provides additional checks to ensure the result we get is not tampered with.
#[derive(Debug)]
pub struct ConfigFs {
    client: ConfigFsClient,
    /// Expected generation number - track this for every write to watch out for tampering.
    expected_generation: u32,
}

#[derive(Debug)]
pub struct ConfigFsClient {
    /// Path to the folder where we can retrieve the Quote/Signed Attestation Report.
    /// And also other things that the provider may want to provide
    device_path: String,
}

impl ConfigFs {
    pub fn new() -> Result<Self> {
        let rand_num = generate_random_number();
        let device_path = format!("{}/report-{}", CONFIGFS_BASE_PATH, rand_num);
        let client = ConfigFsClient::new(device_path)?;
        let expected_generation = client.read_attribute_u32(&TsmReportAttribute::Generation)?;

        Ok(ConfigFs {
            client,
            expected_generation,
        })
    }

    /// Returns the provider of the report.
    /// Examples: tdx_guest, sev_guest, etc...
    pub fn get_provider(&self) -> Result<String> {
        Ok(self
            .client
            .read_attribute_string(&TsmReportAttribute::Provider)?)
    }

    /// Wrapper over client.write_attribute as we want to keep track of self.expected_generation
    fn write_attribute(&mut self, attribute: &TsmReportAttribute, data: &[u8]) -> Result<()> {
        self.client.write_attribute(attribute, data)?;
        self.expected_generation += 1;
        Ok(())
    }

    fn read_attribute(&self, attribute: &TsmReportAttribute) -> Result<Vec<u8>> {
        let output = self.client.read_attribute(attribute)?;
        self.check_tampering()?;
        Ok(output)
    }

    fn check_tampering(&self) -> Result<()> {
        let gen = self
            .client
            .read_attribute_u32(&TsmReportAttribute::Generation)?;
        if self.expected_generation != gen {
            return Err(CocoError::Firmware(
                "Generation number mismatch".to_string(),
            ));
        }
        Ok(())
    }
}

impl Device for ConfigFs {
    fn get_report(&mut self, req: &ReportRequest) -> Result<ReportResponse> {
        let mut certs = None;
        // Wite the report_data/nonce into inblob
        let report_data = req.report_data.unwrap_or_else(generate_random_data);
        self.write_attribute(&TsmReportAttribute::InBlob, &report_data)?;

        if req.vmpl.is_some_and(|x| x <= 3) {
            // Write the privilege level into the privilege attribute.
            let priv_str = req.vmpl.unwrap().to_string();
            let priv_bytes = priv_str.as_bytes();
            self.write_attribute(&TsmReportAttribute::PrivLevel, priv_bytes)?;
        }

        // Get certs from auxblob if requested
        if req.get_certs.is_some_and(|x| x == true) {
            certs = Some(self.read_attribute(&TsmReportAttribute::AuxBlob)?);
        }
        // Now read the outblob. This is the report.
        let report = self.read_attribute(&TsmReportAttribute::OutBlob)?;

        Ok(ReportResponse { certs, report })
    }
}

impl ConfigFsClient {
    pub fn new(device_path: String) -> Result<Self> {
        create_dir_all(&device_path)?;
        Ok(ConfigFsClient { device_path })
    }

    pub fn write_attribute(&mut self, attribute: &TsmReportAttribute, data: &[u8]) -> Result<()> {
        let path = format!("{}/{}", &self.device_path, attribute);
        fs::write(path, data)?;
        Ok(())
    }

    pub fn read_attribute(&self, attribute: &TsmReportAttribute) -> Result<Vec<u8>> {
        let path = format!("{}/{}", &self.device_path, attribute);
        Ok(fs::read(&path)?)
    }

    pub fn read_attribute_string(&self, attribute: &TsmReportAttribute) -> Result<String> {
        let res = self.read_attribute(attribute)?;
        Ok(String::from_utf8(res)?)
    }

    pub fn read_attribute_u32(&self, attribute: &TsmReportAttribute) -> Result<u32> {
        let res = self.read_attribute_string(attribute)?;
        let res = res.replace("\n", "");
        let num: u32 = res.parse()?;
        Ok(num)
    }
}

impl Drop for ConfigFsClient {
    fn drop(&mut self) {
        // Clean up the report folder when it goes out of scope, as it's meant to be temporary.
        match remove_dir(&self.device_path) {
            Ok(_) => (),
            Err(e) => error!("Error cleaning up {}: {:?}", &self.device_path, e),
        }
    }
}
