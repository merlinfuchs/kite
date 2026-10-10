import { MessageComponentSelectMenuType } from "./schema";

export const selectMenuTypes: {
  type: MessageComponentSelectMenuType;
  label: string;
  /** What the flow receives in interaction.values when something is picked. */
  values: string;
}[] = [
  { type: 3, label: "Select Menu", values: "the values of the picked options" },
  { type: 5, label: "User Select", values: "the IDs of the picked users" },
  { type: 6, label: "Role Select", values: "the IDs of the picked roles" },
  {
    type: 7,
    label: "Mentionable Select",
    values: "the IDs of the picked users and roles",
  },
  {
    type: 8,
    label: "Channel Select",
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
  { label: "Announcement Thread", value: "10" },
  { label: "Public Thread", value: "11" },
  { label: "Private Thread", value: "12" },
  { label: "Stage", value: "13" },
  { label: "Forum", value: "15" },
  { label: "Media", value: "16" },
];
