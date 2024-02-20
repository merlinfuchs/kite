import { useDeploymentMetricsSummaryQuery } from "@/lib/api/queries";

interface Props {
  guildId: string;
}

export default function AppDeploymentMetricsSummary({ guildId }: Props) {
  const { data: metricsResp } = useDeploymentMetricsSummaryQuery(guildId);

  return <div></div>;
}
