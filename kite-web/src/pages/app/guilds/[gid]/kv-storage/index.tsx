import AppGuildLayout from "@/components/app/AppGuildLayout";
import AppKVStorageBrowser from "@/components/app/AppKVStorageBrowser";
import { useRouteParams } from "@/hooks/route";

export default function GuildKVPage() {
  const { guildId } = useRouteParams();

  return (
    <AppGuildLayout>
      <div>
        <div className="text-4xl font-bold text-white mb-4">KV Storage</div>
        <div className="text-lg font-light text-gray-300 mb-10">
          The key value storage is a simple way for plugins to store data. Each
          key belongs to a namespace which acts as a group of keys. Usually
          plugins will use their own unique namespace to store data, so they
          don't have to worry about key collisions.
        </div>
        <AppKVStorageBrowser guildId={guildId} />
      </div>
    </AppGuildLayout>
  );
}
