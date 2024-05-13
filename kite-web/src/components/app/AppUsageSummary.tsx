import {
  useAppEntitlementsResolvedQuery,
  useAppUsageSummaryQuery,
} from "@/lib/api/queries";
import { ApexOptions } from "apexcharts";
import Chart from "react-apexcharts";

interface Props {
  appId: string;
}

export default function AppUsageSummary({ appId }: Props) {
  const { data: usageData } = useAppUsageSummaryQuery(appId);
  const { data: entitlementsData } = useAppEntitlementsResolvedQuery(appId);

  const totalExecutionTime = usageData?.success
    ? usageData.data.total_event_execution_time
    : 0;
  const allowedExecutionTime = entitlementsData?.success
    ? entitlementsData.data.monthly_execution_time_limit
    : 1;

  const options: ApexOptions = {
    labels: ["Execution Time"],
    plotOptions: {
      radialBar: {
        hollow: {
          margin: 15,
          size: "65%",
        },

        track: {
          background: "#615d84",
        },

        dataLabels: {
          name: {
            offsetY: -10,
            show: true,
            color: "#C1C5CC",
            fontSize: "13px",
          },
          value: {
            color: "#F3F4F6",
            fontSize: "30px",
            show: true,
          },
        },
      },
    },
  };

  const series = [
    Math.round((totalExecutionTime / allowedExecutionTime) * 100),
  ];

  return (
    <Chart
      options={options}
      series={series}
      type="radialBar"
      width={250}
      height={250}
    />
  );
}
