import { Button } from "@/components/ui/button";
import { CheckIcon, CopyIcon } from "lucide-react";
import { useState } from "react";

export default function CopyButton({ value }: { value: string }) {
  const [copied, setCopied] = useState(false);

  function copy() {
    navigator.clipboard.writeText(value);
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  }

  return (
    <Button variant="outline" size="sm" onClick={copy} className="flex-none">
      {copied ? (
        <CheckIcon className="w-4 h-4 mr-2" />
      ) : (
        <CopyIcon className="w-4 h-4 mr-2" />
      )}
      {copied ? "Copied" : "Copy"}
    </Button>
  );
}
