import { useAssetQuery } from "@/lib/api/queries";
import { useResponseData } from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import { toast } from "sonner";
import { Card } from "../ui/card";
import { Input } from "../ui/input";
import { Button } from "../ui/button";
import { PaperclipIcon, TrashIcon } from "lucide-react";
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
  const deleteAttachment = useCurrentMessage((state) => state.deleteAttachment);

  const appId = useAppId();

  const asset = useResponseData(useAssetQuery(appId, assetId), (res) => {
    if (!res.success) {
      toast.error(
        `Failed to load asset: ${res?.error.message} (${res?.error.code})`
      );
    }
  });

  const remove = useCallback(() => {
    console.log(attachmentIndex);
    deleteAttachment(attachmentIndex);
  }, [deleteAttachment, attachmentIndex]);

  const isImage = useMemo(
    () => asset?.content_type?.startsWith("image/"),
    [asset]
  );

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
      {isImage ? (
        <img
          src={asset?.url}
          alt=""
          className="w-64 rounded-sm bg-muted h-32 object-cover"
        />
      ) : (
        <div className="w-64 rounded-sm bg-muted h-32 flex items-center justify-center">
          <PaperclipIcon className="h-14 w-14 text-muted-foreground" />
        </div>
      )}
    </Card>
  );
}
