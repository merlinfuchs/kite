import { useEffect, useState } from "react";
import JsonSchemaExplorer from "./JsonSchemaExplorer";
import { cn } from "../lib/util";
import type { JsonSchema7Type } from "zod-to-json-schema";

type NodeInfo = {
  title: string;
  description: string;
  color: string;
  dataSchema: JsonSchema7Type | null;
  resultSchema: JsonSchema7Type | null;
  dataFields: string[];
  creditsCost: number | null;
};

export default function NodeInfoExplorer({ type }: { type: string }) {
  const [tab, setTab] = useState<"data" | "result">("result");
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

  useEffect(() => {
    if (data?.resultSchema) {
      setTab("result");
    } else if (data?.dataSchema) {
      setTab("data");
    }
  }, [data]);

  if (!data) {
    return <div>Loading...</div>;
  }

  const schema = tab === "data" ? data.dataSchema : data.resultSchema;

  return (
    <div>
      <div className="flex border-zinc-300 dark:border-zinc-700 bg-zinc-100 dark:bg-zinc-800 p-1.5 rounded-lg mb-3">
        {data.resultSchema && (
          <div
            className={cn(
              "flex-1 px-3 py-1.5 font-bold rounded-md cursor-pointer transition-colors",
              tab === "result" && "bg-white dark:bg-zinc-700"
            )}
            role="button"
            onClick={() => setTab("result")}
          >
            Result Data
          </div>
        )}
        {data.dataSchema && (
          <div
            className={cn(
              "flex-1 px-3 py-1.5 font-bold rounded-md cursor-pointer transition-colors",
              tab === "data" && "bg-white dark:bg-zinc-700"
            )}
            role="button"
            onClick={() => setTab("data")}
          >
            Input Data
          </div>
        )}
      </div>

      <div className="bg-zinc-100 dark:bg-zinc-800 px-6 py-3 rounded-lg">
        {schema && <JsonSchemaExplorer schema={schema} />}
      </div>
    </div>
  );
}
