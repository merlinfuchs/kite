import AppLayout from "@/components/AppLayout";
import DeploymentLogs from "@/components/DeploymentLogs";
import { useRouter } from "next/router";

export default function DeploymentPage() {
  const router = useRouter();
  const guildId = router.query.gid as string;
  const deploymentId = router.query.did as string;

  return (
    <AppLayout>
      <div className="p-5 bg-slate-800 rounded-md space-y-2 h-[500px] overflow-y-scroll flex flex-col justify-end">
        <DeploymentLogs guildId={guildId} deploymentId={deploymentId} />
      </div>
    </AppLayout>
  );
}
