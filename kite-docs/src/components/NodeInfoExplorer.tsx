import { useEffect, useState } from "react";

type NodeInfo = {
  title: string;
  description: string;
  color: string;
  dataSchema: any | null;
  dataFields: string[];
  creditsCost: number | null;
};

export default function NodeInfoExplorer({ type }: { type: string }) {
  const [data, setData] = useState<NodeInfo | null>(null);

  useEffect(() => {
    const abortController = new AbortController();

    fetch(`http://localhost:3000/api/flow/nodes/${type}`, {
      signal: abortController.signal,
    })
      .then((res) => res.json())
      .then((data) => {
        console.log("data", data);
        setData(data);
      })
      .catch((error) => {
        // Only log errors that aren't from cancellation
        if (error.name !== "AbortError") {
          console.error("Failed to fetch node info:", error);
        }
      });

    return () => {
      abortController.abort();
    };
  }, [type]);

  console.log(data);

  if (!data) {
    return <div>Loading...</div>;
  }

  return null;
}
