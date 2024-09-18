import { create, useStore } from "zustand";
import { persist } from "zustand/middleware";
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
} from "./schema";
import { getUniqueId } from "@/lib/utils";
import { TemporalState, temporal } from "zundo";
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
  addComponentRow: (row: MessageComponentActionRow) => void;
  clearComponentRows: () => void;
  moveComponentRowUp: (i: number) => void;
  moveComponentRowDown: (i: number) => void;
  duplicateComponentRow: (i: number) => void;
  deleteComponentRow: (i: number) => void;
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

  getSelectMenu: (i: number, j: number) => MessageComponentSelectMenu | null;
  getButton: (i: number, j: number) => MessageComponentButton | null;
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
                  embed.author = { url, name: "" };
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
                  embed.author = { icon_url, name: "" };
                } else {
                  embed.author.icon_url = icon_url;
                }
              }
            }),
          setEmbedThumbnailUrl: (i: number, url: string | undefined) => {
            set((state) => {
              if (state.embeds && state.embeds[i]) {
                state.embeds[i].thumbnail = url ? { url } : undefined;
              }
            });
          },
          setEmbedImageUrl: (i: number, url: string | undefined) => {
            set((state) => {
              if (state.embeds && state.embeds[i]) {
                state.embeds[i].image = url ? { url } : undefined;
              }
            });
          },
          setEmbedFooterText: (i: number, text: string | undefined) => {
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
            });
          },
          setEmbedFooterIconUrl: (i: number, icon_url: string | undefined) => {
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
            });
          },
          setEmbedColor: (i: number, color: number | undefined) => {
            set((state) => {
              if (state.embeds && state.embeds[i]) {
                state.embeds[i].color = color;
              }
            });
          },
          setEmbedTimestamp: (i: number, timestamp: string | undefined) => {
            set((state) => {
              if (state.embeds && state.embeds[i]) {
                state.embeds[i].timestamp = timestamp;
              }
            });
          },
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
          deleteEmbedField: (i: number, j: number) => {
            set((state) => {
              const embed = state.embeds && state.embeds[i];
              if (!embed) {
                return;
              }
              embed.fields && embed.fields.splice(j, 1);
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
          addComponentRow: (row: MessageComponentActionRow) =>
            set((state) => {
              if (!state.components) {
                state.components = [row];
              } else {
                state.components.push(row);
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

              // This is a bit complex because we can't allow duplicated action set ids
              const newRow: MessageComponentActionRow = {
                id: getUniqueId(),
                type: 1,
                components: row.components.map((comp) => {
                  if (comp.type === 2) {
                    const actionId = getUniqueId().toString();
                    return { ...comp, action_set_id: actionId };
                  } else {
                    return {
                      ...comp,
                      options: comp.options.map((option) => {
                        const actionId = getUniqueId().toString();
                        return {
                          ...option,
                          action_set_id: actionId,
                        };
                      }),
                    };
                  }
                }),
              };

              // TODO: change action set ids
              state.components.splice(i + 1, 0, newRow);
            }),
          deleteComponentRow: (i: number) =>
            set((state) => {
              state.components.splice(i, 1);
            }),
          addButton: (i: number, button: MessageComponentButton) =>
            set((state) => {
              const row = state.components && state.components[i];
              if (!row) {
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
              if (!row) {
                return;
              }

              row.components = [];
            }),
          deleteButton: (i: number, j: number) =>
            set((state) => {
              const row = state.components[i];
              if (!row) {
                return;
              }

              row.components.splice(j, 1);
            }),
          moveButtonUp: (i: number, j: number) =>
            set((state) => {
              const row = state.components[i];
              if (!row) {
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
              if (!row) {
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
              if (!row) {
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
                action_set_id: actionId,
              });
            }),
          setButtonStyle: (
            i: number,
            j: number,
            style: MessageComponentButtonStyle
          ) =>
            set((state) => {
              const row = state.components && state.components[i];
              if (!row) {
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
              if (!row) {
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
              if (!row) {
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
              if (!row) {
                return;
              }
              const button = row.components && row.components[j];
              if (!button || button.type !== 2 || button.style !== 5) {
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
              if (!row) {
                return;
              }
              const button = row.components && row.components[j];
              if (!button) {
                return;
              }
              button.disabled = disabled;
            }),
          setSelectMenuPlaceholder: (
            i: number,
            j: number,
            placeholder: string | undefined
          ) =>
            set((state) => {
              const row = state.components && state.components[i];
              if (!row) {
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
              if (!row) {
                return;
              }
              const selectMenu = row.components && row.components[j];
              if (!selectMenu || selectMenu.type !== 3) {
                return;
              }
              selectMenu.disabled = disabled;
            }),
          addSelectMenuOption: (
            i: number,
            j: number,
            option: MessageComponentSelectMenuOption
          ) =>
            set((state) => {
              const row = state.components && state.components[i];
              if (!row) {
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
              if (!row) {
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
              if (!row) {
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
              if (!row) {
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
              if (!row) {
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

              const actionId = getUniqueId().toString();
              selectMenu.options.splice(k + 1, 0, {
                ...option,
                id: getUniqueId(),
                action_set_id: actionId,
              });
            }),
          deleteSelectMenuOption: (i: number, j: number, k: number) =>
            set((state) => {
              const row = state.components && state.components[i];
              if (!row) {
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
              if (!row) {
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
              if (!row) {
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
              if (!row) {
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

          getSelectMenu: (i: number, j: number) => {
            const state = get();
            const row = state.components && state.components[i];
            if (!row) {
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
            if (!row) {
              return null;
            }

            const button = row.components && row.components[j];
            if (button && button.type === 2) {
              return button;
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
