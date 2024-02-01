import { useQuickAccessItemListQuery } from "@/lib/api/queries";
import {
  CodeBracketSquareIcon,
  DocumentArrowUpIcon,
} from "@heroicons/react/24/outline";
import clsx from "clsx";
import Link from "next/link";
import { useRouter } from "next/router";
import { useMemo } from "react";

export default function AppSidebarQuickAccess({
  guildId,
}: {
  guildId: string;
}) {
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
  }, [resp]);

  if (items.length === 0) return null;

  return (
    <li>
      <div className="text-xs font-semibold leading-6 text-gray-400">
        Quick access
      </div>
      <ul role="list" className="-mx-2 mt-2 space-y-1">
        {items.map((item) => (
          <li key={item.id}>
            <Link
              href={item.href}
              className={clsx(
                item.current
                  ? "bg-dark-3 text-white"
                  : "text-gray-400 hover:text-white hover:bg-dark-3",
                "group flex gap-x-3 rounded-md p-2 text-sm leading-6 font-semibold"
              )}
            >
              <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-lg border border-gray-700 bg-gray-800 text-[0.625rem] font-medium text-gray-400 group-hover:text-white">
                <item.icon className="h-4 w-4 shrink-0" aria-hidden="true" />
              </span>
              <span className="truncate">{item.name}</span>
            </Link>
          </li>
        ))}
      </ul>
    </li>
  );
}
