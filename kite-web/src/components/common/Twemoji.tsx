import { ReactNode, useMemo } from "react";
import Inner from "react-twemoji";

interface Props {
  children: ReactNode;
  options: any;
}

export default function Twemoji({ children, options }: Props) {
  const opt = useMemo(
    () => ({
      ...options,
      base: "https://cdn.jsdelivr.net/gh/twitter/twemoji@14.0.2/assets/",
    }),
    [options]
  );

  return <Inner options={opt}>{children}</Inner>;
}
