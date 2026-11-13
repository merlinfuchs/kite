import { useValidationErrors } from "@/lib/message/state";
import { ValidationTarget } from "@/lib/message/validationStore";
import BaseInput, { BaseInputProps } from "@/tools/common/components/BaseInput";
import { insertAtCursor } from "@/lib/insertText";
import { useCallback, useRef } from "react";
import EmojiInsertButton from "../common/EmojiInsertButton";
import MessagePlaceholderExplorer from "./MessagePlaceholderExplorer";

type Props = BaseInputProps & {
  validation?: ValidationTarget;
  placeholders?: boolean;
  // Adds an emoji picker, "native" without custom emojis for fields that
  // Discord shows as plain text.
  emojis?: boolean | "native";
};

export default function MessageInput(props: Props) {
  const { validation, placeholders, emojis, ...inputProps } = props;

  const containerRef = useRef<HTMLDivElement>(null);

  const issue = useValidationErrors(
    (state) => validation && state.getIssue(validation)?.message
  );

  const insertText = useCallback(
    (text: string) => {
      const element = containerRef.current?.querySelector<
        HTMLInputElement | HTMLTextAreaElement
      >("input, textarea");
      insertAtCursor(element, text, (value) => props.onChange(value as never));
    },
    [props]
  );

  const onPlaceholderSelect = useCallback(
    (placeholder: string, raw?: boolean) =>
      insertText(raw ? placeholder : `{{${placeholder}}}`),
    [insertText]
  );

  const hasInput = ["text", "url", "email", "textarea"].includes(props.type);
  const showPlaceholders = !!placeholders && hasInput;
  const showEmojis =
    !!emojis && (props.type === "text" || props.type === "textarea");

  const actions = (showPlaceholders || showEmojis) && (
    <>
      {showEmojis && (
        <EmojiInsertButton
          onSelect={insertText}
          nativeOnly={emojis === "native"}
        />
      )}
      {showPlaceholders && (
        <MessagePlaceholderExplorer onSelect={onPlaceholderSelect} />
      )}
    </>
  );

  return (
    <div className="w-full" ref={containerRef}>
      <BaseInput
        {...(inputProps as BaseInputProps)}
        {...(hasInput && { actions })}
        error={issue}
      />
    </div>
  );
}
