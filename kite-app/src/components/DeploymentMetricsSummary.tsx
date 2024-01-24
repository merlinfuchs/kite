"use client";

import { ApexOptions } from "apexcharts";
import Chart from "react-apexcharts";

interface Props {
  guildId: string;
  deploymentId: string;
}

const options: ApexOptions = {
  dataLabels: {
    enabled: false,
  },
  stroke: {
    curve: "smooth",
  },
  xaxis: {
    type: "datetime",
    categories: [
      "2018-09-19T00:00:00.000Z",
      "2018-09-19T01:30:00.000Z",
      "2018-09-19T02:30:00.000Z",
      "2018-09-19T03:30:00.000Z",
      "2018-09-19T04:30:00.000Z",
      "2018-09-19T05:30:00.000Z",
      "2018-09-19T06:30:00.000Z",
    ],
  },
  tooltip: {
    x: {
      format: "dd/MM/yy HH:mm",
    },
  },
  theme: {
    palette: "palette1",
  },
  chart: {
    foreColor: "#6B737F",
    toolbar: {
      show: false,
    },
  },
};

const series: ApexAxisChartSeries | ApexNonAxisChartSeries = [
  {
    name: "Events Handled",
    data: [31, 40, 28, 51, 42, 109, 100],
  },
  {
    name: "Calls Made",
    data: [11, 32, 45, 32, 34, 52, 41],
  },
];

export default function DeploymentMetricsSummary({
  guildId,
  deploymentId,
}: Props) {
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
