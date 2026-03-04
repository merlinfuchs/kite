import { useShallow } from "zustand/react/shallow";
import CollapsibleSection from "./MessageCollapsibleSection";
import { useCurrentMessage } from "@/lib/message/state";
import MessageInput from "./MessageInput";

export default function MessageEmbedFooter({
  embedId,
  embedIndex,
}: {
  embedId: number;
  embedIndex: number;
}) {
  const [text, setText] = useCurrentMessage(
    useShallow((state) => [
      state.embeds[embedIndex]?.footer?.text,
      state.setEmbedFooterText,
    ])
  );

  const [iconUrl, setIconUrl] = useCurrentMessage(
    useShallow((state) => [
      state.embeds[embedIndex]?.footer?.icon_url,
      state.setEmbedFooterIconUrl,
    ])
  );

  const [timestamp, setTimestamp] = useCurrentMessage(
    useShallow((state) => [
      state.embeds[embedIndex]?.timestamp,
      state.setEmbedTimestamp,
    ])
  );

  return (
    <CollapsibleSection
      title="Footer"
      size="md"
      valiationPathPrefix={`embeds.${embedIndex}.footer`}
      className="space-y-3"
    >
      <MessageInput
        type="text"
        label="Footer"
        maxLength={2048}
        value={text || ""}
        onChange={(v) => setText(embedIndex, v || undefined)}
        validationPath={`embeds.${embedIndex}.footer.text`}
        placeholders
      />
      <div className="flex space-x-3">
        <MessageInput
          type="url"
          label="Footer Icon URL"
          value={iconUrl || ""}
          onChange={(v) => setIconUrl(embedIndex, v || undefined)}
          validationPath={`embeds.${embedIndex}.footer.icon_url`}
          imageUpload
        />
        <MessageInput
          type="text"
          label="Timestamp"
          value={timestamp || ""}
          onChange={(v) => {
            if (!v) return setTimestamp(embedIndex, undefined);
            let input = v.trim();
            if (/^<t:.*?:[RrTtDdFf]?>$/.test(input)) return setTimestamp(embedIndex, input);
            input = input.replace(/\{\{now\(\)\.Unix\(\)\}\}/g, () =>
              Math.floor(Date.now() / 1000).toString()
            );
            const timezoneMatch = input.match(
              /\{\{now\(\)\.In\(timezone\("(.+?)"\)\)\.Format\("(.+?)"\)\}\}/
            );
            if (timezoneMatch) {
              const tz = timezoneMatch[1], fmt = timezoneMatch[2];
              try {
                const date = new Date();
                const options: Intl.DateTimeFormatOptions = {};
                if (fmt.includes("PM") || fmt.includes("AM") || fmt.includes("03") || fmt.includes("04")) {
                  options.hour = "2-digit"; options.minute = "2-digit";
                  if (fmt.includes("00")) options.second = "2-digit";
                }
                if (fmt.includes("02") || fmt.includes("01") || fmt.includes("2006")) {
                  options.year = "numeric"; options.month = "2-digit"; options.day = "2-digit";
                }
                input = new Intl.DateTimeFormat("en-US", { ...options, timeZone: tz }).format(date);
              } catch {}
            }
            if (/^\d+$/.test(input)) {
              const unix = Number(input);
              const date = input.length === 13 ? new Date(unix) : new Date(unix * 1000);
              if (!isNaN(date.getTime())) return setTimestamp(embedIndex, date.toISOString());
            }
            const parsed = new Date(input);
            if (!isNaN(parsed.getTime())) setTimestamp(embedIndex, parsed.toISOString());
          }}
          validationPath={`embeds.${embedIndex}.timestamp`}
        />
      </div>
    </CollapsibleSection>
  );
}
