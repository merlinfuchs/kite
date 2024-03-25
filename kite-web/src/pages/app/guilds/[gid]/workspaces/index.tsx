import AppGuildLayout from "@/components/app/AppGuildLayout";
import AppGuildPageHeader from "@/components/app/AppGuildPageHeader";
import AppWorkspaceList from "@/components/app/AppWorkspaceList";
import { useRouteParams } from "@/hooks/route";

export default function GuildWorkspacesPage() {
  const { guildId } = useRouteParams();

  return (
    <AppGuildLayout>
      <AppGuildPageHeader
        title="Workspaces"
        description="A workspace is like a online VS Code project that can contain an
        arbitrary number of files and is used to create a private deployment or
        a public plugin."
      />
      <AppWorkspaceList guildId={guildId} />
    </AppGuildLayout>
  );
}
