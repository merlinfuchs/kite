import hljs from "highlight.js/lib/core";
import "highlight.js/styles/atom-one-dark.css";
import typescript from "highlight.js/lib/languages/typescript";
import go from "highlight.js/lib/languages/go";
import rust from "highlight.js/lib/languages/rust";
import python from "highlight.js/lib/languages/python";

hljs.registerLanguage("typescript", typescript);
hljs.registerLanguage("go", go);
hljs.registerLanguage("rust", rust);
hljs.registerLanguage("python", python);

export default function CodeHighlight({
  code,
  language,
}: {
  code: string;
  language: string;
}) {
  const highlightedCode = hljs.highlight(code, { language }).value;

  return (
    <pre className="text-gray-200">
      <code dangerouslySetInnerHTML={{ __html: highlightedCode }}></code>
    </pre>
  );
}
