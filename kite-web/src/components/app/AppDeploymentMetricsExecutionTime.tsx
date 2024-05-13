import { useDeploymentsEventMetricsQuery } from "@/lib/api/queries";
import { ApexOptions } from "apexcharts";
import Chart from "react-apexcharts";

interface Props {
  appId: string;
  deploymentId?: string | null;
}

export default function AppDeploymentMetricsExecutionTime({
  appId,
  deploymentId,
}: Props) {
  const { data: metricsResp } = useDeploymentsEventMetricsQuery(
    appId,
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
          text: "CPU Time (ms)",
        },
      },
    ],
    annotations: {
      yaxis: [
        {
          y: 10,
          borderColor: "#FF4560",
          label: {
            borderColor: "#FF4560",
            style: {
              color: "#fff",
              background: "#FF4560",
            },
            text: "CPU time limit (10 milliseconds)",
          },
        },
      ],
    },
  };

  const series: ApexAxisChartSeries | ApexNonAxisChartSeries = [
    {
      name: "CPU Time (ms)",
      data: metrics.map((m) => Math.floor(m.average_execution_time * 10) / 10),
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
