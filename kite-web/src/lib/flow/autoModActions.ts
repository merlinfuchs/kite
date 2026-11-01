// autoModActions are the actions of AutoMod the Event Filter can pick, as the
// service names them in guildevent.AutoModAction.ActionName.
export const autoModActions: { value: string; label: string }[] = [
  { value: "block_message", label: "Blocked a Message" },
  { value: "send_alert_message", label: "Sent an Alert" },
  { value: "timeout", label: "Timed Out a Member" },
  {
    value: "block_member_interaction",
    label: "Blocked a Member's Interaction",
  },
];
