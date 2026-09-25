import { ReactNode } from "react";
import {
  InputOTP,
  InputOTPGroup,
  InputOTPSeparator,
  InputOTPSlot,
} from "../ui/input-otp";
import { cn } from "@/lib/utils";

const codeLength = 8;
const groupSize = codeLength / 2;

const slotClassName = "h-12 w-10 font-mono text-xl";

export function ShareCodePanel({ children }: { children: ReactNode }) {
  return (
    <div className="flex min-h-36 flex-col items-center justify-center gap-3 rounded-lg border bg-muted/40 px-6 py-8">
      {children}
    </div>
  );
}

export function ShareCodeInput({
  value,
  onChange,
  onSubmit,
}: {
  value: string;
  onChange: (value: string) => void;
  onSubmit: () => void;
}) {
  return (
    <InputOTP
      maxLength={codeLength}
      pattern="^[a-zA-Z0-9]+$"
      value={value}
      onChange={(value) => onChange(value.toUpperCase())}
      onKeyDown={(e) => e.key === "Enter" && onSubmit()}
      autoFocus
    >
      <InputOTPGroup>
        {Array.from({ length: groupSize }, (_, i) => (
          <InputOTPSlot
            key={i}
            index={i}
            className={cn(slotClassName, "bg-background")}
          />
        ))}
      </InputOTPGroup>
      <InputOTPSeparator className="text-muted-foreground" />
      <InputOTPGroup>
        {Array.from({ length: groupSize }, (_, i) => (
          <InputOTPSlot
            key={i}
            index={groupSize + i}
            className={cn(slotClassName, "bg-background")}
          />
        ))}
      </InputOTPGroup>
    </InputOTP>
  );
}

// Looks like ShareCodeInput so both sides of a share match.
export function ShareCodeDisplay({ code }: { code: string | null }) {
  const chars = code?.split("") ?? [];

  const group = (offset: number) => (
    <div className="flex items-center">
      {Array.from({ length: groupSize }, (_, i) => (
        <div
          key={i}
          className={cn(
            slotClassName,
            "flex items-center justify-center border-y border-r border-input bg-background first:rounded-l-md first:border-l last:rounded-r-md",
            !code && "text-muted-foreground/40"
          )}
        >
          {chars[offset + i] ?? "•"}
        </div>
      ))}
    </div>
  );

  return (
    <div className="flex items-center gap-2" aria-label={code ?? undefined}>
      {group(0)}
      <InputOTPSeparator className="text-muted-foreground" />
      {group(groupSize)}
    </div>
  );
}
