import AppList from "@/components/app/AppList";
import { Separator } from "@/components/ui/separator";

export default function AppListPage() {
  return (
    <div className="flex flex-1 justify-center items-center min-h-[100dvh] w-full px-5 pt-10 pb-20">
      <div className="w-full max-w-lg">
        <div>
          <h1 className="text-lg font-semibold md:text-2xl mb-1">Your Apps</h1>
          <p className="text-muted-foreground text-sm">
            Apps are where you manage your plugins, integrations, and settings.
            Create an app or ask your team to invite you.
          </p>
        </div>
        <Separator className="my-4" />
        <AppList />
      </div>
    </div>
  );
}
