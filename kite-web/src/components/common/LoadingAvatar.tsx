import { ReactNode, useState } from "react";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { cn } from "@/lib/utils";

// Avatar that shows a shimmer placeholder while its URL or image is loading,
// falling back to children when unavailable or when the image fails to load.
export default function LoadingAvatar({
  src,
  loading = false,
  className,
  fallbackClassName,
  children,
}: {
  src?: string;
  loading?: boolean;
  className?: string;
  fallbackClassName?: string;
  children: ReactNode;
}) {
  const [erroredSrc, setErroredSrc] = useState<string | null>(null);
  const imageErrored = !!src && erroredSrc === src;
  const showShimmer = loading || (!!src && !imageErrored);

  return (
    <Avatar className={cn("rounded-lg", className)}>
      {src && (
        <AvatarImage
          key={src}
          src={src}
          alt=""
          className="object-cover"
          onLoadingStatusChange={(status) => {
            if (status === "error") {
              setErroredSrc(src);
            } else if (status === "loaded" && erroredSrc === src) {
              setErroredSrc(null);
            }
          }}
        />
      )}
      <AvatarFallback
        className={
          showShimmer
            ? "relative overflow-hidden rounded-none bg-muted"
            : cn("rounded-lg", fallbackClassName)
        }
      >
        {showShimmer ? (
          <span
            aria-hidden="true"
            className="hidden motion-safe:block absolute inset-0 bg-gradient-to-r from-transparent via-foreground/10 dark:via-foreground/15 to-transparent motion-safe:animate-shimmer"
          />
        ) : (
          children
        )}
      </AvatarFallback>
    </Avatar>
  );
}
