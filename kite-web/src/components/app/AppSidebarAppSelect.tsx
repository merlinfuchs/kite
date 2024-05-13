import { useAppsQuery } from "@/lib/api/queries";
import { userAvatarUrl } from "@/lib/discord/cdn";
import { ChevronDownIcon } from "@heroicons/react/20/solid";
import { Menu } from "@headlessui/react";
import clsx from "clsx";
import Link from "next/link";
import { nameAbbreviation } from "@/lib/discord/util";

export default function AppSidebarAppSelect({ appId }: { appId: string }) {
  const { data: resp } = useAppsQuery();

  const apps = resp?.success ? resp.data : [];
  const app = apps.find((g) => g.id === appId);

  return (
    <Menu as="div" className="relative">
      <Menu.Button className="bg-dark-3 px-3 py-2 rounded cursor-pointer w-full hover:bg-dark-4">
        <div className="flex items-center select-none">
          <div className="bg-dark-1 h-10 w-10 rounded-full flex items-center justify-center flex-none mr-2">
            {app && (
              <img
                src={
                  userAvatarUrl({
                    id: app.user_id,
                    discriminator: app.user_discriminator,
                    avatar: app.user_avatar,
                  })!
                }
                alt={nameAbbreviation(app.user_name)}
                className="rounded-full h-full w-full"
              />
            )}
          </div>
          <div className="truncate text-gray-300 flex-auto text-left">
            {app?.user_name || "Unknown App"}
          </div>
          <div className="flex-none">
            <ChevronDownIcon
              className="h-5 w-5 text-gray-300"
              aria-hidden="true"
            />
          </div>
        </div>
      </Menu.Button>
      <Menu.Items className="absolute left-0 right-0 z-10 mt-2 w-full origin-top-right rounded bg-dark-3 shadow-lg focus:outline-none overflow-hidden">
        {apps.map((a) => (
          <Menu.Item key={a.id}>
            {({ active }) => (
              <Link
                className={clsx(
                  active && "bg-dark-4",
                  "px-4 py-2 text-sm flex items-center space-x-2 cursor-pointer"
                )}
                href={`/apps/${a.id}`}
              >
                <div className="bg-dark-1 h-10 w-10 rounded-full flex items-center justify-center flex-none mr-2">
                  <img
                    src={
                      userAvatarUrl({
                        id: a.user_id,
                        discriminator: a.user_discriminator,
                        avatar: a.user_avatar,
                      })!
                    }
                    alt={nameAbbreviation(a.user_name)}
                    className="rounded-full h-full w-full"
                  />
                </div>
                <div className="truncate text-gray-300 flex-auto">
                  {a?.user_name || "Unknown App"}
                </div>
              </Link>
            )}
          </Menu.Item>
        ))}
      </Menu.Items>
    </Menu>
  );
}
