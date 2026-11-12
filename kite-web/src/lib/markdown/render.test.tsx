import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";
import FlowAIMarkdown from "@/components/flow/FlowAIMarkdown";

describe("FlowAIMarkdown", () => {
  it("renders with React 19", () => {
    const html = renderToStaticMarkup(
      <FlowAIMarkdown
        text={"Added a **Send Message** block:\n- reads `user.id`\n- replies"}
      />
    );
    expect(html).toContain("<strong>Send Message</strong>");
    expect(html).toContain("<li>reads <code>user.id</code></li>");
    expect(html).toContain("<li>replies</li>");
  });
});
