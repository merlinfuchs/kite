import { ComponentProps } from "react";

// Only real links, so a javascript: URL typed into the editor can't run on click.
const SAFE_HREF_RE = /^(https?|discord):\/\//i;

/** An external link in the preview that opens in a new tab. */
export default function MessagePreviewLink({
  href,
  ...props
}: ComponentProps<"a">) {
  return (
    <a
      {...props}
      href={href && SAFE_HREF_RE.test(href) ? href : undefined}
      target="_blank"
      rel="noreferrer"
    />
  );
}
