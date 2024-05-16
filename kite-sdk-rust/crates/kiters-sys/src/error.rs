use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize)]
pub struct ModuleError {
    pub code: String,
    pub message: String,
}

impl From<HostError> for ModuleError {
    fn from(err: HostError) -> ModuleError {
        ModuleError {
            code: err.code,
            message: err.message,
        }
    }
}

#[derive(Debug, Deserialize)]
pub struct HostError {
    pub code: String,
    pub message: String,
}

impl HostError {
    pub fn new<A: ToString, B: ToString>(code: A, message: B) -> HostError {
        HostError {
            code: code.to_string(),
            message: message.to_string(),
        }
    }
}
