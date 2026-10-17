import { useValidationErrors } from "@/lib/message/state";
import { ValidationTarget } from "@/lib/message/validationStore";
import BaseInput, { BaseInputProps } from "@/tools/common/components/BaseInput";
import { insertAtCursor } from "@/lib/insertText";
import { cn } from "@/lib/utils";
import { useCallback, useRef } from "react";
import EmojiInsertButton from "../common/EmojiInsertButton";
import MessagePlaceholderExplorer from "./MessagePlaceholderExplorer";

type Props = BaseInputProps & {
  validation?: ValidationTarget;
  placeholders?: boolean;
  // Adds an emoji picker to a text input, "native" without custom emojis for
  // fields that Discord shows as plain text.
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
      // TODO?: This is pretty hacky, we should think about baking placeholder support into the BaseInput component
      const element = containerRef.current?.querySelector<
        HTMLInputElement | HTMLTextAreaElement
      >("input, textarea");

      insertAtCursor(element, text, (value) => props.onChange(value as never));
    },
    [props]
  );

  const onPlaceholderSelect = useCallback(
    (placeholder: string) => insertText(`{{${placeholder}}}`),
    [insertText]
  );

  const isText = props.type === "text" || props.type === "textarea";
  const showEmojis = !!emojis && isText;

  // Keeps the text clear of the icons in the top right of the input.
  const iconPadding =
    placeholders && showEmojis
      ? "[&_input]:pr-16 [&_textarea]:pr-16"
      : placeholders || showEmojis
      ? "[&_input]:pr-9 [&_textarea]:pr-9"
      : undefined;

  return (
    <div className={cn("relative w-full", iconPadding)} ref={containerRef}>
      <BaseInput {...(inputProps as BaseInputProps)} error={issue} />
      {placeholders && (
        <MessagePlaceholderExplorer onSelect={onPlaceholderSelect} />
      )}
      {showEmojis && (
        <EmojiInsertButton
          onSelect={insertText}
          nativeOnly={emojis === "native"}
          className={placeholders ? "top-10 right-9" : "top-10 right-1.5"}
        />
      )}
    </div>
  );
}
