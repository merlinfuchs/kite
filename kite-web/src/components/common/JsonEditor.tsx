import dynamic from "next/dynamic";

const ReactJson = dynamic(() => import("react-json-view"), {
  ssr: false,
}) as any;

type ChangeEvent = {
  updated_src: any;
};

export default function JsonEditor({
  src,
  onChange,
  readOnly,
}: {
  src: any;
  onChange: (src: any) => void;
  readOnly?: boolean;
}) {
  function change(e: ChangeEvent) {
    onChange(e.updated_src);
    return true;
  }

  return (
    <ReactJson
      src={src}
      enableClipboard={true}
      // false hides the buttons to edit, add and delete.
      onEdit={!readOnly && change}
      onAdd={!readOnly && change}
      onDelete={!readOnly && change}
      theme="google"
      style={{
        padding: "10px",
        borderRadius: "5px",
      }}
    />
  );
}
