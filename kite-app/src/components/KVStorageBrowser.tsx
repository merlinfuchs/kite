import {
  useKVStorageKeysQuery,
  useKVStorageNamespacesQuery,
} from "@/lib/api/queries";
import { useEffect, useState } from "react";

export default function KVStorageBrowser({ guildId }: { guildId: string }) {
  const [namespace, setNamespace] = useState<string>("default");

  const { data: namespaceResp } = useKVStorageNamespacesQuery(guildId);
  const { data: keyResp } = useKVStorageKeysQuery(guildId, namespace);

  const namespaces = namespaceResp?.success ? namespaceResp.data : [];
  const keys = keyResp?.success ? keyResp.data : [];

  return (
    <div>
      <div className="flex justify-between mb-3">
        <div>
          {namespaces.length !== 0 && (
            <select className="px-3 py-2 rounded bg-slate-900 min-w-64 text-gray-300">
              {namespaces.map((ns) => (
                <option key={ns.namespace} value={ns.namespace}>
                  {ns.namespace}
                </option>
              ))}
            </select>
          )}
        </div>
      </div>
      <div className="bg-slate-800 p-5 rounded-md">
        {keys.length === 0 ? (
          <div className="text-gray-400">There isn't any data here yet ...</div>
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
      </div>
    </div>
  );
}
