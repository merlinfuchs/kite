// Code generated by tygo. DO NOT EDIT.
import {APIResponse} from "./base"

//////////
// source: auth.go

export interface AuthLoginStartRequest {
}
export interface AuthCLIStartResponseData {
  code: string;
}
export type AuthCLIStartResponse = APIResponse<AuthCLIStartResponseData>;
export interface AuthCLICallbackResponseData {
  message: string;
}
export type AuthCLICallbackResponse = APIResponse<AuthCLICallbackResponseData>;
export interface AuthCLICheckResponseData {
  pending: boolean;
  token?: string;
}
export type AuthCLICheckResponse = APIResponse<AuthCLICheckResponseData>;

//////////
// source: compile.go

export interface CompileJSRequest {
  source: string;
}
export interface CompileJSResponseData {
  wasm_bytes: string;
}
export type CompileJSResponse = APIResponse<CompileJSResponseData>;

//////////
// source: deployment.go

export interface Deployment {
  id: string;
  name: string;
  key: string;
  description: string;
  guild_id: string;
  plugin_version_id: null | string;
  wasm_size: number /* int */;
  manifest_default_config: { [key: string]: string};
  manifest_events: string[];
  manifest_commands: string[];
  config: { [key: string]: string};
  created_at: string /* RFC3339 */;
  updated_at: string /* RFC3339 */;
}
export type DeploymentListResponse = APIResponse<Deployment[]>;
export type DeploymentGetResponse = APIResponse<Deployment>;
export interface DeploymentCreateRequest {
  key: string;
  name: string;
  description: string;
  plugin_version_id: null | string;
  wasm_bytes: string;
  manifest_default_config: { [key: string]: string};
  manifest_events: string[];
  manifest_commands: string[];
  config: { [key: string]: string};
}
export type DeploymentCreateResponse = APIResponse<Deployment>;
export type DeploymentDeleteResponse = APIResponse<{
  }>;

//////////
// source: deployment_log.go

export interface DeploymentLogEntry {
  id: number /* uint64 */;
  deployment_id: string;
  level: string;
  message: string;
  created_at: string /* RFC3339 */;
}
export type DeploymentLogEntryListResponse = APIResponse<DeploymentLogEntry[]>;
export interface DeploymentLogSummary {
  deployment_id: string;
  total_count: number /* int */;
  error_count: number /* int */;
  warn_count: number /* int */;
  info_count: number /* int */;
  debug_count: number /* int */;
}
export type DeploymentLogSummaryGetResponse = APIResponse<DeploymentLogSummary>;

//////////
// source: deployment_metric.go

export interface DeploymentMetricEntry {
  id: number /* uint64 */;
  deployment_id: string;
  type: string;
  metadata: { [key: string]: string};
  event_id: number /* uint64 */;
  event_type: string;
  event_succes: boolean;
  event_execution_time: any /* time.Duration */;
  event_total_time: any /* time.Duration */;
  call_type: string;
  call_succes: boolean;
  call_total_time: any /* time.Duration */;
  timestamp: string /* RFC3339 */;
}
export interface DeploymentEventMetricEntry {
  timestamp: string /* RFC3339 */;
  total_count: number /* int */;
  success_count: number /* int */;
  average_execution_time: number /* int64 */;
  average_total_time: number /* int64 */;
}
export type DeploymentMetricEventsListResponse = APIResponse<DeploymentEventMetricEntry[]>;
export interface DeploymentCallMetricEntry {
  timestamp: string /* RFC3339 */;
  total_count: number /* int */;
  success_count: number /* int */;
  average_total_time: number /* int64 */;
}
export type DeploymentMetricCallsListResponse = APIResponse<DeploymentCallMetricEntry[]>;

//////////
// source: guild.go

export interface Guild {
  id: string;
  name: string;
  icon: null | string;
  description: null | string;
  created_at: string /* RFC3339 */;
  updated_at: string /* RFC3339 */;
  user_is_owner?: boolean;
  user_permissions?: string;
  bot_permissions?: string;
}
export type GuildListResponse = APIResponse<Guild[]>;
export type GuildGetResponse = APIResponse<Guild>;

//////////
// source: kv_storage.go

export interface KVStorageNamespace {
  namespace: string;
  key_count: number /* int */;
}
export interface KVStorageValue {
  namespace: string;
  key: string;
  value: any /* kvmodel.TypedKVValue */;
  created_at: string /* RFC3339 */;
  updated_at: string /* RFC3339 */;
}
export type KVStorageNamespaceListResponse = APIResponse<KVStorageNamespace[]>;
export type KVStorageNamespaceKeyListResponse = APIResponse<KVStorageValue[]>;

//////////
// source: quick_access.go

export interface QuickAccessItem {
  id: string;
  guild_id: string;
  type: string;
  name: string;
  updated_at: string /* RFC3339 */;
}
export type QuickAccessItemListResponse = APIResponse<QuickAccessItem[]>;

//////////
// source: user.go

export interface User {
  id: string;
  username: string;
  discriminator: string;
  global_name: string;
  avatar: string;
  public_flags: number /* int */;
  created_at: string /* RFC3339 */;
  updated_at: string /* RFC3339 */;
}
export type UserGetResponse = APIResponse<User>;

//////////
// source: workspace.go

export interface Workspace {
  id: string;
  guild_id: string;
  name: string;
  description: string;
  files: WorkspaceFile[];
  created_at: string /* RFC3339 */;
  updated_at: string /* RFC3339 */;
}
export interface WorkspaceFile {
  path: string;
  content: string;
}
export type WorkspaceGetResponse = APIResponse<Workspace>;
export type WorkspaceListResponse = APIResponse<Workspace[]>;
export interface WorkspaceCreateRequest {
  name: string;
  description: string;
  files: WorkspaceFile[];
}
export type WorkspaceCreateResponse = APIResponse<Workspace>;
export interface WorkspaceUpdateRequest {
  name: string;
  description: string;
  files: WorkspaceFile[];
}
export type WorkspaceUpdateResponse = APIResponse<Workspace>;
export type WorkspaceDeleteResponse = APIResponse<{
  }>;
