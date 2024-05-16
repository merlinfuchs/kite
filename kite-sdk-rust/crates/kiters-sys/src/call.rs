use serde::{Deserialize, Serialize};

use crate::HostError;

#[derive(Debug, Serialize)]
pub struct Call {
    pub config: CallConfig,
    #[serde(flatten)]
    pub data: CallData,
}

impl Call {
    pub fn kind(&self) -> &'static str {
        self.data.kind()
    }
}

#[derive(Debug, Serialize)]
#[serde(tag = "type", content = "data", rename_all = "snake_case")]
pub enum CallData {
    Sleep { duration: u32 },
}

impl CallData {
    pub fn kind(&self) -> &'static str {
        match self {
            CallData::Sleep { .. } => "SLEEP",
        }
    }
}

#[derive(Debug, Default, Serialize)]
pub struct CallConfig {
    reason: String,
    timeout: u32,
    wait: bool,
}

#[derive(Debug, Deserialize)]
pub struct CallResponse {
    pub success: bool,
    pub error: Option<HostError>,
}
