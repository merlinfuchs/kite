import { Button } from "@/components/ui/button";
import { useAssetCreateMutation } from "@/lib/api/mutations";
import { useAppId } from "@/lib/hooks/params";
import { UploadIcon } from "lucide-react";
import { ChangeEvent, useCallback, useRef } from "react";
import { toast } from "sonner";

export default function ImageUploadButton({
  onImageUploaded,
}: {
  onImageUploaded: (url: string) => void;
}) {
  const createMutation = useAssetCreateMutation(useAppId());
  const inputRef = useRef<HTMLInputElement>(null);

  const onFileUpload = useCallback(
    (e: ChangeEvent<HTMLInputElement>) => {
      const file = e.target.files?.[0];
      if (!file) return;

      const toastId = toast.loading("Uploading attachment...");

      createMutation.mutateAsync(file, {
        onSuccess: (res) => {
          if (res.success) {
            onImageUploaded(res.data.url);
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
    [createMutation, onImageUploaded]
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
        accept="image/*"
      />

      <UploadIcon className="h-5 w-5" />
    </Button>
  );
}
