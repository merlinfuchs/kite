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

function useGlobalPlaceholders(context: "interaction" | "event") {
  return useMemo(() => {
    const res = [
      {
        label: "User",
        placeholders: [
          {
            label: "User",
            value: `user`,
          },
          {
            label: "User ID",
            value: `user.id`,
          },
          {
            label: "User Mention",
            value: `user.mention`,
          },
          {
            label: "User Username",
            value: `user.username`,
          },
          {
            label: "User Discriminator",
            value: `user.discriminator`,
          },
          {
            label: "User Display Name",
            value: `user.display_name`,
          },
          {
            label: "User Avatar URL",
            value: `user.avatar_url`,
          },
          {
            label: "User Banner URL",
            value: `user.banner_url`,
          },
          {
            label: "User Is Bot",
            value: `user.is_bot`,
          },
          {
            label: "User Created At",
            value: `user.created_at`,
          },
          {
            label: "User Joined At",
            value: `user.joined_at`,
          },
          {
            label: "User Top Role",
            value: `user.top_role`,
          },
          {
            label: "User Role Mentions",
            value: `user.role_mentions`,
          },
          {
            label: "User Role Names",
            value: `user.role_names`,
          },
          {
            label: "User Role Count",
            value: `user.role_count`,
          },
          {
            label: "User Color",
            value: `user.color`,
          },
          {
            label: "User Is Booster",
            value: `user.is_booster`,
          },
          {
            label: "User Boosting Since",
            value: `user.boosting_since`,
          },
          {
            label: "User Is Timed Out",
            value: `user.is_timed_out`,
          },
          {
            label: "User Timeout Until",
            value: `user.timeout_until`,
          },
          {
            label: "User Is Owner",
            value: `user.is_owner`,
          },
          {
            label: "User Is Admin",
            value: `user.is_admin`,
          },
          {
            label: "User Permissions",
            value: `user.permissions`,
          },
        ],
      },
      {
        label: "Server",
        placeholders: [
          {
            label: "Server ID",
            value: `guild.id`,
          },
          {
            label: "Server Name",
            value: `guild.name`,
          },
          {
            label: "Server Icon URL",
            value: `guild.icon_url`,
          },
          {
            label: "Server Member Count",
            value: `guild.member_count`,
          },
          {
            label: "Server Boost Count",
            value: `guild.boost_count`,
          },
          {
            label: "Server Owner ID",
            value: `guild.owner_id`,
          },
          {
            label: "Server Boost Level",
            value: `guild.boost_level`,
          },
          {
            label: "Server Created At",
            value: `guild.created_at`,
          },
          {
            label: "Server Banner URL",
            value: `guild.banner_url`,
          },
          {
            label: "Server Description",
            value: `guild.description`,
          },
          {
            label: "Server Vanity URL",
            value: `guild.vanity_url`,
          },
          {
            label: "Server Role Count",
            value: `guild.role_count`,
          },
          {
            label: "Server Channel Count",
            value: `guild.channel_count`,
          },
          {
            label: "Server Emoji Count",
            value: `guild.emoji_count`,
          },
          {
            label: "Server Rules Channel",
            value: `guild.rules_channel`,
          },
          {
            label: "Server System Channel",
            value: `guild.system_channel`,
          },
        ],
      },
      {
        label: "Channel",
        placeholders: [
          {
            label: "Channel ID",
            value: `channel.id`,
          },
          {
            label: "Channel Name",
            value: `channel.name`,
          },
          {
            label: "Channel Mention",
            value: `channel.mention`,
          },
          {
            label: "Channel Type",
            value: `channel.type`,
          },
          {
            label: "Channel Category ID",
            value: `channel.category_id`,
          },
          {
            label: "Channel Category Name",
            value: `channel.category_name`,
          },
        ],
      },
    ];

    if (context === "event") {
      res.push({
        label: "Message",
        placeholders: [
          { label: "Message ID", value: `message.id` },
          { label: "Message Content", value: `message.content` },
        ],
      });
    }
    return res;
  }, [context]);
}
