import Twemoji from "@/components/common/Twemoji";
import MessagePreviewMarkup from "@/components/message/MessagePreviewMarkup";
import { modalInputIsMultiValue, modalMaxFiles } from "@/lib/flow/modal";
import {
  EmojiData,
  ModalComponentData,
  ModalComponentOptionData,
  ModalData,
} from "@/lib/types/flow.gen";
import { discordEmojiUrl } from "@/tools/common/utils/discordCdn";
import { cn } from "@/lib/utils";
import { CheckIcon, ChevronDownIcon, UploadIcon, XIcon } from "lucide-react";
import { useEffect, useState } from "react";

const defaultAvatarUrl = "https://cdn.discordapp.com/embed/avatars/3.png";

// The placeholders Discord shows in selects that don't set their own.
const entitySelectPlaceholders: Record<string, string> = {
  user_select: "Select a user",
  role_select: "Select a role",
  mentionable_select: "Select a user or role",
  channel_select: "Select a channel",
};

/**
 * A preview of a modal as Discord shows it. The inputs can be tried out, and
 * clicking a component selects it in the editor.
 */
export default function ModalPreview({
  modal,
  lightTheme,
  activeIndex,
  errorIndexes,
  onSelect,
}: {
  modal: ModalData;
  lightTheme?: boolean;
  activeIndex?: number | null;
  errorIndexes?: Set<number>;
  onSelect?: (index: number) => void;
}) {
  const components = modal.components ?? [];

  return (
    <div
      className={cn(
        "kite-modal-backdrop h-full",
        lightTheme && "discord-light-theme"
      )}
    >
      <div className="kite-modal" role="dialog" aria-label="Modal preview">
        <div className="kite-modal-header">
          <img src={defaultAvatarUrl} alt="" className="kite-modal-avatar" />
          <div
            className={cn(
              "kite-modal-title",
              !modal.title && "kite-modal-title-empty"
            )}
          >
            {modal.title || "Untitled modal"}
          </div>
          <XIcon className="kite-modal-close h-6 w-6" />
        </div>

        <div className="kite-modal-body">
          {components.length === 0 && (
            <div className="kite-modal-empty">Add an input to see it here.</div>
          )}
          {components.map((component, i) => (
            <div
              // The type and identifier are part of the key, so the answer
              // tried out in the preview is reset when they change.
              key={`${i}:${component.type}:${component.components?.[0]?.type}:${component.components?.[0]?.custom_id}`}
              className={cn(
                "kite-modal-component",
                activeIndex === i && "kite-modal-component-active",
                errorIndexes?.has(i) && "kite-modal-component-error"
              )}
              onClick={() => onSelect?.(i)}
            >
              {component.type === "text_display" ? (
                <PreviewTextDisplay content={component.content} />
              ) : (
                <PreviewLabel component={component} />
              )}
            </div>
          ))}
        </div>

        <div className="kite-modal-footer">
          <button type="button" className="kite-modal-button">
            Cancel
          </button>
          <button
            type="button"
            className="kite-modal-button kite-modal-button-primary"
          >
            Submit
          </button>
        </div>
      </div>
    </div>
  );
}

function PreviewTextDisplay({ content }: { content?: string }) {
  if (!content) {
    return (
      <div className="kite-modal-label kite-modal-label-empty">Empty text</div>
    );
  }

  return (
    <div className="discord-message kite-modal-text-display">
      <div className="discord-message-body">
        <MessagePreviewMarkup content={content} />
      </div>
    </div>
  );
}

function PreviewLabel({ component }: { component: ModalComponentData }) {
  const input = component.components?.[0] ?? {};
  const type = input.type || "text_input";

  const label = (
    <div
      className={cn(
        "kite-modal-label",
        !component.label && "kite-modal-label-empty"
      )}
    >
      {component.label || "Untitled input"}
      {input.required && type !== "checkbox" && (
        <span className="kite-modal-required">*</span>
      )}
    </div>
  );
  const description = component.description ? (
    <div className="kite-modal-description">{component.description}</div>
  ) : null;

  // A checkbox is shown next to its label instead of below it.
  if (type === "checkbox") {
    return <PreviewCheckbox input={input} label={label} desc={description} />;
  }

  return (
    <>
      {label}
      {description}
      <div className="kite-modal-field">
        {type === "text_input" ? (
          <PreviewTextInput input={input} />
        ) : type === "string_select" ? (
          <PreviewStringSelect input={input} />
        ) : type === "radio_group" || type === "checkbox_group" ? (
          <PreviewOptionGroup input={input} />
        ) : type === "file_upload" ? (
          <PreviewFileUpload input={input} />
        ) : (
          <PreviewEntitySelect input={input} />
        )}
      </div>
    </>
  );
}

function PreviewTextInput({ input }: { input: ModalComponentData }) {
  const [value, setValue] = useState(input.value || "");
  // Editing the pre-filled value shows it again.
  useEffect(() => setValue(input.value || ""), [input.value]);
  const paragraph = input.style === 2;
  const max = input.max_length ?? 4000;

  const props = {
    className: cn("kite-modal-input", paragraph && "kite-modal-textarea"),
    placeholder: input.placeholder,
    value,
    maxLength: max,
    onChange: (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) =>
      setValue(e.target.value),
  };

  return (
    <>
      {paragraph ? <textarea {...props} /> : <input type="text" {...props} />}
      {paragraph && (
        <div className="kite-modal-counter">
          {value.length}/{max}
        </div>
      )}
    </>
  );
}

function useOptionPicks(input: ModalComponentData) {
  const options = input.options ?? [];
  const multi =
    input.type === "checkbox_group" || modalInputIsMultiValue(input);

  const defaults = options.flatMap((o, i) => (o.default ? [i] : []));
  const initial = multi ? defaults : defaults.slice(0, 1);
  const initialKey = initial.join(",");

  const [picked, setPicked] = useState<number[]>(initial);
  // Changing the defaults or the options shows the defaults again.
  useEffect(() => {
    setPicked(initialKey ? initialKey.split(",").map(Number) : []);
  }, [initialKey, options.length]);

  const toggle = (i: number) => {
    if (!multi) {
      setPicked([i]);
      return;
    }
    if (picked.includes(i)) {
      setPicked(picked.filter((p) => p !== i));
      return;
    }
    const max =
      input.max_values ??
      (input.type === "checkbox_group" ? options.length : 1);
    if (picked.length >= max) return;
    setPicked([...picked, i].sort((a, b) => a - b));
  };

  return { options, multi, picked, toggle };
}

function PreviewStringSelect({ input }: { input: ModalComponentData }) {
  const [open, setOpen] = useState(false);
  const { options, multi, picked, toggle } = useOptionPicks(input);

  return (
    <div>
      <button
        type="button"
        className={cn("kite-modal-select", open && "kite-modal-select-open")}
        onClick={(e) => {
          e.stopPropagation();
          setOpen(!open);
        }}
      >
        <span className="kite-modal-select-value">
          {picked.length === 0 ? (
            <span className="kite-modal-select-placeholder">
              {input.placeholder || "Make a selection"}
            </span>
          ) : multi ? (
            picked.map((i) => (
              <span key={i} className="kite-modal-chip">
                <OptionEmoji emoji={options[i]?.emoji} />
                {options[i]?.label}
              </span>
            ))
          ) : (
            <span className="flex items-center gap-1.5">
              <OptionEmoji emoji={options[picked[0]]?.emoji} />
              {options[picked[0]]?.label}
            </span>
          )}
        </span>
        <ChevronDownIcon
          className={cn(
            "kite-modal-select-icon h-5 w-5 transition-transform",
            open && "rotate-180"
          )}
        />
      </button>
      {open && (
        <div className="kite-modal-select-menu">
          {options.map((option, i) => (
            <button
              type="button"
              key={i}
              className="kite-modal-option"
              onClick={(e) => {
                e.stopPropagation();
                toggle(i);
                if (!multi) setOpen(false);
              }}
            >
              <OptionEmoji emoji={option.emoji} large />
              <OptionText option={option} />
              {picked.includes(i) && <PickMark radio={!multi} checked />}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}

function PreviewOptionGroup({ input }: { input: ModalComponentData }) {
  const { options, picked, toggle } = useOptionPicks(input);
  const radio = input.type === "radio_group";

  return (
    <div className="kite-modal-options">
      {options.map((option, i) => {
        const checked = picked.includes(i);
        return (
          <button
            type="button"
            key={i}
            className="kite-modal-option"
            onClick={() => toggle(i)}
          >
            <PickMark radio={radio} checked={checked} />
            <OptionText option={option} />
          </button>
        );
      })}
    </div>
  );
}

function PreviewCheckbox({
  input,
  label,
  desc,
}: {
  input: ModalComponentData;
  label: React.ReactNode;
  desc: React.ReactNode;
}) {
  const [checked, setChecked] = useState(!!input.default);
  useEffect(() => setChecked(!!input.default), [input.default]);

  return (
    <button
      type="button"
      className="kite-modal-single-checkbox w-full text-left"
      onClick={() => setChecked(!checked)}
    >
      <PickMark checked={checked} />
      <span className="min-w-0">
        {label}
        {desc}
      </span>
    </button>
  );
}

// The radio button or checkbox of an option.
function PickMark({ radio, checked }: { radio?: boolean; checked: boolean }) {
  if (radio) {
    return (
      <span
        className={cn(
          "kite-modal-radio",
          checked && "kite-modal-radio-checked"
        )}
      />
    );
  }
  return (
    <span
      className={cn(
        "kite-modal-checkbox",
        checked && "kite-modal-checkbox-checked"
      )}
    >
      {checked && <CheckIcon className="h-3.5 w-3.5" strokeWidth={3} />}
    </span>
  );
}

function OptionText({ option }: { option: ModalComponentOptionData }) {
  return (
    <span className="kite-modal-option-text">
      <span className="kite-modal-option-label block">
        {option.label || "Untitled option"}
      </span>
      {option.description && (
        <span className="kite-modal-option-description block">
          {option.description}
        </span>
      )}
    </span>
  );
}

function PreviewEntitySelect({ input }: { input: ModalComponentData }) {
  return (
    <div className="kite-modal-select">
      <span className="kite-modal-select-value">
        <span className="kite-modal-select-placeholder">
          {input.placeholder ||
            entitySelectPlaceholders[input.type ?? ""] ||
            "Make a selection"}
        </span>
      </span>
      <ChevronDownIcon className="kite-modal-select-icon h-5 w-5" />
    </div>
  );
}

function PreviewFileUpload({ input }: { input: ModalComponentData }) {
  const max = Math.min(input.max_values ?? 1, modalMaxFiles);

  return (
    <div className="kite-modal-upload">
      <UploadIcon className="h-6 w-6" />
      <div>
        Drag and drop or{" "}
        <span className="kite-modal-upload-browse">browse</span>
      </div>
      <div className="text-xs opacity-80">
        {max === 1 ? "1 file" : `Up to ${max} files`}
      </div>
    </div>
  );
}

// Kept at module level, react-twemoji re-parses its children whenever the
// options object changes.
const emojiOptions = { className: "kite-modal-emoji" };
const largeEmojiOptions = {
  className: "kite-modal-emoji kite-modal-emoji-large",
};

function OptionEmoji({ emoji, large }: { emoji?: EmojiData; large?: boolean }) {
  if (!emoji?.id && !emoji?.name) return null;

  if (emoji.id) {
    return (
      <img
        src={discordEmojiUrl(emoji.id)}
        alt=""
        className={cn("kite-modal-emoji", large && "kite-modal-emoji-large")}
      />
    );
  }

  return (
    <Twemoji options={large ? largeEmojiOptions : emojiOptions}>
      {emoji.name}
    </Twemoji>
  );
}
