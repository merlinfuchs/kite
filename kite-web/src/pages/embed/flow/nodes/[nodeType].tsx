import EmbeddablePage from "@/components/embed/EmbeddablePage";
import { useNodeValues } from "@/lib/flow/nodes";
import { useRouter } from "next/router";

export default function EmbedFlowNodePage() {
  const router = useRouter();
  const { nodeType } = router.query;

  const values = useNodeValues(nodeType as string);

  return (
    <EmbeddablePage className="p-3 bg-muted relative select-none inline-block w-[400px]">
      <div className="flex items-start space-x-3">
        <div
          className="rounded-md w-8 h-8 flex justify-center items-center flex-none"
          style={{ backgroundColor: values.color }}
        >
          <values.icon className="h-5 w-5 text-white" />
        </div>
        <div className="overflow-hidden flex-1 min-w-0">
          <div className="font-medium text-foreground leading-5 mb-1 truncate">
            {values.defaultTitle}
          </div>
          <div className="text-sm text-muted-foreground">
            {values.defaultDescription}
          </div>
        </div>
      </div>
    </EmbeddablePage>
  );
}
