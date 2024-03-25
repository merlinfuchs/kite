import {
  useKVStorageKeysQuery,
  useKVStorageNamespacesQuery,
} from "@/lib/api/queries";
import { useState } from "react";
import AppIllustrationPlaceholder from "./AppIllustrationPlaceholder";
import { Card, CardContent } from "../ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../ui/select";
import AppGuildPageEmpty from "./AppGuildPageEmpty";

export default function KVStorageBrowser({ guildId }: { guildId: string }) {
  const [namespace, setNamespace] = useState<string>("default");

  const { data: namespaceResp } = useKVStorageNamespacesQuery(guildId);
  const { data: keyResp } = useKVStorageKeysQuery(guildId, namespace);

  const namespaces = namespaceResp?.success ? namespaceResp.data : [];
  const keys = keyResp?.success ? keyResp.data : [];

  if (namespaces.length === 0) {
    return (
      <AppGuildPageEmpty
        title="There are no values yet"
        description="You can store values in the KV from your plugin code."
      />
    );
  }

  return (
    <div>
      <div className="flex justify-between mb-3">
        <div>
          <Select value={namespace} onValueChange={(v) => setNamespace(v)}>
            <SelectTrigger className="min-w-48">
              <SelectValue placeholder="Namespace"></SelectValue>
            </SelectTrigger>
            <SelectContent>
              {namespaces.map((ns) => (
                <SelectItem key={ns.namespace} value={ns.namespace}>
                  {ns.namespace}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </div>
      <Card>
        <CardContent>
          {keys.length === 0 ? (
            <div className="text-gray-400">
              There isn't any data here yet ...
            </div>
          ) : (
            <table className="min-w-full divide-y divide-gray-500">
              <thead className="text-gray-100 font-medium text-left">
                <tr>
                  <th className="p-2">Key</th>
                  <th className="p-2">Value</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-600">
                {keys.map((key) => (
                  <tr key={key.key} className="divide-x divide-gray-600">
                    <td className="text-gray-100 p-2">{key.key}</td>
                    <td className="text-gray-300 p-2">
                      {JSON.stringify(key.value)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
