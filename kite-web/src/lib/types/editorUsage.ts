export interface EditorUsageItem {
  target_type: "command" | "event_listener";
  target_id: string;
  active_seconds: number;
  save_count: number;
  session_count: number;
  last_edited_at: string | null;
}

export interface EditorUsageGetResponse {
  total: {
    active_seconds: number;
    save_count: number;
    session_count: number;
    last_edited_at: string | null;
  };
  items: EditorUsageItem[];
}
