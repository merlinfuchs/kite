import { ReactNode } from "react";
import { useAutoAnimate } from "@formkit/auto-animate/react";

export default function AutoAnimate({
  children,
  className,
}: {
  children: ReactNode;
  className: string;
}) {
  const [parent] = useAutoAnimate();

  return (
    <div className={className} ref={parent}>
      {children}
    </div>
  );
}
