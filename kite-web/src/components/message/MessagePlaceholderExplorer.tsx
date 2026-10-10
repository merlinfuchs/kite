import { useMemo, useState } from "react";
import PlaceholderExplorer from "../common/PlaceholderExplorer";
import { VariableIcon } from "lucide-react";
import { interactionPlaceholders } from "@/lib/flow/placeholders";

export default function MessagePlaceholderExplorer({
  onSelect,
}: {
  onSelect: (value: string) => void;
}) {
  const [context, setContext] = useState<"interaction" | "event">(
    "interaction"
  );

  // The same placeholders as in a flow, as messages are sent from one.
  const placeholders = useMemo(
    () =>
      interactionPlaceholders(
        context === "event" ? "event_discord" : "command"
      ),
    [context]
  );

  return (
    <div className="absolute top-10 right-1.5 z-20">
      <PlaceholderExplorer
        onSelect={onSelect}
        placeholders={placeholders}
        tab={context}
        tabs={[
          {
            label: "Interaction",
            value: "interaction",
          },
          {
            label: "Event",
            value: "event",
          },
        ]}
        onTabChange={(tab) => {
          setContext(tab as "interaction" | "event");
        }}
      >
        <VariableIcon
          className="h-5.5 w-5.5 text-muted-foreground hover:text-foreground cursor-pointer"
          role="button"
        />
      </PlaceholderExplorer>
    </div>
  );
}
