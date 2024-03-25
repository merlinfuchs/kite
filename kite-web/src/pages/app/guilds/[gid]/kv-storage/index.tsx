import AppGuildLayout from "@/components/app/AppGuildLayout";
import AppGuildPageHeader from "@/components/app/AppGuildPageHeader";
import AppKVStorageBrowser from "@/components/app/AppKVStorageBrowser";
import { useRouteParams } from "@/hooks/route";

export default function GuildKVPage() {
  const { guildId } = useRouteParams();

  return (
    <AppGuildLayout>
      <AppGuildPageHeader
        title="KV Storage"
        description="The key value storage is a simple way for plugins to store data. Each
          key belongs to a namespace which acts as a group of keys. Usually
          plugins will use their own unique namespace to store data, so they
          don't have to worry about key collisions."
      />
      <AppKVStorageBrowser guildId={guildId} />
    </AppGuildLayout>
  );
}
