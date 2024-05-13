import { useState } from "react";
import ModalBase from "../ModalBase";
import { useAppCreateMutation } from "@/lib/api/mutations";
import toast from "react-hot-toast";

interface Props {
  open: boolean;
  setOpen: (open: boolean) => void;
}

export default function AppCreateAppModal({ open, setOpen }: Props) {
  const [token, setToken] = useState("");

  const createMutation = useAppCreateMutation();

  function createApp() {
    createMutation.mutate(
      {
        token,
      },
      {
        onSuccess: (res) => {
          if (res.success) {
            setOpen(false);
          } else {
            toast.error("Failed to add app");
          }
        },
      }
    );
  }

  return (
    <ModalBase open={open} setOpen={setOpen}>
      <div className="text-gray-100 font-medium text-xl mb-5">Add App</div>
      <div className="space-y-5">
        <div>
          <div className="text-gray-100 font-medium mb-1">Bot Token</div>
          <div className="mb-2 text-gray-300 text-sm">
            Get the bot token from the Discord Developer Portal.
          </div>
          <input
            className="bg-dark-2 w-full px-3 py-2 rounded focus:outline-none text-gray-100 placeholder:font-light placeholder:text-gray-500"
            type="password"
            value={token}
            placeholder="your-bot-token"
            onChange={(e) => setToken(e.target.value)}
          />
        </div>
      </div>
      <div className="mt-5 sm:mt-6 sm:grid sm:grid-flow-row-dense sm:grid-cols-2 sm:gap-3">
        <button
          type="button"
          className="inline-flex w-full justify-center rounded-md bg-primary px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-primary-dark focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 sm:col-start-2"
          onClick={createApp}
        >
          Add App
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
