import { useEffect, useState } from "react";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "../ui/card";
import { Textarea } from "../ui/textarea";
import { Button } from "../ui/button";
import { Skeleton } from "../ui/skeleton";
import { useCreditLimitSettings } from "@/lib/hooks/api";
import { useCreditLimitSettingsUpdateMutation } from "@/lib/api/mutations";
import { useAppId } from "@/lib/hooks/params";
import { toast } from "sonner";
import CreditLimitMessagePlaceholders from "./CreditLimitMessagePlaceholders";

// The message users see for limits that don't have one of their own.
export default function CreditLimitSettingsCard() {
  const settings = useCreditLimitSettings();
  const mutation = useCreditLimitSettingsUpdateMutation(useAppId());

  const [message, setMessage] = useState("");

  useEffect(() => {
    if (settings) setMessage(settings.message ?? "");
  }, [settings]);

  const saved = settings?.message ?? "";
  const changed = message !== saved;

  function save(value: string) {
    mutation.mutate(
      { message: value.trim() ? value : null },
      {
        onSuccess(res) {
          if (res.success) {
            toast.success("Message saved!");
            setMessage(res.data.message ?? "");
          } else {
            toast.error(
              `Failed to save message: ${res.error.message} (${res.error.code})`
            );
          }
        },
      }
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Default message</CardTitle>
        <CardDescription>
          Shown when a limit without a message of its own is reached. Leave
          empty to use Kite&apos;s message.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-3">
        {!settings ? (
          <Skeleton className="h-20" />
        ) : (
          <Textarea
            rows={3}
            maxLength={2000}
            placeholder="You have reached your usage limit for today. Try again later."
            value={message}
            onChange={(e) => setMessage(e.target.value)}
          />
        )}
        <CreditLimitMessagePlaceholders />
      </CardContent>
      <CardFooter className="flex space-x-3">
        <Button
          size="sm"
          disabled={!changed || mutation.isPending}
          onClick={() => save(message)}
        >
          Save message
        </Button>
        {saved && (
          <Button
            size="sm"
            variant="ghost"
            disabled={mutation.isPending}
            onClick={() => save("")}
          >
            Reset to default
          </Button>
        )}
      </CardFooter>
    </Card>
  );
}
