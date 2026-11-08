import { MessageComponentSelectMenuType } from "./schema";

export const selectMenuTypes: {
  type: MessageComponentSelectMenuType;
  label: string;
  /** What the flow receives in interaction.values when something is picked. */
  values: string;
}[] = [
  { type: 3, label: "Select menu", values: "the values of the picked options" },
  { type: 5, label: "User select", values: "the IDs of the picked users" },
  { type: 6, label: "Role select", values: "the IDs of the picked roles" },
  {
    type: 7,
    label: "Mentionable select",
    values: "the IDs of the picked users and roles",
  },
  {
    type: 8,
    label: "Channel select",
    values: "the IDs of the picked channels",
  },
];

export function selectMenuType(type: number | undefined) {
  return (
    selectMenuTypes.find((t) => t.type === (type ?? 3)) ?? selectMenuTypes[0]
  );
}

/** The channel types a channel select can be limited to. */
export const channelTypeOptions = [
  { label: "Text", value: "0" },
  { label: "Voice", value: "2" },
  { label: "Category", value: "4" },
  { label: "Announcement", value: "5" },
  { label: "Announcement thread", value: "10" },
  { label: "Public thread", value: "11" },
  { label: "Private thread", value: "12" },
  { label: "Stage", value: "13" },
  { label: "Forum", value: "15" },
  { label: "Media", value: "16" },
];
