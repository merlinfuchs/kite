import { PlayIcon } from "lucide-react";
import { useState } from "react";

// Only loads the video, and with it a connection to YouTube, once it's
// clicked, so visiting the page doesn't send the visitor's IP to Google.
export default function YouTubeEmbed({
  videoId,
  title,
}: {
  videoId: string;
  title: string;
}) {
  const [loaded, setLoaded] = useState(false);

  if (loaded) {
    return (
      <iframe
        className="w-full h-full"
        src={`https://www.youtube-nocookie.com/embed/${videoId}?autoplay=1`}
        title={title}
        allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share"
        referrerPolicy="strict-origin-when-cross-origin"
        allowFullScreen
      />
    );
  }

  return (
    <button
      type="button"
      onClick={() => setLoaded(true)}
      className="w-full h-full flex flex-col items-center justify-center gap-3 bg-muted/50 hover:bg-muted transition-colors p-6 text-center"
    >
      <span className="flex size-14 items-center justify-center rounded-full bg-primary text-primary-foreground">
        <PlayIcon className="size-6 ml-0.5" />
      </span>
      <span className="font-medium">{title}</span>
      <span className="text-sm text-muted-foreground max-w-md">
        Playing the video loads it from YouTube, which receives your IP address.
      </span>
    </button>
  );
}
