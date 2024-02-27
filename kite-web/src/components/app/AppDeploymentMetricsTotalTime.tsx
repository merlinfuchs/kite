import { useDeploymentsEventMetricsQuery } from "@/lib/api/queries";
import { ApexOptions } from "apexcharts";
import Chart from "react-apexcharts";

interface Props {
  guildId: string;
  deploymentId?: string | null;
}

export default function GuildMetricsTiming({ guildId, deploymentId }: Props) {
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
    yaxis: [
      {
        title: {
          text: "Total Time (ms)",
        },
      },
    ],
    annotations: {
      yaxis: [
        {
          y: 10000,
          borderColor: "#FF4560",
          label: {
            borderColor: "#FF4560",
            style: {
              color: "#fff",
              background: "#FF4560",
            },
            text: "Total time limit (10 seconds)",
          },
        },
      ],
    },
  };

  const series: ApexAxisChartSeries | ApexNonAxisChartSeries = [
    {
      name: "Total Time (ms)",
      data: metrics.map((m) => Math.floor(m.average_total_time * 10) / 10),
      //color: "#FF4560",
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
