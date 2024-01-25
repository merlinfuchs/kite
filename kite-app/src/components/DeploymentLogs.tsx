import { useDeploymentLogEntriesQuery } from "@/lib/api/queries";
import clsx from "clsx";

interface Props {
  guildId: string;
  deploymentId: string;
}

export default function ({ guildId, deploymentId }: Props) {
  const { data: resp } = useDeploymentLogEntriesQuery(guildId, deploymentId);

  const logEntries = resp?.success ? resp.data : [];

  return logEntries.map((l, i) => (
    <div key={l.id} className="font-light text-sm flex space-x-2">
      <div className="text-gray-400 min-w-32">
        {new Date(l.created_at).toLocaleString()}
      </div>
      <div className="min-w-16 flex">
        <div
          className={clsx(
            "text-gray-100 px-1 rounded",
            l.level === "ERROR"
              ? "bg-red-500"
              : l.level === "WARN"
              ? "bg-yellow-500"
              : l.level === "INFO"
              ? "bg-blue-500"
              : "bg-gray-700"
          )}
        >
          {l.level}
        </div>
      </div>
      <div className="text-gray-300">{l.message}</div>
    </div>
  ));
}
