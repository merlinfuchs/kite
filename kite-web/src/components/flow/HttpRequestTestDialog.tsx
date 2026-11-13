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
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "../ui/dialog";

// Opened from the dialog of the request, where it stays mounted, so the values
// and the result are kept while the request is being changed.
export default function HttpRequestTestDialog({
  data,
}: {
  data: HTTPRequestData;
}) {
  const appId = useAppId();
  const mutation = useHTTPRequestTestMutation(appId);

  const placeholders = httpRequestPlaceholders(data);
  const [values, setValues] = useState<Record<string, string>>({});
  // The result is only shown while the request is the one that was sent.
  const [result, setResult] = useState<{
    request: string;
    res: HTTPRequestTestResponse;
  } | null>(null);
  const request = JSON.stringify(data);

  function send() {
    mutation.mutate(
      {
        data,
        // Empty ones are evaluated like in a run, which works for {{now()}}.
        values: Object.fromEntries(
          placeholders.filter((p) => values[p]).map((p) => [p, values[p]])
        ),
      },
      {
        onSuccess: (res) => {
          if (res.success) {
            setResult({ request, res: res.data });
          } else {
            toast.error(`Failed to send test request: ${res.error.message}`);
          }
        },
      }
    );
  }

  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button variant="secondary" disabled={!data.url}>
          <SendIcon className="h-4 w-4 mr-2" />
          Send test request
        </Button>
      </DialogTrigger>
      <DialogContent className="overflow-y-auto max-h-[90dvh] max-w-3xl">
        <DialogHeader>
          <DialogTitle>Send test request</DialogTitle>
          <DialogDescription>
            See what comes back before you save the flow. It doesn&apos;t run
            the rest of the flow and doesn&apos;t cost credits.
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4 min-w-0">
          {placeholders.length > 0 && (
            <div className="space-y-2">
              <div className="text-sm text-muted-foreground">
                Nothing triggered the flow, so enter a value for each
                placeholder. Secrets are filled in for you.
              </div>
              {placeholders.map((p) => (
                <div
                  key={p}
                  className="grid sm:grid-cols-2 gap-1 sm:gap-2 items-center"
                >
                  <code className="text-sm break-all sm:truncate" title={p}>
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

          <Button onClick={send} disabled={mutation.isPending}>
            {mutation.isPending ? (
              <LoaderCircleIcon className="h-4 w-4 mr-2 animate-spin" />
            ) : (
              <SendIcon className="h-4 w-4 mr-2" />
            )}
            Send
          </Button>

          {result?.request === request && (
            <HttpRequestTestResult result={result.res} />
          )}
        </div>
      </DialogContent>
    </Dialog>
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
          <pre className="bg-muted rounded p-3 text-xs overflow-auto max-h-[50dvh] whitespace-pre-wrap break-all">
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
