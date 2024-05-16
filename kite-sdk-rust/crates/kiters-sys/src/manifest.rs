use serde::Serialize;

#[derive(Debug, Default, Serialize)]
pub struct Manifest {
    pub events: Vec<String>,
    pub discord_commands: Vec<DiscordCommand>,
    pub config_schema: ConfigSchema,
}

#[derive(Debug, Serialize)]
pub struct DiscordCommand {}

#[derive(Debug, Default, Serialize)]
pub struct ConfigSchema {}
