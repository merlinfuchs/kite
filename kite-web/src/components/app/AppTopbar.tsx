import { useUserQuery } from "@/lib/api/queries";
import { userAvatarUrl } from "@/lib/discord/cdn";
import { Bars3Icon } from "@heroicons/react/24/outline";

interface Props {
  setSidebarOpen: (open: boolean) => void;
}

export default function AppTopbar({ setSidebarOpen }: Props) {
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

  return (
    <div className="sticky top-0 z-40 flex items-center gap-x-6 bg-dark-2 px-4 py-4 shadow-sm sm:px-6 lg:hidden">
      <button
        type="button"
        className="-m-2.5 p-2.5 text-gray-400 lg:hidden"
        onClick={() => setSidebarOpen(true)}
      >
        <span className="sr-only">Open sidebar</span>
        <Bars3Icon className="h-6 w-6" aria-hidden="true" />
      </button>
      <div className="flex-1 text-sm font-semibold leading-6 text-white">
        kite.onl
      </div>
      <a href="#">
        <span className="sr-only">Your profile</span>
        <img
          className="h-8 w-8 rounded-full bg-gray-800"
          src={userAvatarUrl(user)}
          alt=""
        />
      </a>
    </div>
  );
}
