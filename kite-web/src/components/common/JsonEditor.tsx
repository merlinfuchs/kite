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
}: {
  src: any;
  onChange: (src: any) => void;
}) {
  return (
    <ReactJson
      src={src}
      enableClipboard={true}
      onEdit={(e: ChangeEvent) => {
        onChange(e.updated_src);
        return true;
      }}
      onAdd={(e: ChangeEvent) => {
        onChange(e.updated_src);
        return true;
      }}
      onDelete={(e: ChangeEvent) => {
        onChange(e.updated_src);
        return true;
      }}
      theme="google"
      style={{
        padding: "10px",
        borderRadius: "5px",
      }}
    />
  );
}
