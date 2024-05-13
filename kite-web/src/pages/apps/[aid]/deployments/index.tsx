import AppLayout from "@/components/app/AppLayout";
import AppDeploymentList from "@/components/app/AppDeploymentList";
import { useRouteParams } from "@/hooks/route";

export default function AppDeploymentsPage() {
  const { appId } = useRouteParams();

  return (
    <AppLayout>
      <div>
        <div className="text-4xl font-bold text-white mb-4">Deployments</div>
        <div className="text-lg font-light text-gray-300 mb-10">
          A deployment is a running instance of a plugin. You can create a
          deployment from a workspace or by using a brebuilt plugin from the
          marketplace.
        </div>
        <AppDeploymentList appId={appId} />
      </div>
    </AppLayout>
  );
}
