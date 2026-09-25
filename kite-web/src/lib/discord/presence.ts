export const statusOptions = [
  { value: "online", label: "Online" },
  { value: "dnd", label: "Do Not Disturb" },
  { value: "idle", label: "AFK" },
  { value: "invisible", label: "Invisible" },
];

// prefix is how Discord renders the activity, e.g. "Listening to Spotify"
export const activityTypeOptions = [
  { value: "0", label: "Playing", prefix: "Playing" },
  { value: "1", label: "Streaming", prefix: "Streaming" },
  { value: "2", label: "Listening", prefix: "Listening to" },
  { value: "3", label: "Watching", prefix: "Watching" },
  { value: "5", label: "Competing", prefix: "Competing in" },
  { value: "4", label: "Custom", prefix: "" },
];
