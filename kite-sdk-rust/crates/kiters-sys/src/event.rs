use serde::{Deserialize, Serialize};

use crate::ModuleError;

#[derive(Debug, Deserialize)]
pub struct Event {
    pub app_id: String,
    pub guild_id: String,
    #[serde(flatten)]
    pub data: EventData,
}

impl Event {
    pub fn kind(&self) -> &'static str {
        self.data.kind()
    }
}

#[derive(Debug, Deserialize)]
#[serde(tag = "type", content = "data", rename_all = "SCREAMING_SNAKE_CASE")]
pub enum EventData {
    Initiate,
}

impl EventData {
    pub fn kind(&self) -> &'static str {
        match self {
            EventData::Initiate => "INITIATE",
        }
    }
}

#[derive(Debug, Serialize)]
pub struct EventResponse {
    pub success: bool,
    pub error: Option<ModuleError>,
}
