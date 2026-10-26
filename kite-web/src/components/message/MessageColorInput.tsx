import { useState } from "react";
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
  value: number | string | undefined;
  onChange: (value: number | string | undefined) => void;
  validation?: ValidationTarget;
}) {
  // Once typed in, text stays text even when it's emptied.
  const [textMode, setTextMode] = useState(false);

  if (textMode || typeof value === "string") {
    return (
      <div className="w-full">
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
        <Button
          variant="link"
          size="sm"
          className="px-0 h-auto mt-1"
          onClick={() => {
            setTextMode(false);
            if (typeof value === "string") onChange(undefined);
          }}
        >
          Use the color picker
        </Button>
      </div>
    );
  }

  return (
    <div className="w-full">
      <MessageInput
        type="color"
        label={label}
        value={value}
        onChange={onChange}
        validation={validation}
      />
      <Button
        variant="link"
        size="sm"
        className="px-0 h-auto mt-1"
        onClick={() => setTextMode(true)}
      >
        Use a placeholder instead
      </Button>
    </div>
  );
}
