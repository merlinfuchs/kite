// Code generated by tygo. DO NOT EDIT.
type Nullable<T> = T | null

//////////
// source: host.go

export type HostErrorType = number /* int */;
export const HostErrorTypeUnknown: HostErrorType = 0;
export const HostErrorTypeTimeout: HostErrorType = 1;
export const HostErrorTypeCanceled: HostErrorType = 2;
export const HostErrorTypeUnimplemented: HostErrorType = 3;
export const HostErrorTypeValidationFailed: HostErrorType = 4;
export const HostErrorTypeDiscordUnknown: HostErrorType = 100;
export const HostErrorTypeDiscordGuildNotFound: HostErrorType = 101;
export const HostErrorTypeDiscordChannelNotFound: HostErrorType = 102;
export const HostErrorTypeDiscordMessageNotFound: HostErrorType = 103;
export const HostErrorTypeDiscordBanNotFound: HostErrorType = 104;
export const HostErrorTypeKVUnknown: HostErrorType = 200;
export const HostErrorTypeKVKeyNotFound: HostErrorType = 201;
export const HostErrorTypeKVValueTypeMismatch: HostErrorType = 202;
export interface HostError {
  code: HostErrorType;
  message: string;
}

//////////
// source: plugin.go

export type ModuleErrorCode = number /* int */;
export const ModuleErrorCodeUnknown: ModuleErrorCode = 0;
export interface ModuleError {
  code: ModuleErrorCode;
  message: string;
}
