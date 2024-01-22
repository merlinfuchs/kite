import Code from "@/components/code/Code";
import { FlatFile } from "@/util/filetree";
import { useState } from "react";

export default function NewDeploymentPage() {
  const [openFilePath, setOpenFilePath] = useState<string | null>("index.ts");

  const files: FlatFile[] = [
    {
      path: "index.ts",
      content: `Kite.makeCall({name: "cool!"});`,
    },
    {
      path: "manifest.toml",
      content: ``,
    },
    {
      path: "util/thing.js",
      content: `export function doThing() {}`,
    },
    {
      path: "util/yeet/thing.ts",
      content: `export function doThing() {}`,
    },
  ];

  return (
    <div>
      <Code
        files={files}
        openFilePath={openFilePath}
        setOpenFilePath={setOpenFilePath}
      />
    </div>
  );
}
