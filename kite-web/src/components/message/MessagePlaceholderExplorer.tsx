import { useMemo, useState } from "react";
import PlaceholderExplorer from "../common/PlaceholderExplorer";
import { VariableIcon } from "lucide-react";

export default function MessagePlaceholderExplorer({
  onSelect,
}: {
  onSelect: (value: string) => void;
}) {
  const [context, setContext] = useState<"interaction" | "event">(
    "interaction"
  );

  const globalPlaceholders = useGlobalPlaceholders(context);

  const placeholders = useMemo(
    () => [...globalPlaceholders],
    [globalPlaceholders]
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

function useGlobalPlaceholders(baseKey: "interaction" | "event") {
  return useMemo(() => {
    const res = [
      {
        label: "User",
        placeholders: [
          {
            label: "User ID",
            value: `${baseKey}.user.id`,
          },
          {
            label: "User Mention",
            value: `${baseKey}.user.mention`,
          },
          {
            label: "User Username",
            value: `${baseKey}.user.username`,
          },
          {
            label: "User Discriminator",
            value: `${baseKey}.user.discriminator`,
          },
          {
            label: "User Display Name",
            value: `${baseKey}.user.display_name`,
          },
          {
            label: "User Avatar URL",
            value: `${baseKey}.user.avatar_url`,
          },
          {
            label: "User Banner URL",
            value: `${baseKey}.user.banner_url`,
          },
        ],
      },
      {
        label: "Server",
        placeholders: [
          {
            label: "Server ID",
            value: `${baseKey}.guild.id`,
          },
        ],
      },
      {
        label: "Channel",
        placeholders: [
          {
            label: "Channel ID",
            value: `${baseKey}.channel.id`,
          },
        ],
      },
    ];

    if (baseKey === "event") {
      res.push({
        label: "Message",
        placeholders: [
          { label: "Message ID", value: `${baseKey}.message.id` },
          { label: "Message Content", value: `${baseKey}.message.content` },
        ],
      });
    }
    return res;
  }, [baseKey]);
}
