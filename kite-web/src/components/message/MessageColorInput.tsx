import { useState } from "react";
import { EmbedColor } from "@/lib/message/schema";
import { useNoPlaceholders } from "@/lib/message/state";
import { ValidationTarget } from "@/lib/message/validationStore";
import { Button } from "../ui/button";
import MessageInput from "./MessageInput";

// A color from the picker, or text with placeholders that flows fill in.
export default function MessageColorInput({
  label,
  value,
  onChange,
  validation,
}: {
  label: string;
  value: EmbedColor;
  onChange: (value: EmbedColor) => void;
  validation?: ValidationTarget;
}) {
  // Nothing fills in placeholders here, like in the message creator, which
  // sends straight to Discord.
  const noPlaceholders = useNoPlaceholders();
  // Once typed in, text stays text even when it's emptied. A number from
  // undo or an import shows the picker again.
  const [textMode, setTextMode] = useState(false);
  const isText =
    typeof value === "string" ||
    (textMode && !noPlaceholders && value === undefined);

  return (
    <div className="w-full">
      {isText ? (
        <MessageInput
          type="text"
          label={label}
          value={typeof value === "string" ? value : ""}
          onChange={(v) => {
            setTextMode(true);
            onChange(v || undefined);
          }}
          validation={validation}
          placeholders
        />
      ) : (
        <MessageInput
          type="color"
          label={label}
          value={value}
          onChange={onChange}
          validation={validation}
        />
      )}
      {(isText || !noPlaceholders) && (
        <Button
          variant="link"
          size="sm"
          className="px-0 h-auto mt-1"
          onClick={() => {
            setTextMode(!isText);
            // The other input can't show it, so it would be set unseen.
            onChange(undefined);
          }}
        >
          {isText ? "Use the color picker" : "Use a placeholder instead"}
        </Button>
      )}
    </div>
  );
}
