import FlowEditor from "./FlowEditor";

export default function Flow() {
  return (
    <div className="h-[100dvh] w-[100dvw] flex">
      <div className="flex-none w-96"></div>
      <div className="flex-auto h-full">
        <FlowEditor />
      </div>
    </div>
  );
}
