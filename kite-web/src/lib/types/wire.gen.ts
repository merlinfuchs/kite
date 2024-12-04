// Code generated by tygo. DO NOT EDIT.
import { FlowData } from './flow.gen';
import { MessageData } from './message.gen';
interface Empty {}

//////////
// source: app.go

export interface App {
  id: string;
  name: string;
  description: null | string;
  enabled: boolean;
  disabled_reason: null | string;
  discord_id: string;
  discord_status?: AppDiscordStatus;
  owner_user_id: string;
  creator_user_id: string;
  created_at: string /* RFC3339 */;
  updated_at: string /* RFC3339 */;
}
export interface AppDiscordStatus {
  status?: string;
  activity_type?: number /* int */;
  activity_name?: string;
  activity_state?: string;
  activity_url?: string;
}
export type AppGetResponse = App;
export interface AppCreateRequest {
  discord_token: string;
}
export type AppCreateResponse = App;
export interface AppUpdateRequest {
  name: string;
  description: null | string;
  discord_status?: AppDiscordStatus;
  enabled: boolean;
}
export type AppUpdateResponse = App;
export interface AppTokenUpdateRequest {
  discord_token: string;
}
export type AppTokenUpdateResponse = App;
export type AppDeleteResponse = Empty;
export type AppListResponse = (App | undefined)[];

//////////
// source: app_state.go

export interface AppStateStatus {
  online: boolean;
}
export type StateStatusGetResponse = AppStateStatus;
export interface Guild {
  id: string;
  name: string;
  description: string;
}
export type StateGuildListResponse = (Guild | undefined)[];
export interface Channel {
  id: string;
  type: number /* int */;
  name: string;
  topic: string;
}
export type StateGuildChannelListResponse = (Channel | undefined)[];

//////////
// source: asset.go

export interface Asset {
  id: string;
  app_id: string;
  module_id: null | string;
  creator_user_id: string;
  url: string;
  name: string;
  content_type: string;
  content_hash: string;
  content_size: number /* int */;
  created_at: string /* RFC3339 */;
  updated_at: string /* RFC3339 */;
  expires_at: null | string /* RFC3339 */;
}
export type AssetCreateResponse = Asset;
export type AssetGetResponse = Asset;

//////////
// source: auth.go

export type AuthLogoutResponse = Empty;

//////////
// source: command.go

export interface Command {
  id: string;
  name: string;
  description: string;
  enabled: boolean;
  app_id: string;
  module_id: null | string;
  creator_user_id: string;
  flow_source: FlowData;
  created_at: string /* RFC3339 */;
  updated_at: string /* RFC3339 */;
  last_deployed_at: null | string /* RFC3339 */;
}
export type CommandGetResponse = Command;
export type CommandListResponse = (Command | undefined)[];
export interface CommandCreateRequest {
  flow_source: FlowData;
  enabled: boolean;
}
export type CommandCreateResponse = Command;
export interface CommandUpdateRequest {
  flow_source: FlowData;
  enabled: boolean;
}
export type CommandUpdateResponse = Command;
export type CommandDeleteResponse = Empty;

//////////
// source: log.go

export interface LogEntry {
  id: number /* int64 */;
  message: string;
  level: string;
  created_at: string /* RFC3339 */;
}
export type LogEntryListResponse = (LogEntry | undefined)[];

//////////
// source: message.go

export interface Message {
  id: string;
  name: string;
  description: null | string;
  app_id: string;
  module_id: null | string;
  creator_user_id: string;
  data: MessageData;
  flow_sources: { [key: string]: FlowData};
  created_at: string /* RFC3339 */;
  updated_at: string /* RFC3339 */;
}
export type MessageGetResponse = Message;
export type MessageListResponse = (Message | undefined)[];
export interface MessageCreateRequest {
  name: string;
  description: null | string;
  data: MessageData;
  flow_sources: { [key: string]: FlowData};
}
export type MessageCreateResponse = Message;
export interface MessageUpdateRequest {
  name: string;
  description: null | string;
  data: MessageData;
  flow_sources: { [key: string]: FlowData};
}
export type MessageUpdateResponse = Message;
export type MessageDeleteResponse = Empty;
export interface MessageInstance {
  id: number /* uint64 */;
  message_id: string;
  discord_guild_id: string;
  discord_channel_id: string;
  discord_message_id: string;
  flow_sources: { [key: string]: FlowData};
  created_at: string /* RFC3339 */;
  updated_at: string /* RFC3339 */;
}
export type MessageInstanceListResponse = (MessageInstance | undefined)[];
export interface MessageInstanceCreateRequest {
  discord_guild_id: string;
  discord_channel_id: string;
}
export type MessageInstanceCreateResponse = MessageInstance;
export interface MessageInstanceUpdateRequest {
}
export type MessageInstanceUpdateResponse = MessageInstance;
export type MessageInstanceDeleteResponse = Empty;

//////////
// source: user.go

export interface User {
  id: string;
  email: string;
  display_name: string;
  discord_id: string;
  discord_username: string;
  discord_avatar: null | string;
  created_at: string /* RFC3339 */;
  updated_at: string /* RFC3339 */;
}
export type UserGetResponse = User;

//////////
// source: variable.go

export interface Variable {
  id: string;
  name: string;
  scoped: boolean;
  app_id: string;
  module_id: null | string;
  total_values: null | number;
  created_at: string /* RFC3339 */;
  updated_at: string /* RFC3339 */;
}
export type VariableGetResponse = Variable;
export type VariableListResponse = (Variable | undefined)[];
export interface VariableCreateRequest {
  name: string;
  scoped: boolean;
}
export type VariableCreateResponse = Variable;
export interface VariableUpdateRequest {
  name: string;
  scoped: boolean;
}
export type VariableUpdateResponse = Variable;
export type VariableDeleteResponse = Empty;
