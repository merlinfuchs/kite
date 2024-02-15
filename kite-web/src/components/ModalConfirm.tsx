import { ExclamationTriangleIcon } from "@heroicons/react/24/solid";
import ModalBase from "./ModalBase";
import { Dialog } from "@headlessui/react";

interface Props {
  open: boolean;
  onConfirm: () => void;
  onCancel: () => void;
  title: string;
  description: string;
}

export default function ModalConfirm({
  open,
  onConfirm,
  onCancel,
  title,
  description,
}: Props) {
  return (
    <ModalBase open={open} setOpen={onCancel}>
      <div className="mb-10">
        <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-orange-200">
          <ExclamationTriangleIcon
            className="h-6 w-6 text-orange-600"
            aria-hidden="true"
          />
        </div>
        <div className="mt-3 text-center sm:mt-5">
          <Dialog.Title
            as="h3"
            className="text-base font-semibold leading-6 text-gray-100"
          >
            {title}
          </Dialog.Title>
          <div className="mt-2">
            <p className="text-sm text-gray-300">{description}</p>
          </div>
        </div>
      </div>
      <div className="mt-5 sm:mt-6 sm:grid sm:grid-flow-row-dense sm:grid-cols-2 sm:gap-3">
        <button
          type="button"
          className="inline-flex w-full justify-center rounded-md bg-primary px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-primary-dark focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 sm:col-start-2"
          onClick={onConfirm}
        >
          Confirm
        </button>
        <button
          type="button"
          className="mt-3 inline-flex w-full justify-center rounded-md text-white bg-dark-8 px-3 py-2 text-sm font-semibold shadow-sm hover:bg-dark-7 sm:col-start-1 sm:mt-0"
          onClick={onCancel}
        >
          Cancel
        </button>
      </div>
    </ModalBase>
  );
}
