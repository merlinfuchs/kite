import { useAssetQuery } from "@/lib/api/queries";
import { useResponseData } from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import { toast } from "sonner";
import { Card } from "../ui/card";
import { Input } from "../ui/input";
import { Button } from "../ui/button";
import { TrashIcon } from "lucide-react";
import { getApiUrl } from "@/lib/api/client";
import { useCallback, useMemo } from "react";
import { useCurrentMessage } from "@/lib/message/state";

export default function MessageAttachment({
  attachmentIndex,
  assetId,
}: {
  attachmentIndex: number;
  assetId: string;
}) {
  const appID = useAppId();

  const deleteAttachment = useCurrentMessage((state) => state.deleteAttachment);

  const asset = useResponseData(useAssetQuery(appID, assetId), (res) => {
    if (!res.success) {
      toast.error(
        `Failed to load asset: ${res?.error.message} (${res?.error.code})`
      );
    }
  });

  const downloadUrl = useMemo(
    () => getApiUrl(`/v1/apps/${appID}/assets/${assetId}/download`),
    [appID, assetId]
  );

  const remove = useCallback(() => {
    deleteAttachment(attachmentIndex);
  }, [deleteAttachment]);

  return (
    <Card className="p-2">
      <div className="pb-2 flex gap-2">
        <Input
          type="text"
          className="text-sm py-0 px-1.5 h-8 rounded-sm w-full text-muted-foreground"
          value={asset?.name || "unknown.oops"}
          readOnly
        />
        <Button variant="outline" size="icon" className="flex-none h-8 w-8">
          <TrashIcon className="h-4 w-4" onClick={remove} />
        </Button>
      </div>
      <img src={downloadUrl} alt="" className="bg-blue-500 w-64 rounded-sm" />
    </Card>
  );
}
