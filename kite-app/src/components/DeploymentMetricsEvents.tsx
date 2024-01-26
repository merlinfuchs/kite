import { useDeploymentsEventMetricsQuery } from "@/lib/api/queries";
import { ApexOptions } from "apexcharts";
import Chart from "react-apexcharts";

interface Props {
  guildId: string;
  deploymentId?: string | null;
}

export default function GuildMetricsEvents({ guildId, deploymentId }: Props) {
  const { data: metricsResp } = useDeploymentsEventMetricsQuery(
    guildId,
    deploymentId
  );

  const metrics = metricsResp?.success ? metricsResp.data : [];

  const options: ApexOptions = {
    dataLabels: {
      enabled: false,
    },
    stroke: {
      curve: "smooth",
    },
    xaxis: {
      type: "datetime",
      categories: metrics.map((m) => m.timestamp),
    },
    tooltip: {
      x: {
        format: "dd/MM/yy HH:mm",
      },
    },
    chart: {
      foreColor: "#6B737F",
      stacked: true,
      toolbar: {
        show: false,
      },
    },
  };

  const series: ApexAxisChartSeries | ApexNonAxisChartSeries = [
    {
      name: "Events Failed",
      data: metrics.map((m) => m.total_count - m.success_count),
      color: "#FF4560",
    },
    {
      name: "Events Handled",
      data: metrics.map((m) => m.success_count),
      color: "#00E396",
    },
  ];

  return (
    <Chart
      options={options}
      series={series}
      type="area"
      width="100%"
      height={300}
    />
  );
}
