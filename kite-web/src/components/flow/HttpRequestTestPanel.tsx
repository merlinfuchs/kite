import {
  formatHttpTestBody,
  formatHttpTestSize,
  getHttpTestValueKeys,
  getSessionHttpTestValues,
  setSessionHttpTestValues,
} from "@/lib/flow/httpTest";
import { useFlowHTTPRequestTestMutation } from "@/lib/api/mutations";
import { useAppId } from "@/lib/hooks/params";
import { HTTPRequestData } from "@/lib/types/flow.gen";
import { FlowHTTPRequestTestResponse } from "@/lib/types/wire.gen";
import { cn } from "@/lib/utils";
import {
  CheckIcon,
  CircleAlertIcon,
  CopyIcon,
  LoaderCircleIcon,
  PlayIcon,
  PlusIcon,
  XIcon,
} from "lucide-react";
import { useCallback, useEffect, useMemo, useState } from "react";
import { Button } from "../ui/button";
import { Input } from "../ui/input";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "../ui/tabs";

const errorStageLabels: Record<string, string> = {
  build: "Couldn't build the request",
  send: "Couldn't send the request",
  status: "Failed on error status",
  transform: "Transform failed",
};

export default function HttpRequestTestPanel({
  nodeId,
  request,
}: {
  nodeId: string;
  request?: HTTPRequestData;
}) {
  const appId = useAppId();
  const testMutation = useFlowHTTPRequestTestMutation(appId);

  const [values, setValuesState] = useState<Record<string, string>>(() =>
    getSessionHttpTestValues(nodeId)
  );
  const [extraKeys, setExtraKeys] = useState<string[]>(() =>
    Object.keys(getSessionHttpTestValues(nodeId))
  );
  const [result, setResult] = useState<FlowHTTPRequestTestResponse | null>(
    null
  );
  const [requestError, setRequestError] = useState<string | null>(null);

  const setValues = useCallback(
    (newValues: Record<string, string>) => {
      setValuesState(newValues);
      setSessionHttpTestValues(nodeId, newValues);
    },
    [nodeId]
  );

  const detectedKeys = useMemo(() => getHttpTestValueKeys(request), [request]);
  // Keys typed in by hand, with their index in extraKeys so rows with the
  // same (e.g. still empty) key can be told apart.
  const manualRows = extraKeys
    .map((key, index) => ({ key, index }))
    .filter(({ key }) => !detectedKeys.includes(key));
  const manualKeys = manualRows.map((r) => r.key);
  const [runId, setRunId] = useState(0);

  // A result for an older version of the request is misleading.
  const requestKey = JSON.stringify(request ?? null);
  const [resultFor, setResultFor] = useState<string | null>(null);
  const isStale = !!result && resultFor !== requestKey;

  const runTest = useCallback(() => {
    const keys = [...detectedKeys, ...manualKeys];
    const sent: Record<string, string> = {};
    for (const key of keys) {
      const k = key.trim();
      if (k && values[key] !== undefined && values[key] !== "") {
        sent[k] = values[key];
      }
    }

    setRequestError(null);
    testMutation.mutate(
      { data: request ?? {}, values: sent },
      {
        onSuccess(res) {
          if (res.success) {
            setResult(res.data);
            setResultFor(requestKey);
            setRunId((id) => id + 1);
          } else {
            setResult(null);
            setRequestError(res.error.message);
          }
        },
        onError(err) {
          setResult(null);
          setRequestError(String(err));
        },
      }
    );
  }, [detectedKeys, manualKeys, values, request, requestKey, testMutation]);

  const withoutKey = (key: string) => {
    const rest = { ...values };
    // Another row may still use the key.
    if (extraKeys.filter((k) => k === key).length <= 1) delete rest[key];
    return rest;
  };

  const renameManualKey = (index: number, newKey: string) => {
    const oldKey = extraKeys[index];
    setExtraKeys(extraKeys.map((k, i) => (i === index ? newKey : k)));
    const value = values[oldKey];
    const rest = withoutKey(oldKey);
    setValues(value !== undefined ? { ...rest, [newKey]: value } : rest);
  };

  const removeManualKey = (index: number) => {
    const key = extraKeys[index];
    setExtraKeys(extraKeys.filter((_, i) => i !== index));
    setValues(withoutKey(key));
  };

  return (
    <div className="rounded-md border">
      <div className="flex items-center justify-between gap-3 px-3 py-2">
        <div className="min-w-0">
          <div className="text-sm font-medium text-foreground">
            Test request
          </div>
          <div className="text-muted-foreground text-xs">
            Sends the request now, without deploying. This is a real request.
          </div>
        </div>
        <Button
          size="sm"
          onClick={runTest}
          disabled={!request?.url || testMutation.isPending}
          className="flex-none"
        >
          {testMutation.isPending ? (
            <LoaderCircleIcon className="h-4 w-4 mr-1.5 animate-spin" />
          ) : (
            <PlayIcon className="h-4 w-4 mr-1.5" />
          )}
          Send
        </Button>
      </div>

      <div className="border-t px-3 py-3 space-y-2">
        <div className="text-xs font-medium text-muted-foreground">
          Test values
        </div>
        {detectedKeys.length === 0 && manualKeys.length === 0 && (
          <div className="text-muted-foreground text-sm">
            This request doesn&apos;t use any placeholders.
          </div>
        )}

        {detectedKeys.map((key) => (
          <div className="flex gap-2 items-center" key={key}>
            <code
              className="w-2/5 flex-none truncate rounded bg-muted px-2 py-2 text-xs"
              title={key}
            >
              {key}
            </code>
            <Input
              type="text"
              className="flex-auto min-w-0"
              placeholder="Value to use while testing"
              value={values[key] || ""}
              onChange={(e) => setValues({ ...values, [key]: e.target.value })}
            />
          </div>
        ))}

        {manualRows.map(({ key, index }) => (
          <div className="flex gap-2 items-center" key={index}>
            <Input
              type="text"
              className="w-2/5 flex-none font-mono text-xs"
              placeholder="user.id"
              value={key}
              onChange={(e) => renameManualKey(index, e.target.value)}
            />
            <Input
              type="text"
              className="flex-auto min-w-0"
              placeholder="Value"
              value={values[key] || ""}
              onChange={(e) => setValues({ ...values, [key]: e.target.value })}
            />
            <Button
              variant="ghost"
              size="icon"
              className="flex-none h-9 w-9 text-muted-foreground"
              title="Remove"
              onClick={() => removeManualKey(index)}
            >
              <XIcon className="h-4 w-4" />
            </Button>
          </div>
        ))}

        <Button
          variant="ghost"
          size="sm"
          className="-ml-2 text-muted-foreground"
          onClick={() => setExtraKeys([...extraKeys, ""])}
        >
          <PlusIcon className="h-4 w-4 mr-1.5" />
          Add value
        </Button>
      </div>

      {requestError && (
        <div className="border-t px-3 py-3">
          <TestError message={requestError} />
        </div>
      )}

      {result && (
        <div
          className={cn(
            "border-t px-3 py-3 space-y-3",
            isStale && "opacity-60"
          )}
        >
          {isStale && (
            <div className="text-muted-foreground text-xs">
              The request changed since this test. Send it again to update.
            </div>
          )}
          <TestResult key={runId} result={result} />
        </div>
      )}
    </div>
  );
}

function TestResult({ result }: { result: FlowHTTPRequestTestResponse }) {
  const response = result.response;
  const hasResult = result.result !== null && result.result !== undefined;

  return (
    <>
      <div className="flex flex-wrap items-center gap-x-3 gap-y-1 text-sm">
        {response && (
          <StatusBadge code={response.status_code} status={response.status} />
        )}
        {response && (
          <span className="text-muted-foreground tabular-nums">
            {result.duration_ms} ms · {formatHttpTestSize(response.body_size)}
          </span>
        )}
      </div>

      {result.error && (
        <TestError
          title={errorStageLabels[result.error_stage]}
          message={result.error}
        />
      )}

      {(response || result.url) && (
        <Tabs
          defaultValue={hasResult ? "result" : response ? "body" : "request"}
        >
          <TabsList className="w-full justify-start">
            {hasResult && <TabsTrigger value="result">Result</TabsTrigger>}
            {response && <TabsTrigger value="body">Body</TabsTrigger>}
            {response && <TabsTrigger value="headers">Headers</TabsTrigger>}
            {result.url && (
              <TabsTrigger value="request">Sent request</TabsTrigger>
            )}
          </TabsList>

          {hasResult && (
            <TabsContent value="result" className="mt-3">
              <CodeBlock text={JSON.stringify(result.result, null, 2)} />
              <div className="text-muted-foreground text-xs pt-2">
                What later blocks get from this block, after the transform.
              </div>
            </TabsContent>
          )}

          {response && (
            <TabsContent value="body" className="mt-3">
              {response.body_binary ? (
                <div className="text-muted-foreground text-sm">
                  The response is binary and can&apos;t be shown.
                </div>
              ) : response.body ? (
                <>
                  <CodeBlock text={formatHttpTestBody(response.body).text} />
                  {response.body_truncated && (
                    <div className="text-muted-foreground text-xs pt-2">
                      Only the start of the body is shown. The flow still gets
                      all of it.
                    </div>
                  )}
                </>
              ) : (
                <div className="text-muted-foreground text-sm">
                  The response has no body.
                </div>
              )}
            </TabsContent>
          )}

          {response && (
            <TabsContent value="headers" className="mt-3">
              <HeaderTable headers={response.headers} />
            </TabsContent>
          )}

          {result.url && (
            <TabsContent value="request" className="mt-3 space-y-3">
              <CodeBlock text={`${result.method} ${result.url}`} />
              <HeaderTable headers={result.request_headers} />
              {result.request_body && (
                <CodeBlock
                  text={formatHttpTestBody(result.request_body).text}
                />
              )}
            </TabsContent>
          )}
        </Tabs>
      )}
    </>
  );
}

function StatusBadge({ code, status }: { code: number; status: string }) {
  const color =
    code >= 500
      ? "bg-red-500/15 text-red-600 dark:text-red-400"
      : code >= 400
      ? "bg-orange-500/15 text-orange-600 dark:text-orange-400"
      : code >= 300
      ? "bg-blue-500/15 text-blue-600 dark:text-blue-400"
      : "bg-green-500/15 text-green-600 dark:text-green-400";

  return (
    <span
      className={cn(
        "rounded px-2 py-0.5 font-mono text-xs font-semibold",
        color
      )}
    >
      {status || code}
    </span>
  );
}

function TestError({ title, message }: { title?: string; message: string }) {
  return (
    <div className="text-red-600 dark:text-red-400 text-sm flex items-start space-x-1.5">
      <CircleAlertIcon className="h-4 w-4 flex-none mt-0.5" />
      <div className="min-w-0 break-words">
        {title && <div className="font-medium">{title}</div>}
        <div className={cn(title && "text-xs")}>{message}</div>
      </div>
    </div>
  );
}

function HeaderTable({ headers }: { headers?: Record<string, string> }) {
  const entries = Object.entries(headers || {}).sort(([a], [b]) =>
    a.localeCompare(b)
  );
  if (entries.length === 0) {
    return <div className="text-muted-foreground text-sm">No headers.</div>;
  }

  return (
    <div className="overflow-x-auto rounded-md border">
      <table className="w-full text-xs">
        <tbody>
          {entries.map(([key, value]) => (
            <tr key={key} className="border-b last:border-0">
              <td className="px-2 py-1.5 font-mono font-medium align-top whitespace-nowrap">
                {key}
              </td>
              <td className="px-2 py-1.5 font-mono break-all text-muted-foreground">
                {value}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function CodeBlock({ text }: { text: string }) {
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    if (!copied) return;
    const timeout = setTimeout(() => setCopied(false), 1500);
    return () => clearTimeout(timeout);
  }, [copied]);

  return (
    <div className="relative">
      <pre className="max-h-72 overflow-auto rounded-md bg-muted p-3 pr-10 text-xs font-mono whitespace-pre-wrap break-all">
        {text}
      </pre>
      <Button
        variant="ghost"
        size="icon"
        className="absolute top-1 right-1 h-7 w-7 text-muted-foreground"
        title="Copy"
        onClick={() => {
          navigator.clipboard?.writeText(text).then(() => setCopied(true));
        }}
      >
        {copied ? (
          <CheckIcon className="h-3.5 w-3.5" />
        ) : (
          <CopyIcon className="h-3.5 w-3.5" />
        )}
      </Button>
    </div>
  );
}
