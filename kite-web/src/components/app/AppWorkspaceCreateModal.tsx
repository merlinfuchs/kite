import ModalBase from "@/components/ModalBase";
import { useWorkspaceCreateMutation } from "@/lib/api/mutations";
import { FlatFile } from "@/lib/code/filetree";
import { useRouter } from "next/router";
import { useState } from "react";
import toast from "react-hot-toast";

const defaultTsFiles: FlatFile[] = [
  {
    path: "index.ts",
    content: `
  import { someText } from "./lib/util.ts";
  import { call, event } from "@merlingg/kite-sdk";
  
  event.on("DISCORD_MESSAGE_CREATE", (msg) => {
      if (msg.content === "!ping") {
          call("DISCORD_MESSAGE_CREATE", {
              channel_id: msg.channel_id,
              content: someText(),
          })
      }
  })
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
  [deployment]
  key = "example@kite.onl"
  name = 'My Plugin'
  description = 'Example Kite plugin'
  
  [module]
  type = 'js'
        `.trim(),
  },
];

const defaultFlowFiles: FlatFile[] = [
  {
    path: "default.flow",
    content: JSON.stringify({
      nodes: [],
      edges: [],
    }),
  },
];

interface Props {
  open: boolean;
  setOpen: (open: boolean) => void;
  appId: string;
}

export default function AppWorkspaceCreateModal({
  open,
  setOpen,
  appId,
}: Props) {
  const router = useRouter();
  const createMutation = useWorkspaceCreateMutation(appId);

  const [type, setType] = useState<"FLOW" | "JS">("JS");
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");

  function createWorkspace() {
    if (!name || !description) {
      toast.error("Name and description are required");
      return;
    }

    const files = type === "FLOW" ? defaultFlowFiles : defaultTsFiles;

    createMutation.mutate(
      {
        type,
        name,
        description,
        files,
      },
      {
        onSuccess: (res) => {
          if (res.success) {
            router.push(`/apps/${appId}/workspaces/${res.data.id}`);
          } else {
            toast.error("Failed to create workspace");
          }
        },
      }
    );
  }

  return (
    <ModalBase open={open} setOpen={setOpen}>
      <div className="text-gray-100 font-medium text-xl mb-5">
        Create Workspace
      </div>
      <div className="space-y-5">
        <div>
          <div className="text-gray-100 font-medium mb-1">Language</div>
          <div className="mb-2 text-gray-300 text-sm">
            Kite supports many different programming languages and also a
            no-code flow editor. Only some languages are supported in Web
            Workspaces.
          </div>
          <select
            className="bg-dark-2 w-full px-3 py-2 rounded focus:outline-none text-gray-100"
            value={type}
            onChange={(e) => setType(e.target.value as any)}
          >
            <option value="FLOW">Flow Editor</option>
            <option value="JS">TypeScript</option>
          </select>
        </div>
        <div>
          <div className="text-gray-100 font-medium mb-1">Name</div>
          <div className="mb-2 text-gray-300 text-sm">
            Give your workspace a name that you can remember.
          </div>
          <input
            className="bg-dark-2 w-full px-3 py-2 rounded focus:outline-none text-gray-100 placeholder:font-light placeholder:text-gray-500"
            placeholder="My Workspace"
            value={name}
            onChange={(e) => setName(e.target.value)}
          />
        </div>
        <div>
          <div className="text-gray-100 font-medium mb-1">Description</div>
          <div className="mb-2 text-gray-300 text-sm">
            Describe what your workspace is for and what you are working on.
          </div>
          <textarea
            className="bg-dark-2 w-full px-3 py-2 rounded focus:outline-none text-gray-100 placeholder:font-light placeholder:text-gray-500 min-h-32"
            placeholder="My really cool plugin that includes a ping command"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
          ></textarea>
        </div>
      </div>
      <div className="mt-5 sm:mt-6 sm:grid sm:grid-flow-row-dense sm:grid-cols-2 sm:gap-3">
        <button
          type="button"
          className="inline-flex w-full justify-center rounded-md bg-primary px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-primary-dark focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 sm:col-start-2"
          onClick={createWorkspace}
        >
          Create Workspace
        </button>
        <button
          type="button"
          className="mt-3 inline-flex w-full justify-center rounded-md text-white bg-dark-8 px-3 py-2 text-sm font-semibold shadow-sm hover:bg-dark-7 sm:col-start-1 sm:mt-0"
          onClick={() => setOpen(false)}
        >
          Cancel
        </button>
      </div>
    </ModalBase>
  );
}
