import { useCurrentMessage } from "@/lib/message/state";
import CollapsibleSection from "./MessageCollapsibleSection";
import MessageEmbed from "./MessageEmbed";
import { useShallow } from "zustand/react/shallow";
import { getUniqueId } from "@/lib/utils";
import { Button } from "@/components/ui/button";

export default function MessageEmbedSection() {
  const embeds = useCurrentMessage(
    useShallow((state) => state.embeds.map((e) => e.id))
  );
  const addEmbed = useCurrentMessage((state) => state.addEmbed);
  const clearEmbeds = useCurrentMessage((state) => state.clearEmbeds);

  return (
    <CollapsibleSection
      title="Embeds"
      valiationPathPrefix="embeds"
      className="space-y-4"
    >
      {embeds.map((id, i) => (
        <MessageEmbed key={id} embedIndex={i} embedId={id} />
      ))}
      <div className="space-x-3">
        <Button
          onClick={() =>
            addEmbed({
              id: getUniqueId(),
              description: "",
              fields: [],
            })
          }
        >
          Add Embed
        </Button>
        <Button onClick={clearEmbeds} variant="destructive">
          Clear Embeds
        </Button>
      </div>
    </CollapsibleSection>
  );
}
