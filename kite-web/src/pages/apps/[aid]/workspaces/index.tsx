import AppLayout from "@/components/app/AppLayout";
import AppWorkspaceList from "@/components/app/AppWorkspaceList";
import { useRouteParams } from "@/hooks/route";

export default function AppWorkspacesPage() {
  const { appId } = useRouteParams();

  return (
    <AppLayout>
      <div>
        <div className="text-4xl font-bold text-white mb-4">Workspaces</div>
        <div className="text-lg font-light text-gray-300 mb-10">
          A workspace is like a online VS Code project that can contain an
          arbitrary number of files and is used to create a private deployment
          or a public plugin.
        </div>
        <AppWorkspaceList appId={appId} />
      </div>
    </AppLayout>
  );
}
