import { useState } from "react";
import { toast } from "sonner";
import { LoaderCircleIcon, SendIcon } from "lucide-react";
import { useHTTPRequestTestMutation } from "@/lib/api/mutations";
import { useAppId } from "@/lib/hooks/params";
import {
  formatResponseBody,
  httpRequestPlaceholders,
} from "@/lib/flow/httpRequestTest";
import { HTTPRequestData } from "@/lib/types/flow.gen";
import { HTTPRequestTestResponse } from "@/lib/types/wire.gen";
import { cn } from "@/lib/utils";
import { Button } from "../ui/button";
import { Input } from "../ui/input";

export default function HttpRequestTestPanel({
  data,
}: {
  data: HTTPRequestData;
}) {
  const appId = useAppId();
  const mutation = useHTTPRequestTestMutation(appId);

  const placeholders = httpRequestPlaceholders(data);
  const [values, setValues] = useState<Record<string, string>>({});
  const [result, setResult] = useState<HTTPRequestTestResponse | null>(null);

  function send() {
    mutation.mutate(
      {
        data,
        values: Object.fromEntries(
          placeholders.map((p) => [p, values[p] || ""])
        ),
      },
      {
        onSuccess: (res) => {
          if (res.success) {
            setResult(res.data);
          } else {
            toast.error(`Failed to send test request: ${res.error.message}`);
          }
        },
      }
    );
  }

  return (
    <div className="border-t pt-4 space-y-3 min-w-0">
      <div>
        <div className="font-medium text-foreground">Try it out</div>
        <div className="text-muted-foreground text-sm">
          Send the request now to see what comes back. It doesn&apos;t run the
          rest of the flow and doesn&apos;t cost credits.
        </div>
      </div>

      {placeholders.length > 0 && (
        <div className="space-y-2">
          <div className="text-sm text-muted-foreground">
            Nothing triggered the flow, so enter a value for each placeholder.
            Secrets are filled in for you.
          </div>
          {placeholders.map((p) => (
            <div key={p} className="grid grid-cols-2 gap-2 items-center">
              <code className="text-sm truncate" title={p}>
                {`{{${p}}}`}
              </code>
              <Input
                type="text"
                value={values[p] || ""}
                onChange={(e) =>
                  setValues((v) => ({ ...v, [p]: e.target.value }))
                }
              />
            </div>
          ))}
        </div>
      )}

      <Button
        variant="secondary"
        onClick={send}
        disabled={mutation.isPending || !data.url}
      >
        {mutation.isPending ? (
          <LoaderCircleIcon className="h-4 w-4 mr-2 animate-spin" />
        ) : (
          <SendIcon className="h-4 w-4 mr-2" />
        )}
        Send test request
      </Button>

      {result && <HttpRequestTestResult result={result} />}
    </div>
  );
}

function HttpRequestTestResult({
  result,
}: {
  result: HTTPRequestTestResponse;
}) {
  const response = result.response;

  if (!response) {
    return (
      <div className="text-sm text-destructive bg-destructive/10 rounded p-3 break-words">
        The request couldn&apos;t be sent: {result.error}
      </div>
    );
  }

  const ok = response.status_code >= 200 && response.status_code < 300;

  return (
    <div className="space-y-2 min-w-0">
      <div className="flex items-baseline gap-2 text-sm">
        <span
          className={cn(
            "font-medium whitespace-nowrap",
            ok ? "text-green-600 dark:text-green-400" : "text-destructive"
          )}
        >
          {response.status || response.status_code}
        </span>
        <span className="text-muted-foreground min-w-0 break-all">
          {result.method} {result.url} in {result.duration_ms} ms
        </span>
      </div>

      <details className="text-sm">
        <summary className="cursor-pointer text-muted-foreground">
          Response headers
        </summary>
        <div className="mt-2 bg-muted rounded p-3 font-mono text-xs space-y-1 overflow-x-auto">
          {Object.entries(response.headers)
            .sort(([a], [b]) => a.localeCompare(b))
            .map(([name, value]) => (
              <div key={name} className="break-all">
                <span className="text-muted-foreground">{name}:</span> {value}
              </div>
            ))}
        </div>
      </details>

      {response.body ? (
        <>
          <pre className="bg-muted rounded p-3 text-xs overflow-auto max-h-96 whitespace-pre-wrap break-all">
            {formatResponseBody(response.body)}
          </pre>
          {response.body_truncated && (
            <div className="text-sm text-muted-foreground">
              Only the first 64 KB of the response are shown.
            </div>
          )}
        </>
      ) : (
        <div className="text-sm text-muted-foreground">
          The response has no body.
        </div>
      )}
    </div>
  );
}
