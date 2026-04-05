import { create } from "zustand";
import { immer } from "zustand/middleware/immer";
import {
  MessageComponentButtonStyle,
  EmbedField,
  Message,
  MessageComponentActionRow,
  MessageComponentButton,
  MessageEmbed,
  MessageComponentSelectMenuOption,
  MessageComponentSelectMenu,
  Emoji,
  MessageAttachment,
  MessageComponentSection,
  MessageComponentTextDisplay,
  MessageComponentThumbnail,
  MessageComponentMediaGallery,
  MessageComponentMediaGalleryItem,
  MessageComponentFile,
  MessageComponentSeparator,
  MessageComponentContainer,
  MediaItem,
} from "./schema";
import { getUniqueId } from "@/lib/utils";
import { temporal } from "zundo";
import debounce from "just-debounce-it";

export interface MessageStore extends Message {
  clear(): void;
  reset(): void;
  replace(message: Message): void;
  setContent: (content: string) => void;
  setUsername: (username: string | undefined) => void;
  setAvatarUrl: (avatar_url: string | undefined) => void;
  setThreadName: (thread_name: string | undefined) => void;
  addAttachment: (attachment: MessageAttachment) => void;
  clearAttachments: () => void;
  deleteAttachment: (i: number) => void;
  addEmbed: (embed: MessageEmbed) => void;
  clearEmbeds: () => void;
  moveEmbedDown: (i: number) => void;
  moveEmbedUp: (i: number) => void;
  duplicateEmbed: (i: number) => void;
  deleteEmbed: (i: number) => void;
  setEmbedDescription: (i: number, description: string | undefined) => void;
  setEmbedTitle: (i: number, title: string | undefined) => void;
  setEmbedUrl: (i: number, url: string | undefined) => void;
  setEmbedAuthorName: (i: number, name: string) => void;
  setEmbedAuthorUrl: (i: number, url: string | undefined) => void;
  setEmbedAuthorIconUrl: (i: number, icon_url: string | undefined) => void;
  setEmbedThumbnailUrl: (i: number, url: string | undefined) => void;
  setEmbedImageUrl: (i: number, url: string | undefined) => void;
  setEmbedFooterText: (i: number, text: string | undefined) => void;
  setEmbedFooterIconUrl: (i: number, icon_url: string | undefined) => void;
  setEmbedColor: (i: number, color: number | undefined) => void;
  setEmbedTimestamp: (i: number, timestamp: string | undefined) => void;
  addEmbedField: (i: number, field: EmbedField) => void;
  setEmbedFieldName: (i: number, j: number, name: string) => void;
  setEmbedFieldValue: (i: number, j: number, value: string) => void;
  setEmbedFieldInline: (
    i: number,
    j: number,
    inline: boolean | undefined
  ) => void;
  moveEmbedFieldDown: (i: number, j: number) => void;
  moveEmbedFieldUp: (i: number, j: number) => void;
  deleteEmbedField: (i: number, j: number) => void;
  duplicateEmbedField: (i: number, j: number) => void;
  clearEmbedFields: (i: number) => void;
  
  // Component Row methods
  addComponentRow: (row: MessageComponentActionRow) => void;
  addContainer: (container: MessageComponentContainer) => void;
  clearComponentRows: () => void;
  moveComponentRowUp: (i: number) => void;
  moveComponentRowDown: (i: number) => void;
  duplicateComponentRow: (i: number) => void;
  deleteComponentRow: (i: number) => void;
  
  // Button methods
  addButton: (i: number, button: MessageComponentButton) => void;
  clearButtons: (i: number) => void;
  moveButtonDown: (i: number, j: number) => void;
  moveButtonUp: (i: number, j: number) => void;
  duplicateButton: (i: number, j: number) => void;
  deleteButton: (i: number, j: number) => void;
  setButtonStyle: (
    i: number,
    j: number,
    style: MessageComponentButtonStyle
  ) => void;
  setButtonLabel: (i: number, j: number, label: string) => void;
  setButtonEmoji: (i: number, j: number, emoji: Emoji | undefined) => void;
  setButtonUrl: (i: number, j: number, url: string) => void;
  setButtonDisabled: (
    i: number,
    j: number,
    disabled: boolean | undefined
  ) => void;
  
  // Select Menu methods
  setSelectMenuPlaceholder: (
    i: number,
    j: number,
    placeholder: string | undefined
  ) => void;
  setSelectMenuDisabled: (
    i: number,
    j: number,
    disabled: boolean | undefined
  ) => void;
  setSelectMenuMinValues: (
    i: number,
    j: number,
    minValues: number | undefined
  ) => void;
  setSelectMenuMaxValues: (
    i: number,
    j: number,
    maxValues: number | undefined
  ) => void;
  addSelectMenuOption: (
    i: number,
    j: number,
    option: MessageComponentSelectMenuOption
  ) => void;
  clearSelectMenuOptions: (i: number, j: number) => void;
  moveSelectMenuOptionDown: (i: number, j: number, k: number) => void;
  moveSelectMenuOptionUp: (i: number, j: number, k: number) => void;
  duplicateSelectMenuOption: (i: number, j: number, k: number) => void;
  deleteSelectMenuOption: (i: number, j: number, k: number) => void;
  setSelectMenuOptionLabel: (
    i: number,
    j: number,
    k: number,
    label: string
  ) => void;
  setSelectMenuOptionDescription: (
    i: number,
    j: number,
    k: number,
    description: string | undefined
  ) => void;
  setSelectMenuOptionEmoji: (
    i: number,
    j: number,
    k: number,
    emoji: Emoji | undefined
  ) => void;
  setSelectMenuOptionDefault: (
    i: number,
    j: number,
    k: number,
    isDefault: boolean | undefined
  ) => void;

  // Section (Type 9) methods
  addSection: (i: number, section: MessageComponentSection) => void;
  addTextDisplayToSection: (i: number, j: number, textDisplay: MessageComponentTextDisplay) => void;
  deleteTextDisplayFromSection: (i: number, j: number, k: number) => void;
  setTextDisplayContent: (i: number, j: number, k: number, content: string) => void;
  setSectionAccessory: (
    i: number,
    j: number,
    accessory: MessageComponentThumbnail | MessageComponentButton | undefined
  ) => void;

  // Text Display (Type 10) methods - when standalone in container
  addTextDisplay: (i: number, textDisplay: MessageComponentTextDisplay) => void;
  setStandaloneTextDisplayContent: (i: number, j: number, content: string) => void;

  // Media Gallery (Type 12) methods
  addMediaGallery: (i: number, gallery: MessageComponentMediaGallery) => void;
  addMediaGalleryItem: (i: number, j: number, item: MessageComponentMediaGalleryItem) => void;
  deleteMediaGalleryItem: (i: number, j: number, k: number) => void;
  setMediaGalleryItemUrl: (i: number, j: number, k: number, url: string) => void;
  setMediaGalleryItemDescription: (
    i: number,
    j: number,
    k: number,
    description: string | undefined
  ) => void;
  setMediaGalleryItemSpoiler: (
    i: number,
    j: number,
    k: number,
    spoiler: boolean | undefined
  ) => void;
  moveMediaGalleryItemUp: (i: number, j: number, k: number) => void;
  moveMediaGalleryItemDown: (i: number, j: number, k: number) => void;

  // File (Type 13) methods
  addFile: (i: number, file: MessageComponentFile) => void;
  setFileUrl: (i: number, j: number, url: string) => void;
  setFileSpoiler: (i: number, j: number, spoiler: boolean | undefined) => void;

  // Separator (Type 14) methods
  addSeparator: (i: number, separator: MessageComponentSeparator) => void;
  setSeparatorDivider: (i: number, j: number, divider: boolean) => void;
  setSeparatorSpacing: (i: number, j: number, spacing: 1 | 2) => void;

  // Container (Type 17) methods
  setContainerAccentColor: (i: number, color: number | undefined) => void;
  setContainerSpoiler: (i: number, spoiler: boolean | undefined) => void;
  addComponentToContainer: (i: number, component: any) => void;
  deleteComponentFromContainer: (i: number, j: number) => void;
  moveContainerComponentUp: (i: number, j: number) => void;
  moveContainerComponentDown: (i: number, j: number) => void;

  // Getter methods
  getSelectMenu: (i: number, j: number) => MessageComponentSelectMenu | null;
  getButton: (i: number, j: number) => MessageComponentButton | null;
  getSection: (i: number, j: number) => MessageComponentSection | null;
  getContainer: (i: number) => MessageComponentContainer | null;
}

export const emptyMessage: Message = {
  username: undefined,
  avatar_url: undefined,
  content: "",
  tts: false,
  attachments: [],
  embeds: [],
  components: [],
};

export const createMessageStore = (initial?: Message) => {
  const defaultMessage = initial || emptyMessage;

  return create<MessageStore>()(
    immer(
      temporal(
        (set, get) => ({
          ...defaultMessage,

          clear: () => set(emptyMessage),
          reset: () => set(defaultMessage),
          replace: (message: Message) => set(message),
          setContent: (content: string) => set({ content }),
          setUsername: (username: string | undefined) => set({ username }),
          setAvatarUrl: (avatar_url: string | undefined) => set({ avatar_url }),
          setThreadName: (thread_name: string | undefined) =>
            set({ thread_name }),
          addAttachment: (attachment: MessageAttachment) =>
            set((state) => {
              if (!state.attachments) {
                state.attachments = [attachment];
              } else {
                state.attachments.push(attachment);
              }
            }),
          clearAttachments: () => set({ attachments: [] }),
          deleteAttachment: (i: number) =>
            set((state) => {
              if (!state.attachments) {
                return;
              }
              state.attachments.splice(i, 1);
            }),
          addEmbed: (embed: MessageEmbed) =>
            set((state) => {
              if (!state.embeds) {
                state.embeds = [embed];
              } else {
                state.embeds.push(embed);
              }
            }),
          clearEmbeds: () => set({ embeds: [] }),
          moveEmbedDown: (i: number) => {
            set((state) => {
              if (!state.embeds) {
                return;
              }
              const embed = state.embeds[i];
              if (!embed) {
                return;
              }
              state.embeds.splice(i, 1);
              state.embeds.splice(i + 1, 0, embed);
            });
          },
          moveEmbedUp: (i: number) => {
            set((state) => {
              if (!state.embeds) {
                return;
              }
              const embed = state.embeds[i];
              if (!embed) {
                return;
              }
              state.embeds.splice(i, 1);
              state.embeds.splice(i - 1, 0, embed);
            });
          },
          duplicateEmbed: (i: number) => {
            set((state) => {
              if (!state.embeds) {
                return;
              }
              const embed = state.embeds[i];
              if (!embed) {
                return;
              }
              state.embeds.splice(i + 1, 0, { ...embed, id: getUniqueId() });
            });
          },
          deleteEmbed: (i: number) => {
            set((state) => {
              if (!state.embeds) {
                return;
              }
              state.embeds.splice(i, 1);
            });
          },
          setEmbedDescription: (i: number, description: string | undefined) => {
            set((state) => {
              if (state.embeds && state.embeds[i]) {
                state.embeds[i].description = description;
              }
            });
          },
          setEmbedTitle: (i: number, title: string | undefined) => {
            set((state) => {
              if (state.embeds && state.embeds[i]) {
                state.embeds[i].title = title;
              }
            });
          },
          setEmbedUrl: (i: number, url: string | undefined) => {
            set((state) => {
              if (state.embeds && state.embeds[i]) {
                state.embeds[i].url = url;
              }
            });
          },
          setEmbedAuthorName: (i: number, name: string) =>
            set((state) => {
              const embed = state.embeds && state.embeds[i];
              if (!embed) {
                return;
              }
              if (!name) {
                if (!embed.author) {
                  return;
                }

                embed.author.name = name;
                if (!embed.author.icon_url && !embed.author.url) {
                  embed.author = undefined;
                }
              } else {
                if (!embed.author) {
                  embed.author = { name };
                } else {
                  embed.author.name = name;
                }
              }
            }),
          setEmbedAuthorUrl: (i: number, url: string | undefined) =>
            set((state) => {
              const embed = state.embeds && state.embeds[i];
              if (!embed) {
                return;
              }
              if (!url) {
                if (!embed.author) {
                  return;
                }
                embed.author.url = undefined;

                if (!embed.author.name && !embed.author.icon_url) {
                  embed.author = undefined;
                }
              } else {
                if (!embed.author) {
                  embed.author = { name: "", url };
                } else {
                  embed.author.url = url;
                }
              }
            }),
          setEmbedAuthorIconUrl: (i: number, icon_url: string | undefined) =>
            set((state) => {
              const embed = state.embeds && state.embeds[i];
              if (!embed) {
                return;
              }
              if (!icon_url) {
                if (!embed.author) {
                  return;
                }
                embed.author.icon_url = undefined;

                if (!embed.author.name && !embed.author.url) {
                  embed.author = undefined;
                }
              } else {
                if (!embed.author) {
                  embed.author = { name: "", icon_url };
                } else {
                  embed.author.icon_url = icon_url;
                }
              }
            }),
          setEmbedThumbnailUrl: (i: number, url: string | undefined) => {
            set((state) => {
              const embed = state.embeds && state.embeds[i];
              if (!embed) {
                return;
              }
              if (!url) {
                embed.thumbnail = undefined;
              } else {
                embed.thumbnail = { url };
              }
            });
          },
          setEmbedImageUrl: (i: number, url: string | undefined) => {
            set((state) => {
              const embed = state.embeds && state.embeds[i];
              if (!embed) {
                return;
              }
              if (!url) {
                embed.image = undefined;
              } else {
                embed.image = { url };
              }
            });
          },
          setEmbedFooterText: (i: number, text: string | undefined) =>
            set((state) => {
              const embed = state.embeds && state.embeds[i];
              if (!embed) {
                return;
              }
              if (!text) {
                if (!embed.footer) {
                  return;
                }
                embed.footer.text = undefined;

                if (!embed.footer.icon_url) {
                  embed.footer = undefined;
                }
              } else {
                if (!embed.footer) {
                  embed.footer = { text };
                } else {
                  embed.footer.text = text;
                }
              }
            }),
          setEmbedFooterIconUrl: (i: number, icon_url: string | undefined) =>
            set((state) => {
              const embed = state.embeds && state.embeds[i];
              if (!embed) {
                return;
              }
              if (!icon_url) {
                if (!embed.footer) {
                  return;
                }
                embed.footer.icon_url = undefined;

                if (!embed.footer.text) {
                  embed.footer = undefined;
                }
              } else {
                if (!embed.footer) {
                  embed.footer = { icon_url };
                } else {
                  embed.footer.icon_url = icon_url;
                }
              }
            }),
          setEmbedColor: (i: number, color: number | undefined) =>
            set((state) => {
              const embed = state.embeds && state.embeds[i];
              if (!embed) {
                return;
              }
              embed.color = color;
            }),
          setEmbedTimestamp: (i: number, timestamp: string | undefined) =>
            set((state) => {
              const embed = state.embeds && state.embeds[i];
              if (!embed) {
                return;
              }
              embed.timestamp = timestamp;
            }),
          addEmbedField: (i: number, field: EmbedField) =>
            set((state) => {
              const embed = state.embeds && state.embeds[i];
              if (!embed) {
                return;
              }
              if (!embed.fields) {
                embed.fields = [field];
              } else {
                embed.fields.push(field);
              }
            }),
          setEmbedFieldName: (i: number, j: number, name: string) =>
            set((state) => {
              const embed = state.embeds && state.embeds[i];
              if (!embed) {
                return;
              }
              const field = embed.fields && embed.fields[j];
              if (!field) {
                return;
              }
              field.name = name;
            }),
          setEmbedFieldValue: (i: number, j: number, value: string) =>
            set((state) => {
              const embed = state.embeds && state.embeds[i];
              if (!embed) {
                return;
              }
              const field = embed.fields && embed.fields[j];
              if (!field) {
                return;
              }
              field.value = value;
            }),
          setEmbedFieldInline: (
            i: number,
            j: number,
            inline: boolean | undefined
          ) =>
            set((state) => {
              const embed = state.embeds && state.embeds[i];
              if (!embed) {
                return;
              }
              const field = embed.fields && embed.fields[j];
              if (!field) {
                return;
              }
              field.inline = inline;
            }),
          moveEmbedFieldUp: (i: number, j: number) => {
            set((state) => {
              const embed = state.embeds && state.embeds[i];
              if (!embed) {
                return;
              }
              const field = embed.fields && embed.fields[j];
              if (!field) {
                return;
              }
              embed.fields && embed.fields.splice(j, 1);
              embed.fields && embed.fields.splice(j - 1, 0, field);
            });
          },
          moveEmbedFieldDown: (i: number, j: number) => {
            set((state) => {
              const embed = state.embeds && state.embeds[i];
              if (!embed) {
                return;
              }
              const field = embed.fields && embed.fields[j];
              if (!field) {
                return;
              }
              embed.fields && embed.fields.splice(j, 1);
              embed.fields && embed.fields.splice(j + 1, 0, field);
            });
          },
          deleteEmbedField: (i: number, j: number) =>
            set((state) => {
              const embed = state.embeds && state.embeds[i];
              if (!embed) {
                return;
              }
              embed.fields && embed.fields.splice(j, 1);
            }),
          duplicateEmbedField: (i: number, j: number) => {
            set((state) => {
              const embed = state.embeds && state.embeds[i];
              if (!embed) {
                return;
              }
              const field = embed.fields && embed.fields[j];
              if (!field) {
                return;
              }
              embed.fields &&
                embed.fields.splice(j + 1, 0, {
                  ...field,
                  id: getUniqueId(),
                });
            });
          },
          clearEmbedFields: (i: number) =>
            set((state) => {
              const embed = state.embeds && state.embeds[i];
              if (!embed) {
                return;
              }
              embed.fields = [];
            }),
          
          // Component Row methods
          addComponentRow: (row: MessageComponentActionRow) =>
            set((state) => {
              if (!state.components) {
                state.components = [row];
              } else {
                state.components.push(row);
              }
            }),
          addContainer: (container: MessageComponentContainer) =>
            set((state) => {
              if (!state.components) {
                state.components = [container];
              } else {
                state.components.push(container);
              }
            }),
          clearComponentRows: () =>
            set((state) => {
              state.components = [];
            }),
          moveComponentRowUp: (i: number) =>
            set((state) => {
              const row = state.components && state.components[i];
              if (!row) {
                return;
              }
              state.components.splice(i, 1);
              state.components.splice(i - 1, 0, row);
            }),
          moveComponentRowDown: (i: number) =>
            set((state) => {
              const row = state.components && state.components[i];
              if (!row) {
                return;
              }
              state.components.splice(i, 1);
              state.components.splice(i + 1, 0, row);
            }),
          duplicateComponentRow: (i: number) =>
            set((state) => {
              const row = state.components && state.components[i];
              if (!row) {
                return;
              }

              if (row.type === 1) {
                // Action Row
                const newRow: MessageComponentActionRow = {
                  id: getUniqueId(),
                  type: 1,
                  components: row.components.map((comp) => {
                    const flowSourceId = getUniqueId().toString();
                    return { ...comp, flow_source_id: flowSourceId, id: getUniqueId() };
                  }),
                };
                state.components.splice(i + 1, 0, newRow);
              } else if (row.type === 17) {
                // Container - duplicate with new IDs
                const newContainer: MessageComponentContainer = {
                  ...row,
                  id: getUniqueId(),
                  components: row.components.map((comp: any) => ({
                    ...comp,
                    id: getUniqueId(),
                  })),
                };
                state.components.splice(i + 1, 0, newContainer);
              }
            }),
          deleteComponentRow: (i: number) =>
            set((state) => {
              state.components.splice(i, 1);
            }),
          
          // Button methods
          addButton: (i: number, button: MessageComponentButton) =>
            set((state) => {
              const row = state.components && state.components[i];
              if (!row || row.type !== 1) {
                return;
              }

              if (!row.components) {
                row.components = [button];
              } else {
                row.components.push(button);
              }
            }),
          clearButtons: (i: number) =>
            set((state) => {
              const row = state.components && state.components[i];
              if (!row || row.type !== 1) {
                return;
              }

              row.components = [];
            }),
          deleteButton: (i: number, j: number) =>
            set((state) => {
              const row = state.components[i];
              if (!row || row.type !== 1) {
                return;
              }

              row.components.splice(j, 1);
            }),
          moveButtonUp: (i: number, j: number) =>
            set((state) => {
              const row = state.components[i];
              if (!row || row.type !== 1) {
                return;
              }
              const button = row.components[j];
              if (!button) {
                return;
              }
              row.components.splice(j, 1);
              row.components.splice(j - 1, 0, button);
            }),
          moveButtonDown: (i: number, j: number) =>
            set((state) => {
              const row = state.components && state.components[i];
              if (!row || row.type !== 1) {
                return;
              }
              const button = row.components[j];
              if (!button) {
                return;
              }
              row.components.splice(j, 1);
              row.components.splice(j + 1, 0, button);
            }),
          duplicateButton: (i: number, j: number) =>
            set((state) => {
              const row = state.components && state.components[i];
              if (!row || row.type !== 1) {
                return;
              }
              const button = row.components && row.components[j];
              if (!button || button.type !== 2) {
                return;
              }

              const actionId = getUniqueId().toString();

              row.components.splice(j + 1, 0, {
                ...button,
                id: getUniqueId(),
                flow_source_id: actionId,
              });
            }),
          setButtonStyle: (
            i: number,
            j: number,
            style: MessageComponentButtonStyle
          ) =>
            set((state) => {
              const row = state.components && state.components[i];
              if (!row || row.type !== 1) {
                return;
              }
              const button = row.components && row.components[j];
              if (!button || button.type !== 2) {
                return;
              }

              button.style = style;
              if (button.style === 5) {
                button.url = "";
              }
            }),
          setButtonLabel: (i: number, j: number, label: string) =>
            set((state) => {
              const row = state.components && state.components[i];
              if (!row || row.type !== 1) {
                return;
              }
              const button = row.components && row.components[j];
              if (!button || button.type !== 2) {
                return;
              }
              button.label = label;
            }),
          setButtonEmoji: (i: number, j: number, emoji: Emoji | undefined) =>
            set((state) => {
              const row = state.components && state.components[i];
              if (!row || row.type !== 1) {
                return;
              }
              const button = row.components && row.components[j];
              if (!button || button.type !== 2) {
                return;
              }
              button.emoji = emoji;
            }),
          setButtonUrl: (i: number, j: number, url: string) =>
            set((state) => {
              const row = state.components && state.components[i];
              if (!row || row.type !== 1) {
                return;
              }
              const button = row.components && row.components[j];
              if (!button || button.type !== 2) {
                return;
              }
              button.url = url;
            }),
          setButtonDisabled: (
            i: number,
            j: number,
            disabled: boolean | undefined
          ) =>
            set((state) => {
              const row = state.components && state.components[i];
              if (!row || row.type !== 1) {
                return;
              }
              const button = row.components && row.components[j];
              if (!button || button.type !== 2) {
                return;
              }
              button.disabled = disabled;
            }),
          
          // Select Menu methods
          setSelectMenuPlaceholder: (
            i: number,
            j: number,
            placeholder: string | undefined
          ) =>
            set((state) => {
              const row = state.components && state.components[i];
              if (!row || row.type !== 1) {
                return;
              }
              const selectMenu = row.components && row.components[j];
              if (!selectMenu || selectMenu.type !== 3) {
                return;
              }
              selectMenu.placeholder = placeholder;
            }),
          setSelectMenuDisabled: (
            i: number,
            j: number,
            disabled: boolean | undefined
          ) =>
            set((state) => {
              const row = state.components && state.components[i];
              if (!row || row.type !== 1) {
                return;
              }
              const selectMenu = row.components && row.components[j];
              if (!selectMenu || selectMenu.type !== 3) {
                return;
              }
              selectMenu.disabled = disabled;
            }),
          setSelectMenuMinValues: (
            i: number,
            j: number,
            minValues: number | undefined
          ) =>
            set((state) => {
              const row = state.components && state.components[i];
              if (!row || row.type !== 1) {
                return;
              }
              const selectMenu = row.components && row.components[j];
              if (!selectMenu || selectMenu.type !== 3) {
                return;
              }
              selectMenu.min_values = minValues;
            }),
          setSelectMenuMaxValues: (
            i: number,
            j: number,
            maxValues: number | undefined
          ) =>
            set((state) => {
              const row = state.components && state.components[i];
              if (!row || row.type !== 1) {
                return;
              }
              const selectMenu = row.components && row.components[j];
              if (!selectMenu || selectMenu.type !== 3) {
                return;
              }
              selectMenu.max_values = maxValues;
            }),
          addSelectMenuOption: (
            i: number,
            j: number,
            option: MessageComponentSelectMenuOption
          ) =>
            set((state) => {
              const row = state.components && state.components[i];
              if (!row || row.type !== 1) {
                return;
              }
              const selectMenu = row.components && row.components[j];
              if (!selectMenu || selectMenu.type !== 3) {
                return;
              }
              if (!selectMenu.options) {
                selectMenu.options = [option];
              } else {
                selectMenu.options.push(option);
              }
            }),
          clearSelectMenuOptions: (i: number, j: number) =>
            set((state) => {
              const row = state.components && state.components[i];
              if (!row || row.type !== 1) {
                return;
              }
              const selectMenu = row.components && row.components[j];
              if (!selectMenu || selectMenu.type !== 3) {
                return;
              }
              selectMenu.options = [];
            }),
          moveSelectMenuOptionDown: (i: number, j: number, k: number) =>
            set((state) => {
              const row = state.components && state.components[i];
              if (!row || row.type !== 1) {
                return;
              }
              const selectMenu = row.components && row.components[j];
              if (!selectMenu || selectMenu.type !== 3) {
                return;
              }
              const option = selectMenu.options[k];
              if (!option) {
                return;
              }
              selectMenu.options.splice(k, 1);
              selectMenu.options.splice(k + 1, 0, option);
            }),
          moveSelectMenuOptionUp: (i: number, j: number, k: number) =>
            set((state) => {
              const row = state.components && state.components[i];
              if (!row || row.type !== 1) {
                return;
              }
              const selectMenu = row.components && row.components[j];
              if (!selectMenu || selectMenu.type !== 3) {
                return;
              }
              const option = selectMenu.options[k];
              if (!option) {
                return;
              }
              selectMenu.options.splice(k, 1);
              selectMenu.options.splice(k - 1, 0, option);
            }),
          duplicateSelectMenuOption: (i: number, j: number, k: number) =>
            set((state) => {
              const row = state.components && state.components[i];
              if (!row || row.type !== 1) {
                return;
              }
              const selectMenu = row.components && row.components[j];
              if (!selectMenu || selectMenu.type !== 3) {
                return;
              }
              const option = selectMenu.options[k];
              if (!option) {
                return;
              }

              selectMenu.options.splice(k + 1, 0, {
                ...option,
                id: getUniqueId(),
                flow_source_id: getUniqueId().toString(),
              });
            }),
          deleteSelectMenuOption: (i: number, j: number, k: number) =>
            set((state) => {
              const row = state.components && state.components[i];
              if (!row || row.type !== 1) {
                return;
              }
              const selectMenu = row.components && row.components[j];
              if (!selectMenu || selectMenu.type !== 3) {
                return;
              }

              selectMenu.options.splice(k, 1);
            }),
          setSelectMenuOptionLabel: (
            i: number,
            j: number,
            k: number,
            label: string
          ) =>
            set((state) => {
              const row = state.components && state.components[i];
              if (!row || row.type !== 1) {
                return;
              }
              const selectMenu = row.components && row.components[j];
              if (!selectMenu || selectMenu.type !== 3) {
                return;
              }
              const option = selectMenu.options && selectMenu.options[k];
              if (!option) {
                return;
              }
              option.label = label;
            }),
          setSelectMenuOptionDescription: (
            i: number,
            j: number,
            k: number,
            description: string | undefined
          ) =>
            set((state) => {
              const row = state.components && state.components[i];
              if (!row || row.type !== 1) {
                return;
              }
              const selectMenu = row.components && row.components[j];
              if (!selectMenu || selectMenu.type !== 3) {
                return;
              }
              const option = selectMenu.options && selectMenu.options[k];
              if (!option) {
                return;
              }
              option.description = description;
            }),
          setSelectMenuOptionEmoji: (
            i: number,
            j: number,
            k: number,
            emoji: Emoji | undefined
          ) =>
            set((state) => {
              const row = state.components && state.components[i];
              if (!row || row.type !== 1) {
                return;
              }
              const selectMenu = row.components && row.components[j];
              if (!selectMenu || selectMenu.type !== 3) {
                return;
              }
              const option = selectMenu.options && selectMenu.options[k];
              if (!option) {
                return;
              }
              option.emoji = emoji;
            }),
          setSelectMenuOptionDefault: (
            i: number,
            j: number,
            k: number,
            isDefault: boolean | undefined
          ) =>
            set((state) => {
              const row = state.components && state.components[i];
              if (!row || row.type !== 1) {
                return;
              }
              const selectMenu = row.components && row.components[j];
              if (!selectMenu || selectMenu.type !== 3) {
                return;
              }
              const option = selectMenu.options && selectMenu.options[k];
              if (!option) {
                return;
              }
              option.default = isDefault;
            }),

          // ===================================================================
          // COMPONENTS V2 METHODS
          // ===================================================================

          // Section (Type 9) methods
          addSection: (i: number, section: MessageComponentSection) =>
            set((state) => {
              const container = state.components && state.components[i];
              if (!container || container.type !== 17) {
                return;
              }
              container.components.push(section);
            }),
          addTextDisplayToSection: (
            i: number,
            j: number,
            textDisplay: MessageComponentTextDisplay
          ) =>
            set((state) => {
              const container = state.components && state.components[i];
              if (!container || container.type !== 17) {
                return;
              }
              const section = container.components[j];
              if (!section || section.type !== 9) {
                return;
              }
              section.components.push(textDisplay);
            }),
          deleteTextDisplayFromSection: (i: number, j: number, k: number) =>
            set((state) => {
              const container = state.components && state.components[i];
              if (!container || container.type !== 17) {
                return;
              }
              const section = container.components[j];
              if (!section || section.type !== 9) {
                return;
              }
              section.components.splice(k, 1);
            }),
          setTextDisplayContent: (
            i: number,
            j: number,
            k: number,
            content: string
          ) =>
            set((state) => {
              const container = state.components && state.components[i];
              if (!container || container.type !== 17) {
                return;
              }
              const section = container.components[j];
              if (!section || section.type !== 9) {
                return;
              }
              const textDisplay = section.components[k];
              if (!textDisplay || textDisplay.type !== 10) {
                return;
              }
              textDisplay.content = content;
            }),
          setSectionAccessory: (
            i: number,
            j: number,
            accessory: MessageComponentThumbnail | MessageComponentButton | undefined
          ) =>
            set((state) => {
              const container = state.components && state.components[i];
              if (!container || container.type !== 17) {
                return;
              }
              const section = container.components[j];
              if (!section || section.type !== 9) {
                return;
              }
              section.accessory = accessory;
            }),

          // Text Display (Type 10) methods - standalone in container
          addTextDisplay: (i: number, textDisplay: MessageComponentTextDisplay) =>
            set((state) => {
              const container = state.components && state.components[i];
              if (!container || container.type !== 17) {
                return;
              }
              container.components.push(textDisplay);
            }),
          setStandaloneTextDisplayContent: (i: number, j: number, content: string) =>
            set((state) => {
              const container = state.components && state.components[i];
              if (!container || container.type !== 17) {
                return;
              }
              const textDisplay = container.components[j];
              if (!textDisplay || textDisplay.type !== 10) {
                return;
              }
              textDisplay.content = content;
            }),

          // Media Gallery (Type 12) methods
          addMediaGallery: (i: number, gallery: MessageComponentMediaGallery) =>
            set((state) => {
              const container = state.components && state.components[i];
              if (!container || container.type !== 17) {
                return;
              }
              container.components.push(gallery);
            }),
          addMediaGalleryItem: (
            i: number,
            j: number,
            item: MessageComponentMediaGalleryItem
          ) =>
            set((state) => {
              const container = state.components && state.components[i];
              if (!container || container.type !== 17) {
                return;
              }
              const gallery = container.components[j];
              if (!gallery || gallery.type !== 12) {
                return;
              }
              gallery.items.push(item);
            }),
          deleteMediaGalleryItem: (i: number, j: number, k: number) =>
            set((state) => {
              const container = state.components && state.components[i];
              if (!container || container.type !== 17) {
                return;
              }
              const gallery = container.components[j];
              if (!gallery || gallery.type !== 12) {
                return;
              }
              gallery.items.splice(k, 1);
            }),
          setMediaGalleryItemUrl: (i: number, j: number, k: number, url: string) =>
            set((state) => {
              const container = state.components && state.components[i];
              if (!container || container.type !== 17) {
                return;
              }
              const gallery = container.components[j];
              if (!gallery || gallery.type !== 12) {
                return;
              }
              const item = gallery.items[k];
              if (!item) {
                return;
              }
              item.media.url = url;
            }),
          setMediaGalleryItemDescription: (
            i: number,
            j: number,
            k: number,
            description: string | undefined
          ) =>
            set((state) => {
              const container = state.components && state.components[i];
              if (!container || container.type !== 17) {
                return;
              }
              const gallery = container.components[j];
              if (!gallery || gallery.type !== 12) {
                return;
              }
              const item = gallery.items[k];
              if (!item) {
                return;
              }
              item.description = description;
            }),
          setMediaGalleryItemSpoiler: (
            i: number,
            j: number,
            k: number,
            spoiler: boolean | undefined
          ) =>
            set((state) => {
              const container = state.components && state.components[i];
              if (!container || container.type !== 17) {
                return;
              }
              const gallery = container.components[j];
              if (!gallery || gallery.type !== 12) {
                return;
              }
              const item = gallery.items[k];
              if (!item) {
                return;
              }
              item.spoiler = spoiler;
            }),
          moveMediaGalleryItemUp: (i: number, j: number, k: number) =>
            set((state) => {
              const container = state.components && state.components[i];
              if (!container || container.type !== 17) {
                return;
              }
              const gallery = container.components[j];
              if (!gallery || gallery.type !== 12) {
                return;
              }
              const item = gallery.items[k];
              if (!item) {
                return;
              }
              gallery.items.splice(k, 1);
              gallery.items.splice(k - 1, 0, item);
            }),
          moveMediaGalleryItemDown: (i: number, j: number, k: number) =>
            set((state) => {
              const container = state.components && state.components[i];
              if (!container || container.type !== 17) {
                return;
              }
              const gallery = container.components[j];
              if (!gallery || gallery.type !== 12) {
                return;
              }
              const item = gallery.items[k];
              if (!item) {
                return;
              }
              gallery.items.splice(k, 1);
              gallery.items.splice(k + 1, 0, item);
            }),

          // File (Type 13) methods
          addFile: (i: number, file: MessageComponentFile) =>
            set((state) => {
              const container = state.components && state.components[i];
              if (!container || container.type !== 17) {
                return;
              }
              container.components.push(file);
            }),
          setFileUrl: (i: number, j: number, url: string) =>
            set((state) => {
              const container = state.components && state.components[i];
              if (!container || container.type !== 17) {
                return;
              }
              const file = container.components[j];
              if (!file || file.type !== 13) {
                return;
              }
              file.file.url = url;
            }),
          setFileSpoiler: (i: number, j: number, spoiler: boolean | undefined) =>
            set((state) => {
              const container = state.components && state.components[i];
              if (!container || container.type !== 17) {
                return;
              }
              const file = container.components[j];
              if (!file || file.type !== 13) {
                return;
              }
              file.spoiler = spoiler;
            }),

          // Separator (Type 14) methods
          addSeparator: (i: number, separator: MessageComponentSeparator) =>
            set((state) => {
              const container = state.components && state.components[i];
              if (!container || container.type !== 17) {
                return;
              }
              container.components.push(separator);
            }),
          setSeparatorDivider: (i: number, j: number, divider: boolean) =>
            set((state) => {
              const container = state.components && state.components[i];
              if (!container || container.type !== 17) {
                return;
              }
              const separator = container.components[j];
              if (!separator || separator.type !== 14) {
                return;
              }
              separator.divider = divider;
            }),
          setSeparatorSpacing: (i: number, j: number, spacing: 1 | 2) =>
            set((state) => {
              const container = state.components && state.components[i];
              if (!container || container.type !== 17) {
                return;
              }
              const separator = container.components[j];
              if (!separator || separator.type !== 14) {
                return;
              }
              separator.spacing = spacing;
            }),

          // Container (Type 17) methods
          setContainerAccentColor: (i: number, color: number | undefined) =>
            set((state) => {
              const container = state.components && state.components[i];
              if (!container || container.type !== 17) {
                return;
              }
              container.accent_color = color;
            }),
          setContainerSpoiler: (i: number, spoiler: boolean | undefined) =>
            set((state) => {
              const container = state.components && state.components[i];
              if (!container || container.type !== 17) {
                return;
              }
              container.spoiler = spoiler;
            }),
          addComponentToContainer: (i: number, component: any) =>
            set((state) => {
              const container = state.components && state.components[i];
              if (!container || container.type !== 17) {
                return;
              }
              container.components.push(component);
            }),
          deleteComponentFromContainer: (i: number, j: number) =>
            set((state) => {
              const container = state.components && state.components[i];
              if (!container || container.type !== 17) {
                return;
              }
              container.components.splice(j, 1);
            }),
          moveContainerComponentUp: (i: number, j: number) =>
            set((state) => {
              const container = state.components && state.components[i];
              if (!container || container.type !== 17) {
                return;
              }
              const component = container.components[j];
              if (!component) {
                return;
              }
              container.components.splice(j, 1);
              container.components.splice(j - 1, 0, component);
            }),
          moveContainerComponentDown: (i: number, j: number) =>
            set((state) => {
              const container = state.components && state.components[i];
              if (!container || container.type !== 17) {
                return;
              }
              const component = container.components[j];
              if (!component) {
                return;
              }
              container.components.splice(j, 1);
              container.components.splice(j + 1, 0, component);
            }),

          // Getter methods
          getSelectMenu: (i: number, j: number) => {
            const state = get();
            const row = state.components && state.components[i];
            if (!row || row.type !== 1) {
              return null;
            }

            const selectMenu = row.components && row.components[j];
            if (selectMenu && selectMenu.type === 3) {
              return selectMenu;
            }
            return null;
          },
          getButton: (i: number, j: number) => {
            const state = get();
            const row = state.components && state.components[i];
            if (!row || row.type !== 1) {
              return null;
            }

            const button = row.components && row.components[j];
            if (button && button.type === 2) {
              return button;
            }
            return null;
          },
          getSection: (i: number, j: number) => {
            const state = get();
            const container = state.components && state.components[i];
            if (!container || container.type !== 17) {
              return null;
            }

            const section = container.components && container.components[j];
            if (section && section.type === 9) {
              return section;
            }
            return null;
          },
          getContainer: (i: number) => {
            const state = get();
            const container = state.components && state.components[i];
            if (container && container.type === 17) {
              return container;
            }
            return null;
          },
        }),
        {
          limit: 10,
          handleSet: (handleSet) => debounce(handleSet, 1000, true),
        }
      )
    )
  );
};