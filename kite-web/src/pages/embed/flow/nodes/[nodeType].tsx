import EmbeddablePage from "@/components/embed/EmbeddablePage";
import DynamicIcon from "@/components/icons/DynamicIcon";
import { getNodeValues, nodeTypes, NodeValues } from "@/lib/flow/nodes";
import { GetStaticPaths, GetStaticProps } from "next";

interface Props {
  nodeType: string;
  values: {
    title: string;
    description: string;
    color: string;
    icon: string;
  };
}

export const getStaticPaths: GetStaticPaths = async () => {
  return {
    paths: Object.keys(nodeTypes).map((nodeType) => ({
      params: { nodeType },
    })),
    fallback: false,
  };
};

export const getStaticProps: GetStaticProps<Props> = async (context) => {
  const values = getNodeValues(context.params?.nodeType as string);

  return {
    props: {
      nodeType: context.params?.nodeType as string,
      values: {
        title: values.defaultTitle,
        description: values.defaultDescription,
        color: values.color,
        icon: values.icon,
      },
    },
  };
};

export default function EmbedFlowNodePage({ values }: Props) {
  return (
    <EmbeddablePage className="p-3 bg-muted relative select-none inline-block w-[400px]">
      <div className="flex items-start space-x-3">
        <div
          className="rounded-md w-8 h-8 flex justify-center items-center flex-none"
          style={{ backgroundColor: values.color }}
        >
          <DynamicIcon
            name={values.icon as any}
            className="h-5 w-5 text-white"
          />
        </div>
        <div className="overflow-hidden flex-1 min-w-0">
          <div className="font-medium text-foreground leading-5 mb-1 truncate">
            {values.title}
          </div>
          <div className="text-sm text-muted-foreground">
            {values.description}
          </div>
        </div>
      </div>
    </EmbeddablePage>
  );
}
