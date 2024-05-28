import ModalBase from "@/components/ModalBase";
import {
  useDeploymentCreateMutation,
  useWorkspaceCreateMutation,
} from "@/lib/api/mutations";
import { useRouter } from "next/router";
import { useRef, useState } from "react";
import toast from "react-hot-toast";

interface Props {
  open: boolean;
  setOpen: (open: boolean) => void;
  appId: string;
}

export default function AppDeploymentCreateModal({
  open,
  setOpen,
  appId,
}: Props) {
  const router = useRouter();
  const createMutation = useDeploymentCreateMutation(appId);

  const [key, setKey] = useState("");
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [module, setModule] = useState<File | null>(null);

  function handleModuleChange(e: React.ChangeEvent<HTMLInputElement>) {
    if (!e.target.files) return;
    setModule(e.target.files[0]);
  }

  function createDeployment() {
    if (!key || !name || !description) {
      toast.error("Key, name, and description are required");
      return;
    }

    if (!module) {
      toast.error("Please upload a WASM module");
      return;
    }

    // encode module file as base64
    const reader = new FileReader();
    reader.onload = () => {
      const moduleData = (reader.result as string).split(",")[1];

      createMutation.mutate(
        {
          key,
          name,
          description,
          wasm_bytes: moduleData,
          plugin_version_id: null,
          config: {},
        },
        {
          onSuccess: (res) => {
            if (res.success) {
              router.push(`/apps/${appId}/deployments/${res.data.id}`);
            } else {
              toast.error("Failed to create deployment");
            }
          },
        }
      );
    };
    reader.readAsDataURL(module);
  }

  return (
    <ModalBase open={open} setOpen={setOpen}>
      <div className="text-gray-100 font-medium text-xl mb-5">
        Create Deployment
      </div>
      <div className="space-y-5">
        <div>
          <div className="text-gray-100 font-medium mb-1">Key</div>
          <div className="mb-2 text-gray-300 text-sm">
            The unique key that identifies this deployment. Re-using the same
            key will override the existing deployment.
          </div>
          <input
            className="bg-dark-2 w-full px-3 py-2 rounded focus:outline-none text-gray-100 placeholder:font-light placeholder:text-gray-500"
            placeholder="plugin@example.com"
            value={key}
            onChange={(e) => setKey(e.target.value)}
          />
        </div>
        <div>
          <div className="text-gray-100 font-medium mb-1">Name</div>
          <div className="mb-2 text-gray-300 text-sm">
            Give your deployment a name that you can remember.
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
            Describe what your deployment does and what it doesn't.
          </div>
          <textarea
            className="bg-dark-2 w-full px-3 py-2 rounded focus:outline-none text-gray-100 placeholder:font-light placeholder:text-gray-500 min-h-32"
            placeholder="My really cool plugin that includes a ping command"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
          ></textarea>
        </div>
        <div>
          <div className="text-gray-100 font-medium mb-1">WASM Module</div>
          <div className="mb-2 text-gray-300 text-sm">
            Upload the compiled WASM module for your deployment.
          </div>
          <input
            className="bg-dark-2 w-full px-3 py-2 rounded focus:outline-none text-gray-100 placeholder:font-light placeholder:text-gray-500"
            type="file"
            onChange={handleModuleChange}
          />
        </div>
      </div>
      <div className="mt-5 sm:mt-6 sm:grid sm:grid-flow-row-dense sm:grid-cols-2 sm:gap-3">
        <button
          type="button"
          className="inline-flex w-full justify-center rounded-md bg-primary px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-primary-dark focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 sm:col-start-2"
          onClick={createDeployment}
        >
          Create Deployment
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
