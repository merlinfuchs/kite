import AppLayout from "@/components/app/AppLayout";
import AppStateEmojiList from "@/components/app/AppStateEmojiList";
import { Separator } from "@/components/ui/separator";
import { useApp } from "@/lib/hooks/api";
import Link from "next/link";

const breadcrumbs = [
  {
    label: "Emoji Explorer",
  },
];

export default function AppGuildsPage() {
  const app = useApp();

  const appEmojisUrl = `https://discord.com/developers/applications/${app?.discord_id}/emojis`;

  return (
    <AppLayout title="App Analytics" breadcrumbs={breadcrumbs}>
      <div>
        <h1 className="text-lg font-semibold md:text-2xl mb-1">
          Emoji Explorer
        </h1>
        <p className="text-muted-foreground text-sm">
          Explore the emojis of your app. Create new emojis in the{" "}
          <Link
            href={appEmojisUrl}
            target="_blank"
            className="text-primary hover:underline"
          >
            Discord Developer Portal
          </Link>
          .
        </p>
      </div>
      <Separator className="my-4" />
      <div className="space-y-5">
        <AppStateEmojiList />
      </div>
    </AppLayout>
  );
}
