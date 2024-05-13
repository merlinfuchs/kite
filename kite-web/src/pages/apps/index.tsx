import AppCreateAppModal from "@/components/app/AppCreateAppModal";
import AppsLayout from "@/components/app/AppsLayout";
import { useAppsQuery } from "@/lib/api/queries";
import { userAvatarUrl } from "@/lib/discord/cdn";
import { nameAbbreviation } from "@/lib/discord/util";
import { PlusCircleIcon } from "@heroicons/react/24/outline";
import Link from "next/link";
import { useState } from "react";

export default function AppsPage() {
  const [createModalOpen, setCreateModalOpen] = useState(false);

  const { data: appsResp } = useAppsQuery();

  const apps = appsResp?.success ? appsResp.data : [];

  return (
    <AppsLayout>
      <AppCreateAppModal open={createModalOpen} setOpen={setCreateModalOpen} />
      <div className="max-w-5xl mx-auto pb-20 pt-10 lg:pt-20 px-5">
        <div className="text-4xl font-bold text-white mb-4">
          Welcome to Kite!
        </div>
        <div className="text-lg font-light text-gray-300 mb-10">
          To get started select your app from below that you want to use Kite
          with. If you haven't add an app yet, you can do that here.
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-5">
          {apps.map((app) => (
            <Link
              key={app.id}
              className="bg-dark-2 rounded-md px-3 py-3 flex items-center hover:scale-101 transition-transform"
              href={`/apps/${app.id}`}
            >
              <div className="bg-dark-1 h-14 w-14 rounded-full flex items-center justify-center flex-none mr-4">
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
              </div>
              <div className="truncate">
                <div className="text-lg font-medium text-gray-100 mb-2 truncate">
                  {app.user_name}
                </div>
                <div className="text-gray-400 text-sm truncate">{app.id}</div>
              </div>
            </Link>
          ))}
          <div
            className="rounded-md px-3 py-3 border-2 border-dashed border-dark-7 hover:bg-dark-4 flex items-center group transition-colors"
            role="button"
            onClick={() => setCreateModalOpen(true)}
          >
            <PlusCircleIcon className="h-14 w-14 text-gray-400 group-hover:text-gray-300 mr-3" />
            <div className="text-lg font-medium text-gray-100">Add App</div>
          </div>
        </div>
      </div>
    </AppsLayout>
  );
}
