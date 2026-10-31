import { Button } from "@/components/ui/button";
import {
  AlertDialog,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { useUserQuery } from "@/lib/api/queries";
import env from "@/lib/env/client";
import { useLogout } from "@/lib/hooks/auth";

// KiteBanDialog tells users that Kite banned them and why, as every request
// they make is refused. It can't be closed, only left by logging out.
export default function KiteBanDialog() {
  const { data } = useUserQuery();
  const logout = useLogout();
  const error = data && !data.success ? data.error : undefined;

  return (
    <AlertDialog open={error?.code === "kite_banned"}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>You can&apos;t use Kite anymore</AlertDialogTitle>
          <AlertDialogDescription className="whitespace-pre-wrap">
            {error?.message}
          </AlertDialogDescription>
          <AlertDialogDescription>
            Your bots were stopped. To appeal, ask on our Discord server.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <Button variant="outline" onClick={logout}>
            Log out
          </Button>
          <Button asChild>
            <a href={env.NEXT_PUBLIC_DISCORD_LINK} target="_blank">
              Discord server
            </a>
          </Button>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
