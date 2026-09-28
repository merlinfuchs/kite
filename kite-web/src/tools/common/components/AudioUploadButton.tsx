import { Button } from "@/components/ui/button";
import { useAssetCreateMutation } from "@/lib/api/mutations";
import { useAppId } from "@/lib/hooks/params";
import { UploadIcon } from "lucide-react";
import { ChangeEvent, useCallback, useRef } from "react";
import { toast } from "sonner";

// Keep this well under the account's general asset size limit: the whole
// file is downloaded and Opus-encoded in memory when a flow plays it back.
const MAX_AUDIO_FILE_SIZE_BYTES = 8 * 1024 * 1024; // 8 MB

export default function AudioUploadButton({
  onAudioUploaded,
}: {
  onAudioUploaded: (url: string) => void;
}) {
  const createMutation = useAssetCreateMutation(useAppId());
  const inputRef = useRef<HTMLInputElement>(null);

  const onFileUpload = useCallback(
    (e: ChangeEvent<HTMLInputElement>) => {
      const file = e.target.files?.[0];
      if (!file) return;

      if (file.size > MAX_AUDIO_FILE_SIZE_BYTES) {
        toast.error(
          `Audio file is too large (max ${
            MAX_AUDIO_FILE_SIZE_BYTES / (1024 * 1024)
          } MB)`
        );
        e.target.value = "";
        return;
      }

      const toastId = toast.loading("Uploading audio file...");

      createMutation.mutateAsync(file, {
        onSuccess: (res) => {
          if (res.success) {
            onAudioUploaded(res.data.url);
          } else {
            toast.error(
              `Failed to upload asset: ${res.error.message} (${res.error.code})`
            );
          }
        },
        onSettled: () => {
          toast.dismiss(toastId);
          e.target.value = "";
        },
      });
    },
    [createMutation, onAudioUploaded]
  );

  return (
    <Button
      size="icon"
      variant="outline"
      onClick={() => inputRef.current?.click()}
      disabled={!!inputRef.current?.value}
    >
      <input
        type="file"
        className="hidden"
        ref={inputRef}
        onChange={onFileUpload}
        accept="audio/mpeg,.mp3"
      />

      <UploadIcon className="h-5 w-5" />
    </Button>
  );
}
