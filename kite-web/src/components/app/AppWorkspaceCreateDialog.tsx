import { useWorkspaceCreateMutation } from "@/lib/api/mutations";
import { FlatFile } from "@/lib/code/filetree";
import { useRouter } from "next/router";
import { useState } from "react";
import { toast } from "sonner";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTrigger,
} from "../ui/dialog";
import { Button } from "../ui/button";
import { DialogTitle } from "@radix-ui/react-dialog";
import { Input } from "../ui/input";
import { Textarea } from "../ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../ui/select";

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
  guildId: string;
}

export default function AppWorkspaceCreateDialog({ guildId }: Props) {
  const router = useRouter();
  const createMutation = useWorkspaceCreateMutation(guildId);

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
            router.push(`/app/guilds/${guildId}/workspaces/${res.data.id}`);
          } else {
            toast.error("Failed to create workspace");
          }
        },
      }
    );
  }

  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button>New Workspace</Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create Workspace</DialogTitle>
          <DialogDescription>
            Create a new workspace to start building your plugin.
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-5">
          <div>
            <div className="text-foreground font-medium mb-1">Language</div>
            <div className="mb-2 text-muted-foreground text-sm">
              Kite supports many different programming languages and also a
              no-code flow editor. Only some languages are supported in Web
              Workspaces.
            </div>
            <Select value={type} onValueChange={(v) => setType(v as any)}>
              <SelectTrigger>
                <SelectValue placeholder="Language" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="FLOW">Flow Editor</SelectItem>
                <SelectItem value="JS">TypeScript</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div>
            <div className="text-foreground font-medium mb-1">Name</div>
            <div className="mb-2 text-muted-foreground text-sm">
              Give your workspace a name that you can remember.
            </div>
            <Input
              placeholder="My Workspace"
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          </div>
          <div>
            <div className="text-foreground font-medium mb-1">Description</div>
            <div className="mb-2 text-muted-foreground text-sm">
              Describe what your workspace is for and what you are working on.
            </div>
            <Textarea
              className="min-h-32"
              placeholder="My really cool plugin that includes a ping command"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
            ></Textarea>
          </div>
        </div>
        <div className="mt-5 sm:mt-6 sm:grid sm:grid-flow-row-dense sm:grid-cols-2 sm:gap-3">
          <DialogClose asChild>
            <Button variant="outline">Cancel</Button>
          </DialogClose>
          <Button onClick={createWorkspace}>Create Workspace</Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
