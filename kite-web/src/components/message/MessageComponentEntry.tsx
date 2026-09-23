import { NodeId } from "@/lib/message/document";
import { useNode } from "@/lib/message/state";
import MessageComponentButton from "./MessageComponentButton";
import MessageComponentContainer from "./MessageComponentContainer";
import MessageComponentFile from "./MessageComponentFile";
import MessageComponentMediaGallery from "./MessageComponentMediaGallery";
import MessageComponentRow from "./MessageComponentRow";
import MessageComponentSection from "./MessageComponentSection";
import MessageComponentSeparator from "./MessageComponentSeparator";
import MessageComponentTextDisplay from "./MessageComponentTextDisplay";
import MessageComponentThumbnail from "./MessageComponentThumbnail";

/** Picks the editor for a component by its type, at any depth. */
export default function MessageComponentEntry({
  id,
  title,
  disableFlowEditor,
}: {
  id: NodeId;
  title?: string;
  disableFlowEditor?: boolean;
}) {
  const type = useNode(id)?.type;

  switch (type) {
    case "actionRow":
      return (
        <MessageComponentRow rowId={id} disableFlowEditor={disableFlowEditor} />
      );
    case "button":
      return (
        <MessageComponentButton
          buttonId={id}
          title={title}
          disableFlowEditor={disableFlowEditor}
        />
      );
    case "container":
      return (
        <MessageComponentContainer
          id={id}
          disableFlowEditor={disableFlowEditor}
        />
      );
    case "section":
      return (
        <MessageComponentSection
          id={id}
          disableFlowEditor={disableFlowEditor}
        />
      );
    case "textDisplay":
      return <MessageComponentTextDisplay id={id} />;
    case "thumbnail":
      return <MessageComponentThumbnail id={id} title={title} />;
    case "mediaGallery":
      return <MessageComponentMediaGallery id={id} />;
    case "file":
      return <MessageComponentFile id={id} />;
    case "separator":
      return <MessageComponentSeparator id={id} />;
    case undefined:
      return null;
    default:
      return (
        <div className="text-muted-foreground">
          Unsupported component: {type}
        </div>
      );
  }
}
