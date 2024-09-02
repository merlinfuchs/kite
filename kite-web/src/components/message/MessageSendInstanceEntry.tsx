import { MessageInstance } from "@/lib/types/wire.gen";
import LoadingButton from "../common/LoadingButton";
import { useMessageInstanceUpdateMutation } from "@/lib/api/mutations";
import { useAppId, useMessageId } from "@/lib/hooks/params";
import { useCallback } from "react";
import { toast } from "sonner";

export default function MessageSendInstanceEntry({
  instance,
}: {
  instance: MessageInstance;
}) {
  const updateMutation = useMessageInstanceUpdateMutation(
    useAppId(),
    useMessageId(),
    instance.id
  );

  const updateInstance = useCallback(() => {
    if (updateMutation.isPending) return;

    updateMutation.mutate(
      {},
      {
        onSuccess(res) {
          if (res.success) {
            toast.success("Message instance updated!");
          } else {
            toast.error(
              `Failed to update message instance: ${res.error.message} (${res.error.code})`
            );
          }
        },
      }
    );
  }, [updateMutation]);

  return (
    <div className="grid grid-cols-2 sm:grid-cols-5 gap-x-3 gap-y-1 items-center">
      <div className="sm:col-span-2 space-y-1">
        <div className="text-sm truncate flex sm:flex-col gap-1">
          <div className="truncate">{instance.discord_guild_id}</div>
          <div className="text-muted-foreground truncate">
            #{instance.discord_channel_id}
          </div>
        </div>
        <div className="text-muted-foreground sm:hidden text-sm">
          {new Date().toLocaleString()}
        </div>
      </div>
      <div className="text-sm text-muted-foreground sm:col-span-2 hidden sm:block">
        {new Date().toLocaleString()}
      </div>
      <LoadingButton
        onClick={updateInstance}
        loading={updateMutation.isPending}
        size="sm"
        variant="secondary"
      >
        Update
      </LoadingButton>
    </div>
  );
}
