import { useState } from "react";
import { useRouter } from "next/router";
import { StarIcon } from "lucide-react";
import { toast } from "sonner";
import { Button } from "../ui/button";
import { minShownStars, Template } from "@/lib/flow/templates";
import { useBundleStarMutation } from "@/lib/api/mutations";
import { useUserQuery } from "@/lib/api/queries";
import { loginUrlTo } from "@/lib/api/client";
import { useStarredTemplateIds } from "@/lib/hooks/templates";
import { cn, formatNumber } from "@/lib/utils";

// TemplateStarButton stars a template of the marketplace, which anyone
// logged in can do, except for their own ones.
export default function TemplateStarButton({
  template,
}: {
  template: Template;
}) {
  const router = useRouter();
  const { data: user } = useUserQuery();
  const starredIds = useStarredTemplateIds();
  const starMutation = useBundleStarMutation();
  // Pages rendered on the server have the count from before.
  const [count, setCount] = useState<number>();

  const own =
    template.author.kind === "you" ||
    (template.author.kind === "user" &&
      user?.success &&
      user.data.discord_id === template.author.discordId);
  if (!template.listed || own) return null;

  const starred = !!starredIds?.includes(template.id);
  const stars = count ?? template.stars ?? 0;
  const content = (
    <>
      <StarIcon
        className={cn("h-4 w-4", starred && "fill-yellow-400 text-yellow-400")}
      />
      {starred ? "Starred" : "Star"}
      {stars >= minShownStars && (
        <span className="text-muted-foreground">{formatNumber(stars)}</span>
      )}
    </>
  );

  if (user && !user.success) {
    return (
      <Button variant="outline" size="sm" className="gap-1.5" asChild>
        <a href={loginUrlTo(router.asPath)}>{content}</a>
      </Button>
    );
  }

  async function toggle() {
    const res = await starMutation.mutateAsync({
      bundleId: template.id,
      starred: !starred,
    });
    if (!res.success) {
      toast.error(
        `Failed to star ${template.name}: ${res.error.message} (${res.error.code})`
      );
      return;
    }
    setCount(res.data.star_count);
  }

  return (
    <Button
      variant="outline"
      size="sm"
      className="gap-1.5"
      onClick={toggle}
      disabled={!starredIds || starMutation.isPending}
    >
      {content}
    </Button>
  );
}
