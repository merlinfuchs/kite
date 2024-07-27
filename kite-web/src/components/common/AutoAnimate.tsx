import { useAutoAnimate } from "@formkit/auto-animate/react";

export default function AutoAnimate(props: any) {
  const [parent] = useAutoAnimate();

  return <div ref={parent} {...props}></div>;
}
