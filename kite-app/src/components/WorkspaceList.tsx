import {
  useWorkspaceCreateMutation,
  useWorkspaceDeleteMutation,
} from "@/lib/api/mutations";
import { useWorkspacesQuery } from "@/lib/api/queries";
import Link from "next/link";
import AutoAnimate from "./AutoAnimate";
import { FlatFile } from "@/util/filetree";
import { useRouter } from "next/router";
import toast from "react-hot-toast";
import clsx from "clsx";
import IllustrationPlaceholder from "./IllustrationPlaceholder";

const defaultFiles: FlatFile[] = [
  {
    path: "index.ts",
    content: `
import { someText } from "./lib/util.ts";

Kite.handle = function(event) {
    if (event.type === "DISCORD_MESSAGE_CREATE" && event.data.content === "!ping") {
        Kite.call({
            type: "DISCORD_MESSAGE_CREATE",
            data: {
                channel_id: event.data.channel_id,
                content: someText(),
            },
        });
    }

    return { success: true };
}
      `.trim(),
  },
  {
    path: "lib/util.ts",
    content: `
export function someText() {
    return "Pong!";
}
      `.trim(),
  },
  {
    path: "manifest.toml",
    content: `
[plugin]
key = "example@kite.bot"
name = 'My Plugin'
description = 'Example Kite plugin'
type = 'js'
      
events = ["DISCORD_MESSAGE_CREATE"]
      `.trim(),
  },
];

export default function WorkspaceList({ guildId }: { guildId: string }) {
  const router = useRouter();

  const { data: resp, refetch } = useWorkspacesQuery(guildId);

  const workspaces = resp?.success ? resp.data : [];

  const deleteMutation = useWorkspaceDeleteMutation(guildId);

  function deleteWorkspace(workspaceId: string) {
    if (confirm("Are you sure you want to delete this workspace?")) {
      deleteMutation.mutate(
        { workspaceId },
        {
          onSuccess: (res) => {
            if (res.success) {
              refetch();
            } else {
              toast.error("Failed to delete workspace");
            }
          },
        }
      );
    }
  }

  const createMutation = useWorkspaceCreateMutation(guildId);

  function createWorkspace() {
    createMutation.mutate(
      {
        name: "New Workspace",
        description: "A new Workspace for my new cool Plugin!",
        files: defaultFiles.map((file) => ({
          path: file.path,
          content: file.content,
        })),
      },
      {
        onSuccess: (res) => {
          if (res.success) {
            refetch();
            router.push(`/guilds/${guildId}/workspaces/${res.data.id}`);
          } else {
            toast.error("Failed to create workspace");
          }
        },
      }
    );
  }

  return (
    <div>
      <div>
        <AutoAnimate
          className={clsx(
            "flex flex-col space-y-5",
            workspaces.length !== 0 && "mb-10"
          )}
        >
          {workspaces.map((w) => (
            <div className="bg-slate-800 px-5 py-4 rounded-md" key={w.id}>
              <div className="flex">
                <div className="flex-auto">
                  <div className="text-gray-100 text-lg font-medium mb-1">
                    {w.name}
                  </div>
                  <div className="font-light text-gray-300">
                    {w.description}
                  </div>
                </div>
                <div className="flex-none flex space-x-3 items-start">
                  <button
                    className="px-3 py-2 bg-slate-700 hover:bg-slate-600 text-gray-100 rounded select-none"
                    onClick={() => deleteWorkspace(w.id)}
                  >
                    Delete
                  </button>
                  <Link
                    className="px-3 py-2 bg-slate-700 hover:bg-slate-600 text-gray-100 rounded select-none"
                    href={`/guilds/${guildId}/workspaces/${w.id}`}
                  >
                    Open Editor
                  </Link>
                </div>
              </div>
            </div>
          ))}
        </AutoAnimate>
        <div>
          <button
            className="px-4 py-2 text-gray-100 rounded border-2 border-slate-400 hover:bg-slate-600 text-lg"
            onClick={createWorkspace}
          >
            New Workspace
          </button>
        </div>
      </div>
      {workspaces.length === 0 && (
        <IllustrationPlaceholder
          svgPath="/illustrations/software_engineer.svg"
          title="Create you first workspace and get coding without worrying about the boring stuff!"
          className="mt-10"
        />
      )}
    </div>
  );
}
