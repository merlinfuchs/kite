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
  enabled: boolean;
}
export type AppUpdateResponse = App;
export interface AppStatusUpdateRequest {
  discord_status?: AppDiscordStatus;
}
export type AppStatusUpdateResponse = App;
export interface AppTokenUpdateRequest {
  discord_token: string;
}
export type AppTokenUpdateResponse = App;
export type AppDeleteResponse = Empty;
export type AppListResponse = (App | undefined)[];
export type AppEmojiListResponse = (AppEmoji | undefined)[];
export interface AppEmoji {
  id: string;
  name: string;
  animated: boolean;
  available: boolean;
}
export type AppEntityListResponse = (AppEntity | undefined)[];
export interface AppEntity {
  id: string;
  type: string;
  name: string;
}

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
  icon_url: null | string;
  created_at: string /* RFC3339 */;
}
export type StateGuildListResponse = (Guild | undefined)[];
export interface Channel {
  id: string;
  type: number /* int */;
  name: string;
  topic: string;
}
export type StateGuildChannelListResponse = (Channel | undefined)[];
export type StateGuildLeaveResponse = Empty;

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
// source: billing.go

export interface BillingWebhookRequest {
  meta: {
    event_name: string;
    custom_data: { [key: string]: any};
  };
  data: {
    id: string;
    attributes: {
      store_id: number /* int */;
      customer_id: number /* int */;
      order_id: number /* int */;
      order_item_id: number /* int */;
      product_id: number /* int */;
      variant_id: number /* int */;
      product_name: string;
      variant_name: string;
      user_name: string;
      user_email: string;
      status: string;
      status_formatted: string;
      card_brand: string;
      card_last_four: string;
      cancelled: boolean;
      trial_ends_at: null | string /* RFC3339 */;
      billing_anchor: number /* int */;
      renews_at: string /* RFC3339 */;
      ends_at: null | string /* RFC3339 */;
      created_at: string /* RFC3339 */;
      updated_at: string /* RFC3339 */;
      test_mode: boolean;
    };
  };
}
export interface BillingWebhookResponse {
}
export interface BillingCheckoutRequest {
  lemonsqueezy_variant_id: string;
}
export interface BillingCheckoutResponse {
  url: string;
}
export interface SubscriptionManageResponse {
  update_payment_method_url: string;
  customer_portal_url: string;
}
export interface Subscription {
  id: string;
  display_name: string;
  source: string;
  status: string;
  status_formatted: string;
  created_at: string /* RFC3339 */;
  updated_at: string /* RFC3339 */;
  renews_at: string /* RFC3339 */;
  trial_ends_at: null | string /* RFC3339 */;
  ends_at: null | string /* RFC3339 */;
  user_id: string;
  lemonsqueezy_subscription_id: null | string;
  lemonsqueezy_customer_id: null | string;
  lemonsqueezy_order_id: null | string;
  lemonsqueezy_product_id: null | string;
  lemonsqueezy_variant_id: null | string;
  manageable: boolean;
}
export type SubscriptionListResponse = (Subscription | undefined)[];
export interface BillingPlan {
  id: string;
  title: string;
  description: string;
  price: number /* float32 */;
  default: boolean;
  popular: boolean;
  hidden: boolean;
  lemonsqueezy_product_id: string;
  lemonsqueezy_variant_id: string;
  discord_role_id: string;
  feature_max_collaborators: number /* int */;
  feature_usage_credits_per_month: number /* int */;
  feature_max_guilds: number /* int */;
  feature_priority_support: boolean;
}
export type BillingPlanListResponse = (BillingPlan | undefined)[];

//////////
// source: collaborator.go

export interface AppCollaborator {
  user: User;
  role: string;
  created_at: string /* RFC3339 */;
  updated_at: string /* RFC3339 */;
}
export type AppCollaboratorListResponse = (AppCollaborator | undefined)[];
export interface AppCollaboratorCreateRequest {
  discord_user_id: string;
  role: string;
}
export type AppCollaboratorCreateResponse = AppCollaborator;
export type AppCollaboratorDeleteResponse = Empty;

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
export interface CommandsImportRequest {
  commands: CommandCreateRequest[];
}
export type CommandsImportResponse = (Command | undefined)[];
export interface CommandUpdateRequest {
  flow_source: FlowData;
  enabled: boolean;
}
export type CommandUpdateResponse = Command;
export interface CommandUpdateEnabledRequest {
  enabled: boolean;
}
export type CommandUpdateEnabledResponse = Command;
export type CommandDeleteResponse = Empty;

//////////
// source: event_listener.go

export interface EventListener {
  id: string;
  source: string;
  type: string;
  description: string;
  enabled: boolean;
  app_id: string;
  module_id: null | string;
  creator_user_id: string;
  filter?: EventListenerFilter;
  flow_source: FlowData;
  created_at: string /* RFC3339 */;
  updated_at: string /* RFC3339 */;
}
export interface EventListenerFilter {
}
export type EventListenerGetResponse = EventListener;
export type EventListenerListResponse = (EventListener | undefined)[];
export interface EventListenerCreateRequest {
  source: string;
  flow_source: FlowData;
  enabled: boolean;
}
export type EventListenerCreateResponse = EventListener;
export interface EventListenersImportRequest {
  event_listeners: EventListenerCreateRequest[];
}
export type EventListenersImportResponse = (EventListener | undefined)[];
export interface EventListenerUpdateRequest {
  flow_source: FlowData;
  enabled: boolean;
}
export type EventListenerUpdateResponse = EventListener;
export interface EventListenerUpdateEnabledRequest {
  enabled: boolean;
}
export type EventListenerUpdateEnabledResponse = EventListener;
export type EventListenerDeleteResponse = Empty;

//////////
// source: feature.go

export interface Features {
  max_collaborators: number /* int */;
  usage_credits_per_month: number /* int */;
  max_guilds: number /* int */;
  priority_support: boolean;
}
export type FeaturesGetResponse = Features;

//////////
// source: log.go

export interface LogEntry {
  id: number /* int64 */;
  message: string;
  level: string;
  command_id: null | string;
  event_listener_id: null | string;
  message_id: null | string;
  created_at: string /* RFC3339 */;
}
export type LogEntryListResponse = (LogEntry | undefined)[];
export interface LogSummary {
  total_entries: number /* int64 */;
  total_errors: number /* int64 */;
  total_warnings: number /* int64 */;
  total_infos: number /* int64 */;
  total_debugs: number /* int64 */;
}
export type LogSummaryGetResponse = LogSummary;

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
export interface MessagesImportRequest {
  messages: MessageCreateRequest[];
}
export type MessagesImportResponse = (Message | undefined)[];
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
// source: plugin.go

export interface Plugin {
  id: string;
  name: string;
  description: string;
  icon: string;
  author: string;
  version: string;
  config: PluginConfig;
}
export interface PluginConfig {
  sections: PluginConfigSection[];
}
export interface PluginConfigSection {
  name: string;
  description: string;
  fields: PluginConfigField[];
}
export interface PluginConfigField {
  key: string;
  type: string;
  item_type: string;
  name: string;
  description: string;
}
export interface PluginInstance {
  app_id: string;
  plugin_id: string;
  enabled: boolean;
  config: Record<string, any> | null;
  created_at: string /* RFC3339 */;
  updated_at: string /* RFC3339 */;
}
export type PluginListResponse = Plugin[];
export type PluginInstanceGetResponse = PluginInstance;
export interface PluginInstanceUpdateRequest {
  enabled: boolean;
  config: Record<string, any> | null;
}
export type PluginInstanceUpdateResponse = PluginInstance;

//////////
// source: usage.go

export interface UsageCreditsGetResponse {
  credits_used: number /* int */;
}
export type UsageByDayListResponse = (UsageByDayEntry | undefined)[];
export interface UsageByDayEntry {
  date: string /* RFC3339 */;
  credits_used: number /* int */;
}
export type UsageByTypeListResponse = (UsageByTypeEntry | undefined)[];
export interface UsageByTypeEntry {
  type: string;
  credits_used: number /* int */;
}

//////////
// source: user.go

export interface User {
  id: string;
  email: null | string;
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
export interface VariablesImportRequest {
  variables: VariableCreateRequest[];
}
export type VariablesImportResponse = (Variable | undefined)[];
export interface VariableUpdateRequest {
  name: string;
  scoped: boolean;
}
export type VariableUpdateResponse = Variable;
export type VariableDeleteResponse = Empty;
