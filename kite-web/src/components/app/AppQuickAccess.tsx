import { useQuickAccessItemListQuery } from "@/lib/api/queries";
import {
  CodeBracketSquareIcon,
  DocumentArrowUpIcon,
} from "@heroicons/react/24/outline";
import clsx from "clsx";
import Link from "next/link";
import { useRouter } from "next/router";
import { useMemo } from "react";

export default function AppQuickAccess({ guildId }: { guildId: string }) {
  const router = useRouter();
  const { data: resp } = useQuickAccessItemListQuery(guildId);

  const items = useMemo(() => {
    if (!resp || !resp.success) return [];

    return resp.data.map((i) => {
      const url = `/app/guilds/${guildId}/${
        i.type === "DEPLOYMENT" ? "deployments" : "workspaces"
      }/${i.id}`;

      return {
        id: i.id,
        name: i.name,
        href: url,
        icon:
          i.type === "DEPLOYMENT" ? DocumentArrowUpIcon : CodeBracketSquareIcon,
        current: router.asPath === url,
      };
    });
  }, [resp, router.asPath]);

  if (items.length === 0) return null;

  return (
    <div className="lg:px-4 px-2">
      <div className="text-xs font-semibold leading-6 text-muted-foreground mb-2">
        Quick access
      </div>
      <nav className="grid items-start text-sm font-medium space-y-1">
        {items.map((item) => (
          <Link
            href={item.href}
            key={item.id}
            className={clsx(
              item.current ? "text-primary bg-muted" : "text-muted-foreground",
              "flex items-center gap-3 rounded-lg px-3 py-2 transition-all hover:text-primary"
            )}
          >
            <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-lg bg-muted">
              <item.icon className="h-4 w-4 shrink-0" aria-hidden="true" />
            </span>
            <span className="truncate">{item.name}</span>
          </Link>
        ))}
      </nav>
    </div>
  );
}
