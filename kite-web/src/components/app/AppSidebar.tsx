import clsx from "clsx";
import {
  ArrowLeftStartOnRectangleIcon,
  CircleStackIcon,
  CodeBracketSquareIcon,
  DocumentArrowUpIcon,
  HomeIcon,
  ShoppingCartIcon,
  XMarkIcon,
} from "@heroicons/react/24/outline";
import { Fragment, useMemo } from "react";
import { Dialog, Transition } from "@headlessui/react";
import { useRouter } from "next/router";
import Link from "next/link";
import AppSidebarQuickAccess from "./AppSidebarQuickAccess";
import AppSidebarAppSelect from "./AppSidebarAppSelect";
import { useUserQuery } from "@/lib/api/queries";
import { userAvatarUrl } from "@/lib/discord/cdn";
import { getApiUrl } from "@/lib/api/client";

interface Props {
  open: boolean;
  setOpen: (open: boolean) => void;
}

export default function AppSideBar({ open, setOpen }: Props) {
  const router = useRouter();
  const appId = router.query.gid as string;

  const { data: userResp } = useUserQuery();

  const user = userResp?.success
    ? userResp.data
    : {
        id: "0",
        username: "user",
        global_name: "User",
        discriminator: "0",
        avatar: null,
      };

  const navigation = useMemo(() => {
    return [
      {
        name: "Home",
        href: `/apps/${appId}`,
        icon: HomeIcon,
        current: router.pathname === "/apps/[aid]",
      },
      {
        name: "Deployments",
        href: `/apps/${appId}/deployments`,
        icon: DocumentArrowUpIcon,
        current: router.pathname.startsWith(`/apps/[aid]/deployments`),
      },
      {
        name: "Workspaces",
        href: `/apps/${appId}/workspaces`,
        icon: CodeBracketSquareIcon,
        current: router.pathname.startsWith(`/apps/[aid]/workspaces`),
      },
      {
        name: "KV Storage",
        href: `/apps/${appId}/kv-storage`,
        icon: CircleStackIcon,
        current: router.pathname.startsWith(`/apps/[aid]/kv-storage`),
      },
      {
        name: "Marketplace",
        href: `/apps/${appId}/marketplace`,
        icon: ShoppingCartIcon,
        current: router.pathname.startsWith(`/apps/[aid]/marketplace`),
      },
    ];
  }, [appId]);

  return (
    <div>
      <div className="lg:w-72 flex-none"></div>

      {/* Static sidebar for desktop */}
      <div className="hidden lg:fixed lg:inset-y-0 lg:z-40 lg:flex lg:w-72 lg:flex-col">
        {/* Sidebar component, swap this element with another sidebar if you like */}
        <div className="flex grow flex-col gap-y-5 overflow-y-auto bg-dark-2 px-4">
          <div className="mt-8 mb-5">
            <AppSidebarAppSelect appId={appId} />
          </div>
          <nav className="flex flex-1 flex-col px-2">
            <ul role="list" className="flex flex-1 flex-col gap-y-7">
              <li>
                <ul role="list" className="-mx-2 space-y-1">
                  {navigation.map((item) => (
                    <li key={item.name}>
                      <Link
                        href={item.href}
                        className={clsx(
                          item.current
                            ? "bg-dark-3 text-white"
                            : "text-gray-400 hover:text-white hover:bg-dark-3",
                          "group flex gap-x-3 rounded-md p-2 text-sm leading-6 font-semibold"
                        )}
                      >
                        <item.icon
                          className="h-6 w-6 shrink-0"
                          aria-hidden="true"
                        />
                        {item.name}
                      </Link>
                    </li>
                  ))}
                </ul>
              </li>
              <AppSidebarQuickAccess appId={appId} />
              <li className="-mx-6 mt-auto">
                <div className="flex items-center gap-x-4 px-6 py-3 text-sm font-semibold leading-6 text-white">
                  <img
                    className="h-8 w-8 rounded-full bg-dark-3"
                    src={userAvatarUrl(user)}
                    alt=""
                  />
                  <span className="sr-only">Your profile</span>
                  <span aria-hidden="true" className="truncate flex-auto">
                    {user?.global_name || user?.global_name || "User"}
                  </span>
                  <a
                    href={getApiUrl("/v1/auth/logout")}
                    aria-label="Logout"
                    className="hover:bg-dark-3 text-gray-300 hover:text-gray-300 rounded-full p-1"
                  >
                    <ArrowLeftStartOnRectangleIcon className="h-6 w-6" />
                  </a>
                </div>
              </li>
            </ul>
          </nav>
        </div>
      </div>

      {/* Dynamic sidebar for mobile */}
      <Transition.Root show={open} as={Fragment}>
        <Dialog as="div" className="relative z-40 lg:hidden" onClose={setOpen}>
          <Transition.Child
            as={Fragment}
            enter="transition-opacity ease-linear duration-300"
            enterFrom="opacity-0"
            enterTo="opacity-100"
            leave="transition-opacity ease-linear duration-300"
            leaveFrom="opacity-100"
            leaveTo="opacity-0"
          >
            <div className="fixed inset-0 bg-gray-900/80" />
          </Transition.Child>

          <div className="fixed inset-0 flex">
            <Transition.Child
              as={Fragment}
              enter="transition ease-in-out duration-300 transform"
              enterFrom="-translate-x-full"
              enterTo="translate-x-0"
              leave="transition ease-in-out duration-300 transform"
              leaveFrom="translate-x-0"
              leaveTo="-translate-x-full"
            >
              <Dialog.Panel className="relative mr-16 flex w-full max-w-xs flex-1">
                <Transition.Child
                  as={Fragment}
                  enter="ease-in-out duration-300"
                  enterFrom="opacity-0"
                  enterTo="opacity-100"
                  leave="ease-in-out duration-300"
                  leaveFrom="opacity-100"
                  leaveTo="opacity-0"
                >
                  <div className="absolute left-full top-0 flex w-16 justify-center pt-5">
                    <button
                      type="button"
                      className="-m-2.5 p-2.5"
                      onClick={() => setOpen(false)}
                    >
                      <span className="sr-only">Close sidebar</span>
                      <XMarkIcon
                        className="h-6 w-6 text-white"
                        aria-hidden="true"
                      />
                    </button>
                  </div>
                </Transition.Child>
                {/* Sidebar component, swap this element with another sidebar if you like */}
                <div className="flex grow flex-col gap-y-5 overflow-y-auto bg-dark-2 px-4 pb-2 ring-1 ring-white/10">
                  <div className="mt-8 mb-5">
                    <AppSidebarAppSelect appId={appId} />
                  </div>
                  <nav className="flex flex-1 flex-col px-2">
                    <ul role="list" className="flex flex-1 flex-col gap-y-7">
                      <li>
                        <ul role="list" className="-mx-2 space-y-1">
                          {navigation.map((item) => (
                            <li key={item.name}>
                              <Link
                                href={item.href}
                                className={clsx(
                                  item.current
                                    ? "bg-dark-3 text-white"
                                    : "text-gray-400 hover:text-white hover:bg-dark-3",
                                  "group flex gap-x-3 rounded-md p-2 text-sm leading-6 font-semibold"
                                )}
                              >
                                <item.icon
                                  className="h-6 w-6 shrink-0"
                                  aria-hidden="true"
                                />
                                {item.name}
                              </Link>
                            </li>
                          ))}
                        </ul>
                      </li>
                      <AppSidebarQuickAccess appId={appId} />
                    </ul>
                  </nav>
                </div>
              </Dialog.Panel>
            </Transition.Child>
          </div>
        </Dialog>
      </Transition.Root>
    </div>
  );
}
