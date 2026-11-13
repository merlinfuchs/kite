import { useState } from "react";
import { CheckIcon, LinkIcon } from "lucide-react";
import { Button } from "../ui/button";
import env from "@/lib/env/client";
import { Template, templatePath } from "@/lib/flow/templates";

// TemplateCopyLinkButton copies the link of the public page of a template,
// also in an app, where the URL has the app's ID and only works for its
// members.
export default function TemplateCopyLinkButton({
  template,
}: {
  template: Template;
}) {
  const [copied, setCopied] = useState(false);

  if (!template.listed) return null;

  return (
    <Button
      variant="outline"
      size="sm"
      className="gap-1.5"
      onClick={() => {
        navigator.clipboard.writeText(
          `${env.NEXT_PUBLIC_SITE_URL}${templatePath(template.id)}`
        );
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
      }}
    >
      {copied ? (
        <CheckIcon className="h-4 w-4" />
      ) : (
        <LinkIcon className="h-4 w-4" />
      )}
      {copied ? "Copied" : "Copy link"}
    </Button>
  );
}
