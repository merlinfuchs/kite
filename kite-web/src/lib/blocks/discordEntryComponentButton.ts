import { BlockDefinition } from "./types";

export const discordEntryComponentButton: BlockDefinition = {
  type: "entry_component_button",
  title: "Button",
  description:
    "This gets triggered when a user clicks the button. Drop different actions here!",
  icon: "mouse-pointer-click",
  contexts: ["component_button", "component_select_menu"],
  fixed: true,
  component: "entry_component_button",
  requires: ["discord"],
  custom_label: false,
  run: { kind: "custom" },
};
